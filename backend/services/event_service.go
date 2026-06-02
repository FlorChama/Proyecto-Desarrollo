package services

import (
	"errors"
	"ticketek-backend/dao"
	"ticketek-backend/domain"
)

type EventService struct {
	eventDAO *dao.EventDAO
}

func NewEventService(eventDAO *dao.EventDAO) *EventService {
	return &EventService{eventDAO: eventDAO}
}

func (s *EventService) GetAll(filter domain.EventFilterRequest) ([]domain.Event, error) {
	return s.eventDAO.FindAll(filter)
}

func (s *EventService) GetByID(id uint) (*domain.Event, error) {
	event, err := s.eventDAO.FindByID(id)
	if err != nil {
		return nil, errors.New("evento no encontrado")
	}
	return event, nil
}

func (s *EventService) Create(req domain.CreateEventRequest) (*domain.Event, error) {
	event := &domain.Event{
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Duration:    req.Duration,
		Venue:       req.Venue,
		Capacity:    req.Capacity,
		Available:   req.Capacity,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Status:      domain.EventStatusActive,
	}

	if err := s.eventDAO.Create(event); err != nil {
		return nil, errors.New("error al crear el evento")
	}
	return event, nil
}

func (s *EventService) Update(id uint, req domain.UpdateEventRequest) (*domain.Event, error) {
	event, err := s.eventDAO.FindByID(id)
	if err != nil {
		return nil, errors.New("evento no encontrado")
	}
	if event.Status == domain.EventStatusCancelled {
		return nil, errors.New("no se puede modificar un evento cancelado")
	}

	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if !req.Date.IsZero() {
		event.Date = req.Date
	}
	if req.Duration > 0 {
		event.Duration = req.Duration
	}
	if req.Venue != "" {
		event.Venue = req.Venue
	}
	if req.Capacity > 0 {
		diff := req.Capacity - event.Capacity
		event.Capacity = req.Capacity
		event.Available += diff
		if event.Available < 0 {
			event.Available = 0
		}
	}
	if req.Category != "" {
		event.Category = req.Category
	}
	if req.ImageURL != "" {
		event.ImageURL = req.ImageURL
	}
	if req.Price >= 0 {
		event.Price = req.Price
	}

	if err := s.eventDAO.Update(event); err != nil {
		return nil, errors.New("error al actualizar el evento")
	}
	return event, nil
}

func (s *EventService) Cancel(id uint) error {
	_, err := s.eventDAO.FindByID(id)
	if err != nil {
		return errors.New("evento no encontrado")
	}
	return s.eventDAO.Delete(id)
}
