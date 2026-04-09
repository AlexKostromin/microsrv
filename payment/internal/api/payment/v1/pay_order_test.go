package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexKostromin/microsrv/payment/internal/model"
	"github.com/AlexKostromin/microsrv/payment/internal/service/mocks"
	paymentV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Api_PayOrder_Success(t *testing.T) {
	svc := mocks.NewPaymentService(t)
	api := NewApi(svc)

	req := &paymentV1.PayOrderRequest{
		OrderUuid:     "order-123",
		UserUuid:      "user-456",
		PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CARD,
	}

	svc.On("PayOrder", mock.Anything, &model.Payment{
		OrderUuid:     "order-123",
		UserUuid:      "user-456",
		PaymentMethod: model.PaymentMethodCard,
	}).Return(&model.PaymentResponse{
		TransactionUuid: "txn-789",
	}, nil)

	resp, err := api.PayOrder(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "txn-789", resp.TransactionUuid)
}

func Test_Api_PayOrder_ServiceError(t *testing.T) {
	svc := mocks.NewPaymentService(t)
	api := NewApi(svc)

	req := &paymentV1.PayOrderRequest{
		OrderUuid:     "order-123",
		UserUuid:      "user-456",
		PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_SBP,
	}

	svc.On("PayOrder", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))

	resp, err := api.PayOrder(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "service error", err.Error())
}
