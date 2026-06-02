package domain

import (
	"time"

	"gorm.io/gorm"
)

const (
	EventStatusActive    = "active"
	EventStatusCancelled = "cancelled"
)

type Event struct {
	gorm.Model
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Date        time.Time `gorm:"not null" json:"date"`
	Duration    int       `json:"duration"` // minutos
	Venue       string    `gorm:"not null" json:"venue"`
	Capacity    int       `gorm:"not null" json:"capacity"`
	Available   int       `gorm:"not null" json:"available"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `gorm:"not null" json:"price"`
	Status      string    `gorm:"default:'active'" json:"status"`
	Tickets     []Ticket  `json:"tickets,omitempty"`
}
