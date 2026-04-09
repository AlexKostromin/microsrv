package v1

import (
	"github.com/AlexKostromin/microsrv/payment/internal/service"
	payment_v1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
)

type Api struct {
	service.PaymentService
	payment_v1.UnimplementedPaymentServiceServer
}

func NewApi(s service.PaymentService) *Api {
	return &Api{
		PaymentService: s,
	}
}
