package domain

import "time"

// Auth
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Event
type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Date        time.Time `json:"date" binding:"required"`
	ExtraDates  string    `json:"extra_dates"`
	Duration    int       `json:"duration"`
	Venue       string    `json:"venue" binding:"required"`
	Capacity    int       `json:"capacity" binding:"required,min=1"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price" binding:"required,min=0"`
}

type UpdateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	ExtraDates  string    `json:"extra_dates"`
	Duration    int       `json:"duration"`
	Venue       string    `json:"venue"`
	Capacity    int       `json:"capacity"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
}

type EventFilterRequest struct {
	Category  string `form:"category"`
	Search    string `form:"search"`
	Available bool   `form:"available"`
}

// Ticket
type BuyTicketRequest struct {
	EventID       uint    `json:"event_id" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	TicketType    string  `json:"ticket_type"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
}

type TransferRequest struct {
	TargetEmail string `json:"target_email" binding:"required,email"`
}

// Report
type EventReportResponse struct {
	EventID       uint   `json:"event_id"`
	Title         string `json:"title"`
	TotalCapacity int    `json:"total_capacity"`
	TotalSold     int    `json:"total_sold"`
	TotalCancelled int   `json:"total_cancelled"`
	Available     int    `json:"available"`
	Buyers        []User `json:"buyers"`
}
