package domain

import "gorm.io/gorm"

const (
	PaymentStatusPending  = "pending"
	PaymentStatusApproved = "approved"
	PaymentStatusFailed   = "failed"

	PaymentMethodCreditCard = "credit_card"
	PaymentMethodDebitCard  = "debit_card"
	PaymentMethodModo       = "modo"
	PaymentMethodMercadoPago = "mercadopago"
)

type Payment struct {
	gorm.Model
	TicketID uint    `gorm:"not null" json:"ticket_id"`
	UserID   uint    `gorm:"not null" json:"user_id"`
	Amount   float64 `gorm:"not null" json:"amount"`
	Method   string  `gorm:"not null" json:"method"`
	Status   string  `gorm:"default:'approved'" json:"status"`
	Ticket   Ticket  `json:"ticket,omitempty"`
	User     User    `json:"user,omitempty"`
}
