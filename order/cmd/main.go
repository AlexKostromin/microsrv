package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderV1 "github.com/AlexKostromin/microsrv/shared/pkg/openapi/order/v1"
	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
	paymentV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
)

const (
	httpPort          = "8080"
	inventoryGRPCAddr = "localhost:50051"
	paymentGRPCAddr   = "localhost:50052"
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// Order внутренняя структура заказа
type Order struct {
	UUID            string
	UserUUID        string
	PartUUIDs       []string
	TotalPrice      float64
	TransactionUUID *string
	PaymentMethod   *string
	Status          string
}

// OrderHandler реализует интерфейс orderV1.Handler
type OrderHandler struct {
	mu              sync.RWMutex
	orders          map[string]*Order
	inventoryClient inventoryV1.InventoryServiceClient
	paymentClient   paymentV1.PaymentServiceClient
}

func NewOrderHandler(invClient inventoryV1.InventoryServiceClient, payClient paymentV1.PaymentServiceClient) *OrderHandler {
	return &OrderHandler{
		orders:          make(map[string]*Order),
		inventoryClient: invClient,
		paymentClient:   payClient,
	}
}

// CreateOrder создаёт заказ
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderV1.CreateOrderRequest) (orderV1.CreateOrderRes, error) {
	// Получаем детали из InventoryService
	resp, err := h.inventoryClient.ListParts(ctx, &inventoryV1.ListPartsRequest{
		Filter: &inventoryV1.PartsFilter{
			Uuids: req.PartUuids,
		},
	})
	if err != nil {
		return &orderV1.InternalServerError{
			Code:    500,
			Message: "failed to get parts from payment: " + err.Error(),
		}, nil
	}

	// Проверяем, что все детали существуют
	foundParts := make(map[string]struct{}, len(resp.Parts))
	for _, p := range resp.Parts {
		foundParts[p.Uuid] = struct{}{}
	}
	for _, partUUID := range req.PartUuids {
		if _, ok := foundParts[partUUID]; !ok {
			return &orderV1.BadRequestError{
				Code:    400,
				Message: "part not found: " + partUUID,
			}, nil
		}
	}

	// Считаем total_price
	var totalPrice float64
	for _, p := range resp.Parts {
		totalPrice += p.Price
	}

	// Создаём заказ
	orderUUID := uuid.NewString()
	order := &Order{
		UUID:       orderUUID,
		UserUUID:   req.UserUUID,
		PartUUIDs:  req.PartUuids,
		TotalPrice: totalPrice,
		Status:     "PENDING_PAYMENT",
	}

	h.mu.Lock()
	h.orders[orderUUID] = order
	h.mu.Unlock()

	log.Printf("Создан заказ %s, total_price: %.2f\n", orderUUID, totalPrice)

	return &orderV1.CreateOrderResponse{
		UUID:       orderUUID,
		TotalPrice: totalPrice,
	}, nil
}

// GetOrder возвращает заказ по UUID
func (h *OrderHandler) GetOrder(_ context.Context, params orderV1.GetOrderParams) (orderV1.GetOrderRes, error) {
	h.mu.RLock()
	order, ok := h.orders[params.OrderUUID.String()]
	h.mu.RUnlock()

	if !ok {
		return &orderV1.NotFoundError{
			Code:    404,
			Message: "order not found: " + params.OrderUUID.String(),
		}, nil
	}

	dto := &orderV1.OrderDto{
		UUID:       order.UUID,
		UserUUID:   order.UserUUID,
		PartUuids:  order.PartUUIDs,
		TotalPrice: order.TotalPrice,
		Status:     orderV1.OrderStatus(order.Status),
	}

	if order.TransactionUUID != nil {
		dto.TransactionUUID = orderV1.NewOptString(*order.TransactionUUID)
	}
	if order.PaymentMethod != nil {
		dto.PaymentMethod = orderV1.NewOptPaymentMethod(orderV1.PaymentMethod(*order.PaymentMethod))
	}

	return dto, nil
}

