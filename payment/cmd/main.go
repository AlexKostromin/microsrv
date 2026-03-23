package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	paymentV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
)

const grpcPort = 50052

type paymentService struct {
	paymentV1.UnimplementedPaymentServiceServer
}

func (s *paymentService) PayOrder(_ context.Context, req *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	transactionUUID := uuid.NewString()

	log.Printf("Оплата прошла успешно, transaction_uuid: %s (order: %s, user: %s, method: %s)\n",
		transactionUUID, req.GetOrderUuid(), req.GetUserUuid(), req.GetPaymentMethod().String())

	return &paymentV1.PayOrderResponse{
		TransactionUuid: transactionUUID,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	paymentV1.RegisterPaymentServiceServer(s, &paymentService{})

	reflection.Register(s)

	go func() {
		log.Printf("gRPC PaymentService listening on :%d\n", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down PaymentService...")
	s.GracefulStop()
	log.Println("PaymentService stopped")
}
