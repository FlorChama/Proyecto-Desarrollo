package dao

import (
	"ticketek-backend/domain"

	"gorm.io/gorm"
)

type EventDAO struct {
	db *gorm.DB
}

func NewEventDAO(db *gorm.DB) *EventDAO {
	return &EventDAO{db: db}
}

func (d *EventDAO) Create(event *domain.Event) error {
	return d.db.Create(event).Error
}

func (d *EventDAO) FindAllAdmin() ([]domain.Event, error) {
	var events []domain.Event
	err := d.db.Order("created_at DESC").Find(&events).Error
	return events, err
}

func (d *EventDAO) FindAll(filter domain.EventFilterRequest) ([]domain.Event, error) {
	var events []domain.Event
	query := d.db.Where("status = ?", domain.EventStatusActive)

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Search != "" {
		query = query.Where("title LIKE ? OR venue LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.Available {
		query = query.Where("available > 0")
	}

	err := query.Find(&events).Error
	return events, err
}

func (d *EventDAO) FindByID(id uint) (*domain.Event, error) {
	var event domain.Event
	err := d.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *EventDAO) Update(event *domain.Event) error {
	return d.db.Save(event).Error
}

func (d *EventDAO) Delete(id uint) error {
	return d.db.Model(&domain.Event{}).Where("id = ?", id).Update("status", domain.EventStatusCancelled).Error
}
