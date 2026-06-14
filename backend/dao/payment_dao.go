package dao

import (
	"ticketek-backend/domain"

	"gorm.io/gorm"
)

type PaymentDAO struct {
	db *gorm.DB
}

func NewPaymentDAO(db *gorm.DB) *PaymentDAO {
	return &PaymentDAO{db: db}
}

func (d *PaymentDAO) Create(payment *domain.Payment) error {
	return d.db.Create(payment).Error
}

func (d *PaymentDAO) FindByUserID(userID uint) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := d.db.Preload("Ticket").Preload("Ticket.Event").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&payments).Error
	return payments, err
}

func (d *PaymentDAO) FindByTicketID(ticketID uint) (*domain.Payment, error) {
	var payment domain.Payment
	err := d.db.Where("ticket_id = ?", ticketID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
