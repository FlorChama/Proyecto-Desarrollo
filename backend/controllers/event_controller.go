package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"ticketek-backend/domain"
	"ticketek-backend/services"
	"ticketek-backend/utils"
	"time"

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

// GetAllAdmin - GET /api/admin/events (admin)
func (ctrl *EventController) GetAllAdmin(c *gin.Context) {
	events, err := ctrl.eventService.GetAllAdmin()
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
		utils.HandleServiceError(c, err)
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
		utils.HandleServiceError(c, err)
		return
	}

	utils.SuccessResponse(c, event)
}

// UploadImage - POST /api/admin/events/upload (admin)
func (ctrl *EventController) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "no se recibió ninguna imagen")
		return
	}

	uploadDir := "./uploads/events"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "error al crear directorio")
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "error al guardar imagen")
		return
	}

	utils.SuccessResponse(c, gin.H{"url": "/uploads/events/" + filename})
}

// Cancel - DELETE /api/admin/events/:id (admin)
func (ctrl *EventController) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	if err := ctrl.eventService.Cancel(uint(id)); err != nil {
		utils.HandleServiceError(c, err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "evento cancelado exitosamente"})
}
