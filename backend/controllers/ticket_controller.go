package controllers

import (
	"net/http"
	"strconv"
	"ticketek-backend/domain"
	"ticketek-backend/services"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
)

type TicketController struct {
	ticketService *services.TicketService
}

func NewTicketController(ticketService *services.TicketService) *TicketController {
	return &TicketController{ticketService: ticketService}
}

// Buy - POST /api/tickets (cliente autenticado)
func (ctrl *TicketController) Buy(c *gin.Context) {
	userID := c.GetUint("user_id")

	var body struct {
		EventID uint `json:"event_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	ticket, err := ctrl.ticketService.Buy(userID, body.EventID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.CreatedResponse(c, ticket)
}

// GetMyTickets - GET /api/tickets/my (cliente autenticado)
func (ctrl *TicketController) GetMyTickets(c *gin.Context) {
	userID := c.GetUint("user_id")

	tickets, err := ctrl.ticketService.GetMyTickets(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, tickets)
}

// Cancel - DELETE /api/tickets/:id (cliente autenticado)
func (ctrl *TicketController) Cancel(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	if err := ctrl.ticketService.Cancel(uint(id), userID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "entrada cancelada exitosamente"})
}

// Transfer - POST /api/tickets/:id/transfer (cliente autenticado)
func (ctrl *TicketController) Transfer(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	var req domain.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	ticket, err := ctrl.ticketService.Transfer(uint(id), userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, ticket)
}

// GetEventReport - GET /api/admin/events/:id/report (admin)
func (ctrl *TicketController) GetEventReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "id inválido")
		return
	}

	report, err := ctrl.ticketService.GetEventReport(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, report)
}
