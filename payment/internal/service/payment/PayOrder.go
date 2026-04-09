package payment

import (
	"context"

	"github.com/AlexKostromin/microsrv/payment/internal/model"
	"github.com/google/uuid"
)

func (s *Service) PayOrder(ctx context.Context, req *model.Payment) (*model.PaymentResponse, error) {
	UUID := uuid.New()
	return &model.PaymentResponse{
		TransactionUuid: UUID.String(),
	}, nil
}
