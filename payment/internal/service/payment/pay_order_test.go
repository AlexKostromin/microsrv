package payment

import (
	"context"
	"testing"

	"github.com/AlexKostromin/microsrv/payment/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Service_PayOrder_Success(t *testing.T) {
	svc := NewService()

	req := &model.Payment{
		OrderUuid:     "order-123",
		UserUuid:      "user-456",
		PaymentMethod: model.PaymentMethodCard,
	}

	resp, err := svc.PayOrder(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.TransactionUuid)

	_, parseErr := uuid.Parse(resp.TransactionUuid)
	assert.NoError(t, parseErr, "TransactionUuid должен быть валидным UUID")
}

func Test_Service_PayOrder_UniqueUUID(t *testing.T) {
	svc := NewService()

	req := &model.Payment{
		OrderUuid:     "order-123",
		UserUuid:      "user-456",
		PaymentMethod: model.PaymentMethodSbp,
	}

	resp1, err1 := svc.PayOrder(context.Background(), req)
	resp2, err2 := svc.PayOrder(context.Background(), req)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, resp1.TransactionUuid, resp2.TransactionUuid, "каждый вызов должен генерировать уникальный UUID")
}