// PayOrder оплачивает заказ
func (h *OrderHandler) PayOrder(ctx context.Context, req *orderV1.PayOrderRequest, params orderV1.PayOrderParams) (orderV1.PayOrderRes, error) {
	orderUUID := params.OrderUUID.String()

	h.mu.RLock()
	order, ok := h.orders[orderUUID]
	h.mu.RUnlock()

	if !ok {
		return &orderV1.NotFoundError{
			Code:    404,
			Message: "order not found: " + orderUUID,
		}, nil
	}

	// Маппинг OpenAPI PaymentMethod -> Proto PaymentMethod
	protoMethod := mapPaymentMethod(req.PaymentMethod)

	// Вызываем PaymentService
	payResp, err := h.paymentClient.PayOrder(ctx, &paymentV1.PayOrderRequest{
		OrderUuid:     orderUUID,
		UserUuid:      order.UserUUID,
		PaymentMethod: protoMethod,
	})
	if err != nil {
		return &orderV1.InternalServerError{
			Code:    500,
			Message: "payment failed: " + err.Error(),
		}, nil
	}

	// Обновляем заказ
	h.mu.Lock()
	order.Status = "PAID"
	txUUID := payResp.TransactionUuid
	order.TransactionUUID = &txUUID
	method := string(req.PaymentMethod)
	order.PaymentMethod = &method
	h.mu.Unlock()

	log.Printf("Заказ %s оплачен, transaction: %s\n", orderUUID, txUUID)

	return &orderV1.PayOrderResponse{
		TransactionUUID: txUUID,
	}, nil
}

// CancelOrder отменяет заказ
func (h *OrderHandler) CancelOrder(_ context.Context, params orderV1.CancelOrderParams) (orderV1.CancelOrderRes, error) {
	orderUUID := params.OrderUUID.String()

	h.mu.Lock()
	defer h.mu.Unlock()

	order, ok := h.orders[orderUUID]
	if !ok {
		return &orderV1.NotFoundError{
			Code:    404,
			Message: "order not found: " + orderUUID,
		}, nil
	}

	if order.Status == "PAID" {
		return &orderV1.ConflictError{
			Code:    409,
			Message: "order already paid and cannot be cancelled",
		}, nil
	}

	order.Status = "CANCELLED"
	log.Printf("Заказ %s отменён\n", orderUUID)

	return &orderV1.CancelOrderNoContent{}, nil
}

// NewError создаёт ошибку
func (h *OrderHandler) NewError(_ context.Context, err error) *orderV1.GenericErrorStatusCode {
	return &orderV1.GenericErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: orderV1.GenericError{
			Code:    orderV1.NewOptInt(http.StatusInternalServerError),
			Message: orderV1.NewOptString(err.Error()),
		},
	}
}

func mapPaymentMethod(method orderV1.PaymentMethod) paymentV1.PaymentMethod {
	switch method {
	case orderV1.PaymentMethodPAYMENTMETHODCARD:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_CARD
	case orderV1.PaymentMethodPAYMENTMETHODSBP:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_SBP
	case orderV1.PaymentMethodPAYMENTMETHODCREDITCARD:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case orderV1.PaymentMethodPAYMENTMETHODINVESTORMONEY:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY
	default:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}

func main() {
	// Подключаемся к InventoryService
	invConn, err := grpc.NewClient(inventoryGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to InventoryService: %v", err)
	}
	defer invConn.Close()
	invClient := inventoryV1.NewInventoryServiceClient(invConn)

	// Подключаемся к PaymentService
	payConn, err := grpc.NewClient(paymentGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to PaymentService: %v", err)
	}
	defer payConn.Close()
	payClient := paymentV1.NewPaymentServiceClient(payConn)

	// Создаём handler
	handler := NewOrderHandler(invClient, payClient)

	// Создаём OpenAPI сервер
	orderServer, err := orderV1.NewServer(handler)
	if err != nil {
		log.Fatalf("failed to create OpenAPI server: %v", err)
	}

	// Запускаем HTTP-сервер
	server := &http.Server{
		Addr:              net.JoinHostPort("", httpPort),
		Handler:           orderServer,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		log.Printf("HTTP OrderService listening on :%s\n", httpPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down OrderService...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown: %v", err)
	}
	log.Println("OrderService stopped")
}
