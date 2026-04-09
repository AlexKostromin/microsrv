package v1

import (
	"context"

	"github.com/AlexKostromin/microsrv/payment/internal/converter"
	paymentV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
)

func (a *Api) PayOrder(ctx context.Context, req *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	payment, err := a.PaymentService.PayOrder(ctx, converter.ToPaymentFromProto(req))
	if err != nil {
		return nil, err
	}
	return converter.ToTransactionUuidFromService(payment), nil
}
