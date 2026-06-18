package services

import (
	"errors"
	"fmt"
	"ticketek-backend/clients"
	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/utils"
)

type TicketService struct {
	ticketDAO   *dao.TicketDAO
	eventDAO    *dao.EventDAO
	userDAO     *dao.UserDAO
	paymentDAO  *dao.PaymentDAO
	emailClient *clients.EmailClient
}

func NewTicketService(ticketDAO *dao.TicketDAO, eventDAO *dao.EventDAO, userDAO *dao.UserDAO, paymentDAO *dao.PaymentDAO, emailClient *clients.EmailClient) *TicketService {
	return &TicketService{
		ticketDAO:   ticketDAO,
		eventDAO:    eventDAO,
		userDAO:     userDAO,
		paymentDAO:  paymentDAO,
		emailClient: emailClient,
	}
}

func (s *TicketService) Buy(userID, eventID uint, paymentMethod string, amount float64) (*domain.Ticket, error) {
	event, err := s.eventDAO.FindByID(eventID)
	if err != nil {
		return nil, domain.ErrEventoNoEncontrado
	}
	if event.Status == domain.EventStatusCancelled {
		return nil, errors.New("el evento está cancelado")
	}
	if event.Available <= 0 {
		return nil, errors.New("no hay entradas disponibles")
	}

	user, err := s.userDAO.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUsuarioNoEncontrado
	}

	ticket := &domain.Ticket{
		UserID:  userID,
		EventID: eventID,
		Status:  domain.TicketStatusActive,
		QRCode:  "",
	}

	if err := s.ticketDAO.Create(ticket); err != nil {
		return nil, errors.New("error al procesar la compra")
	}

	qrData := utils.GenerateTicketQRData(ticket.ID, userID, eventID)
	qrCode, err := utils.GenerateQRCode(qrData)
	if err != nil {
		return nil, fmt.Errorf("error generando QR: %w", err)
	}

	ticket.QRCode = qrCode
	if err := s.ticketDAO.Update(ticket); err != nil {
		return nil, errors.New("error guardando QR")
	}

	event.Available--
	if err := s.eventDAO.Update(event); err != nil {
		return nil, errors.New("error actualizando disponibilidad")
	}

	payment := &domain.Payment{
		TicketID: ticket.ID,
		UserID:   userID,
		Amount:   amount,
		Method:   paymentMethod,
		Status:   domain.PaymentStatusApproved,
	}
	_ = s.paymentDAO.Create(payment)

	// Bonus: notificación por email
	go s.emailClient.SendPurchaseConfirmation(user.Email, user.Name, event.Title, qrCode)

	return ticket, nil
}

func (s *TicketService) GetMyTickets(userID uint) ([]domain.Ticket, error) {
	return s.ticketDAO.FindByUserID(userID)
}

func (s *TicketService) GetMyPayments(userID uint) ([]domain.Payment, error) {
	return s.paymentDAO.FindByUserID(userID)
}

func (s *TicketService) Cancel(ticketID, userID uint) error {
	ticket, err := s.ticketDAO.FindByID(ticketID)
	if err != nil {
		return domain.ErrEntradaNoEncontrada
	}
	if ticket.UserID != userID {
		return domain.ErrSinPermiso
	}
	if ticket.Status != domain.TicketStatusActive {
		return errors.New("la entrada no está activa")
	}

	ticket.Status = domain.TicketStatusCancelled
	if err := s.ticketDAO.Update(ticket); err != nil {
		return errors.New("error al cancelar la entrada")
	}

	event, err := s.eventDAO.FindByID(ticket.EventID)
	if err == nil {
		event.Available++
		s.eventDAO.Update(event)
	}

	return nil
}

func (s *TicketService) Transfer(ticketID, ownerID uint, req domain.TransferRequest) (*domain.Ticket, error) {
	ticket, err := s.ticketDAO.FindByID(ticketID)
	if err != nil {
		return nil, domain.ErrEntradaNoEncontrada
	}
	if ticket.UserID != ownerID {
		return nil, domain.ErrSinPermiso
	}
	if ticket.Status != domain.TicketStatusActive {
		return nil, errors.New("la entrada no está activa")
	}

	targetUser, err := s.userDAO.FindByEmail(req.TargetEmail)
	if err != nil {
		return nil, domain.ErrUsuarioNoEncontrado
	}
	if targetUser.ID == ownerID {
		return nil, errors.New("no podés traspasar la entrada a vos mismo")
	}

	owner, err := s.userDAO.FindByID(ownerID)
	if err != nil {
		return nil, errors.New("error obteniendo datos del propietario")
	}

	// Marcar el ticket original como traspasado (el dueño original lo sigue viendo)
	ticket.Status = domain.TicketStatusTransferred
	ticket.User = domain.User{}
	ticket.Event = domain.Event{}
	if err := s.ticketDAO.Update(ticket); err != nil {
		return nil, errors.New("error al traspasar la entrada")
	}

	// Generar nuevo QR para el nuevo titular
	qrData := utils.GenerateTicketQRData(ticket.ID, targetUser.ID, ticket.EventID)
	newQR, err := utils.GenerateQRCode(qrData)
	if err != nil {
		return nil, fmt.Errorf("error generando nuevo QR: %w", err)
	}

	// Crear ticket nuevo para el nuevo titular
	newTicket := &domain.Ticket{
		UserID:  targetUser.ID,
		EventID: ticket.EventID,
		Status:  domain.TicketStatusActive,
		QRCode:  newQR,
	}
	if err := s.ticketDAO.Create(newTicket); err != nil {
		return nil, errors.New("error al crear entrada para el nuevo titular")
	}

	event, _ := s.eventDAO.FindByID(ticket.EventID)
	eventTitle := ""
	if event != nil {
		eventTitle = event.Title
	}

	// Bonus: notificación por email al nuevo titular
	go s.emailClient.SendTransferNotification(targetUser.Email, targetUser.Name, owner.Name, eventTitle, newQR)

	return newTicket, nil
}

func (s *TicketService) GetEventReport(eventID uint) (*domain.EventReportResponse, error) {
	event, err := s.eventDAO.FindByID(eventID)
	if err != nil {
		return nil, domain.ErrEventoNoEncontrado
	}

	allTickets, err := s.ticketDAO.FindAllByEventID(eventID)
	if err != nil {
		return nil, errors.New("error obteniendo tickets")
	}

	sold := 0
	cancelled := 0
	buyers := []domain.User{}
	seen := map[uint]bool{}

	for _, t := range allTickets {
		if t.Status == domain.TicketStatusCancelled {
			cancelled++
		} else {
			sold++
		}
		if !seen[t.UserID] {
			buyers = append(buyers, t.User)
			seen[t.UserID] = true
		}
	}

	return &domain.EventReportResponse{
		EventID:        event.ID,
		Title:          event.Title,
		TotalCapacity:  event.Capacity,
		TotalSold:      sold,
		TotalCancelled: cancelled,
		Available:      event.Available,
		Buyers:         buyers,
	}, nil
}
