package dao

import (
	"ticketek-backend/domain"

	"gorm.io/gorm"
)

type TicketDAO struct {
	db *gorm.DB
}

func NewTicketDAO(db *gorm.DB) *TicketDAO {
	return &TicketDAO{db: db}
}

func (d *TicketDAO) Create(ticket *domain.Ticket) error {
	return d.db.Create(ticket).Error
}

func (d *TicketDAO) FindByID(id uint) (*domain.Ticket, error) {
	var ticket domain.Ticket
	err := d.db.Preload("User").Preload("Event").First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (d *TicketDAO) FindByUserID(userID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := d.db.Preload("Event").Where("user_id = ?", userID).Find(&tickets).Error
	return tickets, err
}

func (d *TicketDAO) FindByEventID(eventID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := d.db.Preload("User").Where("event_id = ? AND status = ?", eventID, domain.TicketStatusActive).Find(&tickets).Error
	return tickets, err
}

// FindAllByEventID trae todas las entradas de un evento sin importar su estado
// (activas y canceladas). Se usa para los reportes.
func (d *TicketDAO) FindAllByEventID(eventID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := d.db.Preload("User").Where("event_id = ?", eventID).Find(&tickets).Error
	return tickets, err
}

func (d *TicketDAO) Update(ticket *domain.Ticket) error {
	return d.db.Save(ticket).Error
}
