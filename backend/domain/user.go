package domain

import "gorm.io/gorm"

const (
	RoleClient = "client"
	RoleAdmin  = "admin"
)

type User struct {
	gorm.Model
	Name     string   `gorm:"not null" json:"name"`
	Email    string   `gorm:"unique;not null" json:"email"`
	Password string   `gorm:"not null" json:"-"`
	Role     string   `gorm:"default:'client'" json:"role"`
	Tickets  []Ticket `json:"tickets,omitempty"`
}
