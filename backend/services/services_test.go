package services

import (
	"testing"
	"time"

	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// entorno agrupa los services ya cableados sobre una base SQLite de test.
type entorno struct {
	db     *gorm.DB
	auth   *AuthService
	event  *EventService
	ticket *TicketService
}

func nuevoEntorno(t *testing.T) *entorno {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("error abriendo SQLite: %v", err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Ticket{}, &domain.Payment{}); err != nil {
		t.Fatalf("error migrando: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

	userDAO := dao.NewUserDAO(db)
	eventDAO := dao.NewEventDAO(db)
	ticketDAO := dao.NewTicketDAO(db)
	paymentDAO := dao.NewPaymentDAO(db)

	return &entorno{
		db:     db,
		auth:   NewAuthService(userDAO),
		event:  NewEventService(eventDAO),
		ticket: NewTicketService(ticketDAO, eventDAO, userDAO, paymentDAO),
	}
}

// crearEventoActivo inserta un evento activo y devuelve su ID.
func (e *entorno) crearEventoActivo(t *testing.T, capacidad int) uint {
	t.Helper()
	ev, err := e.event.Create(domain.CreateEventRequest{
		Title: "Recital", Date: time.Now().Add(48 * time.Hour),
		Venue: "Estadio", Capacity: capacidad, Category: "nacional", Price: 1000,
	})
	if err != nil {
		t.Fatalf("crearEventoActivo: %v", err)
	}
	return ev.ID
}

// registrar crea un usuario y devuelve su ID.
func (e *entorno) registrar(t *testing.T, name, email string) uint {
	t.Helper()
	resp, err := e.auth.Register(domain.RegisterRequest{Name: name, Email: email, Password: "secret123"})
	if err != nil {
		t.Fatalf("registrar %s: %v", email, err)
	}
	return resp.User.ID
}

// ---- AuthService ----

func TestAuthService(t *testing.T) {
	e := nuevoEntorno(t)

	// Register OK -> rol client
	resp, err := e.auth.Register(domain.RegisterRequest{Name: "Ana", Email: "ana@test.com", Password: "secret123"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.User.Role != domain.RoleClient {
		t.Errorf("el registro público debe crear un cliente, dio %s", resp.User.Role)
	}
	if resp.Token == "" {
		t.Error("Register debe devolver token")
	}

	// Register con email duplicado -> error
	if _, err := e.auth.Register(domain.RegisterRequest{Name: "Ana2", Email: "ana@test.com", Password: "secret123"}); err == nil {
		t.Error("Register con email duplicado debe fallar")
	}

	// Login OK
	if _, err := e.auth.Login(domain.LoginRequest{Email: "ana@test.com", Password: "secret123"}); err != nil {
		t.Errorf("Login correcto no debe fallar: %v", err)
	}
	// Login con contraseña incorrecta
	if _, err := e.auth.Login(domain.LoginRequest{Email: "ana@test.com", Password: "mala"}); err == nil {
		t.Error("Login con contraseña incorrecta debe fallar")
	}
	// Login de usuario inexistente
	if _, err := e.auth.Login(domain.LoginRequest{Email: "nadie@test.com", Password: "secret123"}); err == nil {
		t.Error("Login de usuario inexistente debe fallar")
	}

	// ChangePassword OK
	if err := e.auth.ChangePassword(resp.User.ID, domain.ChangePasswordRequest{CurrentPassword: "secret123", NewPassword: "nueva123"}); err != nil {
		t.Errorf("ChangePassword correcto: %v", err)
	}
	// ChangePassword con actual incorrecta
	if err := e.auth.ChangePassword(resp.User.ID, domain.ChangePasswordRequest{CurrentPassword: "mala", NewPassword: "otra123"}); err == nil {
		t.Error("ChangePassword con actual incorrecta debe fallar")
	}
	// ChangePassword con nueva muy corta
	if err := e.auth.ChangePassword(resp.User.ID, domain.ChangePasswordRequest{CurrentPassword: "nueva123", NewPassword: "123"}); err == nil {
		t.Error("ChangePassword con nueva corta debe fallar")
	}
	// ChangePassword de usuario inexistente
	if err := e.auth.ChangePassword(99999, domain.ChangePasswordRequest{CurrentPassword: "x", NewPassword: "nueva123"}); err == nil {
		t.Error("ChangePassword de usuario inexistente debe fallar")
	}

	// ResetPassword OK
	if err := e.auth.ResetPassword(domain.ResetPasswordRequest{Email: "ana@test.com", NewPassword: "reset123"}); err != nil {
		t.Errorf("ResetPassword correcto: %v", err)
	}
	// ResetPassword de email inexistente
	if err := e.auth.ResetPassword(domain.ResetPasswordRequest{Email: "nadie@test.com", NewPassword: "reset123"}); err == nil {
		t.Error("ResetPassword de email inexistente debe fallar")
	}
	// El login con la contraseña reseteada anda
	if !utils.CheckPassword("reset123", mustUser(t, e, "ana@test.com").Password) {
		t.Error("la contraseña no quedó reseteada")
	}
}

func mustUser(t *testing.T, e *entorno, email string) domain.User {
	t.Helper()
	var u domain.User
	if err := e.db.Where("email = ?", email).First(&u).Error; err != nil {
		t.Fatalf("buscando usuario %s: %v", email, err)
	}
	return u
}

// ---- EventService ----

func TestEventService(t *testing.T) {
	e := nuevoEntorno(t)

	ev, err := e.event.Create(domain.CreateEventRequest{
		Title: "Recital", Date: time.Now().Add(24 * time.Hour),
		Venue: "Estadio", Capacity: 100, Category: "nacional", Price: 1000, VIPPrice: 2000,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// GetAll / GetAllAdmin / GetByID
	if todos, _ := e.event.GetAll(domain.EventFilterRequest{}); len(todos) != 1 {
		t.Errorf("GetAll: esperado 1, obtenido %d", len(todos))
	}
	if admin, _ := e.event.GetAllAdmin(); len(admin) != 1 {
		t.Errorf("GetAllAdmin: esperado 1, obtenido %d", len(admin))
	}
	if _, err := e.event.GetByID(ev.ID); err != nil {
		t.Errorf("GetByID: %v", err)
	}
	if _, err := e.event.GetByID(99999); err != domain.ErrEventoNoEncontrado {
		t.Errorf("GetByID inexistente: esperado ErrEventoNoEncontrado, obtenido %v", err)
	}

	// Update de varios campos
	actualizado, err := e.event.Update(ev.ID, domain.UpdateEventRequest{
		Title: "Editado", Description: "desc", Venue: "Nuevo", Capacity: 120,
		Category: "internacional", Price: 1500, VIPPrice: 3000, Duration: 90,
		Date: time.Now().Add(72 * time.Hour), ExtraDates: "2026-12-25", ImageURL: "x.png",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if actualizado.Title != "Editado" || actualizado.Capacity != 120 {
		t.Error("el Update no aplicó los cambios")
	}
	// Al subir capacidad, available sube en la diferencia
	if actualizado.Available != 120 {
		t.Errorf("available esperado 120, obtenido %d", actualizado.Available)
	}

	// Update de evento inexistente
	if _, err := e.event.Update(99999, domain.UpdateEventRequest{Title: "x"}); err != domain.ErrEventoNoEncontrado {
		t.Errorf("Update inexistente: esperado ErrEventoNoEncontrado, obtenido %v", err)
	}

	// Cancel OK y luego no se puede actualizar
	if err := e.event.Cancel(ev.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if _, err := e.event.Update(ev.ID, domain.UpdateEventRequest{Title: "x"}); err == nil {
		t.Error("Update de evento cancelado debe fallar")
	}
	// Cancel de inexistente
	if err := e.event.Cancel(99999); err != domain.ErrEventoNoEncontrado {
		t.Errorf("Cancel inexistente: esperado ErrEventoNoEncontrado, obtenido %v", err)
	}
}

// ---- TicketService: compra ----

func TestTicketBuy(t *testing.T) {
	e := nuevoEntorno(t)
	eventID := e.crearEventoActivo(t, 2)
	userID := e.registrar(t, "Cli", "cli@test.com")

	// Compra general OK con QR y pago
	ticket, err := e.ticket.Buy(userID, eventID, "credit_card", "general")
	if err != nil {
		t.Fatalf("Buy general: %v", err)
	}
	if ticket.QRCode == "" {
		t.Error("la compra debe generar QR")
	}
	if ticket.Price != 1000 {
		t.Errorf("precio esperado 1000, obtenido %v", ticket.Price)
	}
	if pagos, _ := e.ticket.GetMyPayments(userID); len(pagos) != 1 {
		t.Errorf("debe haber 1 pago, hay %d", len(pagos))
	}

	// Comprar evento inexistente
	if _, err := e.ticket.Buy(userID, 99999, "credit_card", "general"); err != domain.ErrEventoNoEncontrado {
		t.Errorf("Buy evento inexistente: esperado ErrEventoNoEncontrado, obtenido %v", err)
	}
	// Comprar con usuario inexistente
	if _, err := e.ticket.Buy(99999, eventID, "credit_card", "general"); err != domain.ErrUsuarioNoEncontrado {
		t.Errorf("Buy usuario inexistente: esperado ErrUsuarioNoEncontrado, obtenido %v", err)
	}
	// VIP sin precio VIP configurado
	if _, err := e.ticket.Buy(userID, eventID, "credit_card", "vip"); err == nil {
		t.Error("Buy VIP sin precio VIP debe fallar")
	}

	// Agotar stock (queda 1 disponible) y comprar sin stock
	if _, err := e.ticket.Buy(userID, eventID, "credit_card", "general"); err != nil {
		t.Fatalf("segunda compra: %v", err)
	}
	if _, err := e.ticket.Buy(userID, eventID, "credit_card", "general"); err == nil {
		t.Error("Buy sin stock debe fallar")
	}
}

func TestTicketBuyVIPYEventoCancelado(t *testing.T) {
	e := nuevoEntorno(t)
	userID := e.registrar(t, "Cli", "cli@test.com")

	// Evento con precio VIP
	ev, _ := e.event.Create(domain.CreateEventRequest{
		Title: "VIP Show", Date: time.Now().Add(24 * time.Hour),
		Venue: "Luna Park", Capacity: 5, Category: "nacional", Price: 1000, VIPPrice: 5000,
	})

	vip, err := e.ticket.Buy(userID, ev.ID, "credit_card", "vip")
	if err != nil {
		t.Fatalf("Buy VIP: %v", err)
	}
	if vip.Price != 5000 || vip.TicketType != "vip" {
		t.Errorf("entrada VIP mal armada: tipo=%s precio=%v", vip.TicketType, vip.Price)
	}

	// Comprar en evento cancelado
	e.event.Cancel(ev.ID)
	if _, err := e.ticket.Buy(userID, ev.ID, "credit_card", "general"); err == nil {
		t.Error("Buy en evento cancelado debe fallar")
	}
}

// ---- TicketService: cancelar ----

func TestTicketCancel(t *testing.T) {
	e := nuevoEntorno(t)
	eventID := e.crearEventoActivo(t, 5)
	dueño := e.registrar(t, "Dueño", "dueno@test.com")
	otro := e.registrar(t, "Otro", "otro@test.com")

	ticket, _ := e.ticket.Buy(dueño, eventID, "credit_card", "general")

	// Cancelar entrada ajena -> sin permiso
	if err := e.ticket.Cancel(ticket.ID, otro); err != domain.ErrSinPermiso {
		t.Errorf("cancelar ajena: esperado ErrSinPermiso, obtenido %v", err)
	}
	// Cancelar inexistente
	if err := e.ticket.Cancel(99999, dueño); err != domain.ErrEntradaNoEncontrada {
		t.Errorf("cancelar inexistente: esperado ErrEntradaNoEncontrada, obtenido %v", err)
	}
	// Cancelar OK
	if err := e.ticket.Cancel(ticket.ID, dueño); err != nil {
		t.Errorf("cancelar propia: %v", err)
	}
	// Cancelar de nuevo (ya no está activa)
	if err := e.ticket.Cancel(ticket.ID, dueño); err == nil {
		t.Error("cancelar dos veces debe fallar")
	}
}

// ---- TicketService: traspaso (el fix de esta entrega) ----

func TestTicketTransferOcultaYDevuelve(t *testing.T) {
	e := nuevoEntorno(t)
	eventID := e.crearEventoActivo(t, 5)
	a := e.registrar(t, "A", "a@test.com")
	b := e.registrar(t, "B", "b@test.com")

	original, _ := e.ticket.Buy(a, eventID, "credit_card", "general")

	// A le traspasa la entrada a B
	nuevo, err := e.ticket.Transfer(original.ID, a, domain.TransferRequest{TargetEmail: "b@test.com"})
	if err != nil {
		t.Fatalf("Transfer A->B: %v", err)
	}
	if nuevo.UserID != b {
		t.Errorf("la entrada nueva debe ser de B (%d), es de %d", b, nuevo.UserID)
	}
	if nuevo.QRCode == original.QRCode {
		t.Error("el traspaso debe generar un QR nuevo")
	}
	if nuevo.Price != original.Price || nuevo.TicketType != original.TicketType {
		t.Error("el traspaso debe conservar tipo y precio")
	}

	// La entrada ya NO aparece en "mis entradas" de A
	if mias, _ := e.ticket.GetMyTickets(a); len(mias) != 0 {
		t.Errorf("A no debe ver más la entrada traspasada, ve %d", len(mias))
	}
	// B sí la ve
	if deB, _ := e.ticket.GetMyTickets(b); len(deB) != 1 {
		t.Errorf("B debe ver 1 entrada, ve %d", len(deB))
	}

	// Confirmar que en la base quedó el registro traspasado de A
	var traspasadosA int64
	e.db.Model(&domain.Ticket{}).Where("user_id = ? AND status = ?", a, domain.TicketStatusTransferred).Count(&traspasadosA)
	if traspasadosA != 1 {
		t.Fatalf("debe quedar 1 registro traspasado de A, hay %d", traspasadosA)
	}

	// B se la devuelve a A: el registro traspasado viejo de A debe borrarse
	devuelto, err := e.ticket.Transfer(nuevo.ID, b, domain.TransferRequest{TargetEmail: "a@test.com"})
	if err != nil {
		t.Fatalf("Transfer B->A (devolución): %v", err)
	}
	if devuelto.UserID != a {
		t.Errorf("la entrada devuelta debe ser de A (%d), es de %d", a, devuelto.UserID)
	}

	// Ya no debe quedar ningún registro traspasado de A (se borró al volver)
	e.db.Model(&domain.Ticket{}).Where("user_id = ? AND status = ?", a, domain.TicketStatusTransferred).Count(&traspasadosA)
	if traspasadosA != 0 {
		t.Errorf("el registro traspasado de A debió borrarse al volver, hay %d", traspasadosA)
	}

	// A vuelve a ver exactamente 1 entrada activa
	if mias, _ := e.ticket.GetMyTickets(a); len(mias) != 1 {
		t.Errorf("A debe ver 1 entrada tras la devolución, ve %d", len(mias))
	}
}

func TestTicketTransferErrores(t *testing.T) {
	e := nuevoEntorno(t)
	eventID := e.crearEventoActivo(t, 5)
	a := e.registrar(t, "A", "a@test.com")
	e.registrar(t, "B", "b@test.com")

	ticket, _ := e.ticket.Buy(a, eventID, "credit_card", "general")

	// Traspasar inexistente
	if _, err := e.ticket.Transfer(99999, a, domain.TransferRequest{TargetEmail: "b@test.com"}); err != domain.ErrEntradaNoEncontrada {
		t.Errorf("transfer inexistente: esperado ErrEntradaNoEncontrada, obtenido %v", err)
	}
	// Traspasar entrada ajena
	if _, err := e.ticket.Transfer(ticket.ID, 99999, domain.TransferRequest{TargetEmail: "b@test.com"}); err != domain.ErrSinPermiso {
		t.Errorf("transfer ajena: esperado ErrSinPermiso, obtenido %v", err)
	}
	// Traspasar a email inexistente
	if _, err := e.ticket.Transfer(ticket.ID, a, domain.TransferRequest{TargetEmail: "nadie@test.com"}); err != domain.ErrUsuarioNoEncontrado {
		t.Errorf("transfer a inexistente: esperado ErrUsuarioNoEncontrado, obtenido %v", err)
	}
	// Traspasar a uno mismo
	if _, err := e.ticket.Transfer(ticket.ID, a, domain.TransferRequest{TargetEmail: "a@test.com"}); err == nil {
		t.Error("transfer a uno mismo debe fallar")
	}
	// Traspasar una entrada no activa (cancelada)
	e.ticket.Cancel(ticket.ID, a)
	if _, err := e.ticket.Transfer(ticket.ID, a, domain.TransferRequest{TargetEmail: "b@test.com"}); err == nil {
		t.Error("transfer de entrada no activa debe fallar")
	}
}

// ---- TicketService: reporte ----

func TestTicketGetEventReport(t *testing.T) {
	e := nuevoEntorno(t)
	eventID := e.crearEventoActivo(t, 10)
	c1 := e.registrar(t, "C1", "c1@test.com")
	c2 := e.registrar(t, "C2", "c2@test.com")

	t1, _ := e.ticket.Buy(c1, eventID, "credit_card", "general")
	e.ticket.Buy(c2, eventID, "credit_card", "general")
	e.ticket.Cancel(t1.ID, c1) // una cancelada

	rep, err := e.ticket.GetEventReport(eventID)
	if err != nil {
		t.Fatalf("GetEventReport: %v", err)
	}
	if rep.TotalSold != 1 {
		t.Errorf("vendidas: esperado 1, obtenido %d", rep.TotalSold)
	}
	if rep.TotalCancelled != 1 {
		t.Errorf("canceladas: esperado 1, obtenido %d", rep.TotalCancelled)
	}
	if len(rep.Buyers) != 2 {
		t.Errorf("compradores: esperado 2, obtenido %d", len(rep.Buyers))
	}

	// Reporte de evento inexistente
	if _, err := e.ticket.GetEventReport(99999); err != domain.ErrEventoNoEncontrado {
		t.Errorf("reporte inexistente: esperado ErrEventoNoEncontrado, obtenido %v", err)
	}
}
