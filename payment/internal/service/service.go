package service

import (
	"context"

	"github.com/AlexKostromin/microsrv/payment/internal/model"
)

type PaymentService interface {
	PayOrder(context.Context, *model.Payment) (*model.PaymentResponse, error)
}
