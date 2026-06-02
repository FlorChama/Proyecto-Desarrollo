package controllers

import (
	"net/http"
	"strconv"
	"ticketek-backend/domain"
	"ticketek-backend/services"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	eventService *services.EventService
}

func NewEventController(eventService *services.EventService) *EventController {
	return &EventController{eventService: eventService}
}

// GetAll - GET /api/events (público)
func (ctrl *EventController) GetAll(c *gin.Context) {
	var filter domain.EventFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	events, err := ctrl.eventService.GetAll(filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, events)
}

// GetByID - GET /api/events/:id (público)
func (ctrl *EventController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	event, err := ctrl.eventService.GetByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, event)
}

// Create - POST /api/admin/events (admin)
func (ctrl *EventController) Create(c *gin.Context) {
	var req domain.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	event, err := ctrl.eventService.Create(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.CreatedResponse(c, event)
}

// Update - PUT /api/admin/events/:id (admin)
func (ctrl *EventController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	var req domain.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	event, err := ctrl.eventService.Update(uint(id), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, event)
}

// Cancel - DELETE /api/admin/events/:id (admin)
func (ctrl *EventController) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	if err := ctrl.eventService.Cancel(uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "evento cancelado exitosamente"})
}
