package model

type Payment struct {
	OrderUuid     string
	UserUuid      string
	PaymentMethod PaymentMethod
}

type PaymentResponse struct {
	TransactionUuid string
}
type PaymentMethod int32

const (
	PaymentMethodUnspecified   PaymentMethod = 0
	PaymentMethodCard          PaymentMethod = 1
	PaymentMethodSbp           PaymentMethod = 2
	PaymentMethodCreditCard    PaymentMethod = 3
	PaymentMethodInvestorMoney PaymentMethod = 4
)
