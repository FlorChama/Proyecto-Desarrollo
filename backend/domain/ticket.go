package domain

import "gorm.io/gorm"

const (
	TicketStatusActive      = "active"
	TicketStatusCancelled   = "cancelled"
	TicketStatusTransferred = "transferred"
)

type Ticket struct {
	gorm.Model
	UserID     uint    `gorm:"not null" json:"user_id"`
	EventID    uint    `gorm:"not null" json:"event_id"`
	Status     string  `gorm:"default:'active'" json:"status"`
	QRCode     string  `gorm:"not null" json:"qr_code"`
	TicketType string  `gorm:"default:'general'" json:"ticket_type"`
	Price      float64 `gorm:"default:0" json:"price"`
	User       User    `json:"user,omitempty"`
	Event      Event   `json:"event,omitempty"`
}
