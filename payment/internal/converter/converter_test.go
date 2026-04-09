package converter

import (
	"testing"

	"github.com/AlexKostromin/microsrv/payment/internal/model"
	paymentV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/payment/v1"
	"github.com/stretchr/testify/assert"
)

func Test_ToPaymentFromProto_Success(t *testing.T) {
	req := &paymentV1.PayOrderRequest{
		OrderUuid:     "order-123",
		UserUuid:      "user-456",
		PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CARD,
	}

	result := ToPaymentFromProto(req)

	assert.NotNil(t, result)
	assert.Equal(t, "order-123", result.OrderUuid)
	assert.Equal(t, "user-456", result.UserUuid)
	assert.Equal(t, model.PaymentMethodCard, result.PaymentMethod)
}

func Test_ToPaymentFromProto_AllMethods(t *testing.T) {
	tests := []struct {
		name     string
		proto    paymentV1.PaymentMethod
		expected model.PaymentMethod
	}{
		{"Unspecified", paymentV1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED, model.PaymentMethodUnspecified},
		{"Card", paymentV1.PaymentMethod_PAYMENT_METHOD_CARD, model.PaymentMethodCard},
		{"SBP", paymentV1.PaymentMethod_PAYMENT_METHOD_SBP, model.PaymentMethodSbp},
		{"CreditCard", paymentV1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD, model.PaymentMethodCreditCard},
		{"InvestorMoney", paymentV1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY, model.PaymentMethodInvestorMoney},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &paymentV1.PayOrderRequest{
				PaymentMethod: tt.proto,
			}
			result := ToPaymentFromProto(req)
			assert.Equal(t, tt.expected, result.PaymentMethod)
		})
	}
}

func Test_ToTransactionUuidFromService_Success(t *testing.T) {
	resp := &model.PaymentResponse{
		TransactionUuid: "txn-789",
	}

	result := ToTransactionUuidFromService(resp)

	assert.NotNil(t, result)
	assert.Equal(t, "txn-789", result.TransactionUuid)
}
