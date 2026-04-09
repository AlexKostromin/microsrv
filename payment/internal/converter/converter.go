package converter

import (
	"github.com/AlexKostromin/microsrv/payment/internal/model"
	paymentV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
)

func ToTransactionUuidFromService(uuid *model.PaymentResponse) *paymentV1.PayOrderResponse {
	transactionUuid := uuid.TransactionUuid
	return &paymentV1.PayOrderResponse{
		TransactionUuid: transactionUuid,
	}
}
func ToPaymentFromProto(req *paymentV1.PayOrderRequest) *model.Payment {
	payment := &model.Payment{
		UserUuid:      req.UserUuid,
		OrderUuid:     req.OrderUuid,
		PaymentMethod: model.PaymentMethod(req.PaymentMethod),
	}
	return payment
}
