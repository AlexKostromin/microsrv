package payment

import "github.com/AlexKostromin/microsrv/payment/internal/service"

type Service struct{}

var _ service.PaymentService = (*Service)(nil)

func NewService() *Service {
	return &Service{}
}
