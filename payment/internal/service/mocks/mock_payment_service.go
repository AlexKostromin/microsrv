package mocks

import (
	"context"

	"github.com/AlexKostromin/microsrv/payment/internal/model"
	"github.com/stretchr/testify/mock"
)

type PaymentService struct {
	mock.Mock
}

func (m *PaymentService) PayOrder(ctx context.Context, req *model.Payment) (*model.PaymentResponse, error) {
	ret := m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for PayOrder")
	}

	var r0 *model.PaymentResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Payment) (*model.PaymentResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.Payment) *model.PaymentResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PaymentResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.Payment) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewPaymentService(t interface {
	mock.TestingT
	Cleanup(func())
}) *PaymentService {
	m := &PaymentService{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
