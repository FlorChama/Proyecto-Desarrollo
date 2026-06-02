package domain

import "gorm.io/gorm"

const (
	TicketStatusActive      = "active"
	TicketStatusCancelled   = "cancelled"
	TicketStatusTransferred = "transferred"
)

type Ticket struct {
	gorm.Model
	UserID  uint   `gorm:"not null" json:"user_id"`
	EventID uint   `gorm:"not null" json:"event_id"`
	Status  string `gorm:"default:'active'" json:"status"`
	QRCode  string `gorm:"not null" json:"qr_code"`
	User    User   `json:"user,omitempty"`
	Event   Event  `json:"event,omitempty"`
}
