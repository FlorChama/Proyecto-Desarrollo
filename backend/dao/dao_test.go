package dao

import (
	"testing"

	"ticketek-backend/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// nuevaDB abre una base SQLite en archivo temporal, aislada por test. Se cierra
// al terminar para que Windows pueda borrar el archivo.
func nuevaDB(t *testing.T) *gorm.DB {
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
	return db
}

func TestUserDAO(t *testing.T) {
	db := nuevaDB(t)
	d := NewUserDAO(db)

	u := &domain.User{Name: "Ana", Email: "ana@test.com", Password: "hash", Role: domain.RoleClient}
	if err := d.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	porEmail, err := d.FindByEmail("ana@test.com")
	if err != nil || porEmail.ID != u.ID {
		t.Fatalf("FindByEmail: %v", err)
	}

	porID, err := d.FindByID(u.ID)
	if err != nil || porID.Email != "ana@test.com" {
		t.Fatalf("FindByID: %v", err)
	}

	if err := d.UpdatePassword(u.ID, "nuevoHash"); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}
	actualizado, _ := d.FindByID(u.ID)
	if actualizado.Password != "nuevoHash" {
		t.Error("la contraseña no se actualizó")
	}

	// Casos de error
	if _, err := d.FindByEmail("noexiste@test.com"); err == nil {
		t.Error("FindByEmail de inexistente debe fallar")
	}
	if _, err := d.FindByID(99999); err == nil {
		t.Error("FindByID de inexistente debe fallar")
	}
}

func TestEventDAO(t *testing.T) {
	db := nuevaDB(t)
	d := NewEventDAO(db)

	e := &domain.Event{Title: "Recital", Venue: "Estadio", Capacity: 10, Available: 10,
		Category: "nacional", Price: 1000, Status: domain.EventStatusActive}
	if err := d.Create(e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	porID, err := d.FindByID(e.ID)
	if err != nil || porID.Title != "Recital" {
		t.Fatalf("FindByID: %v", err)
	}

	// FindAll con filtros
	todos, _ := d.FindAll(domain.EventFilterRequest{})
	if len(todos) != 1 {
		t.Errorf("FindAll: esperado 1, obtenido %d", len(todos))
	}
	porCategoria, _ := d.FindAll(domain.EventFilterRequest{Category: "nacional"})
	if len(porCategoria) != 1 {
		t.Errorf("FindAll categoría: esperado 1, obtenido %d", len(porCategoria))
	}
	porBusqueda, _ := d.FindAll(domain.EventFilterRequest{Search: "Recital", Available: true})
	if len(porBusqueda) != 1 {
		t.Errorf("FindAll búsqueda+available: esperado 1, obtenido %d", len(porBusqueda))
	}
	sinMatch, _ := d.FindAll(domain.EventFilterRequest{Category: "internacional"})
	if len(sinMatch) != 0 {
		t.Errorf("FindAll sin match: esperado 0, obtenido %d", len(sinMatch))
	}

	// FindAllAdmin trae todo
	admin, _ := d.FindAllAdmin()
	if len(admin) != 1 {
		t.Errorf("FindAllAdmin: esperado 1, obtenido %d", len(admin))
	}

	// Update
	e.Title = "Recital Editado"
	if err := d.Update(e); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Delete pasa el estado a cancelado y lo saca del listado público
	if err := d.Delete(e.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	publicos, _ := d.FindAll(domain.EventFilterRequest{})
	if len(publicos) != 0 {
		t.Errorf("tras Delete no debe haber eventos públicos, hay %d", len(publicos))
	}

	if _, err := d.FindByID(99999); err == nil {
		t.Error("FindByID de inexistente debe fallar")
	}
}

func TestTicketDAO(t *testing.T) {
	db := nuevaDB(t)
	d := NewTicketDAO(db)

	// Datos base
	db.Create(&domain.User{Name: "Cli", Email: "cli@test.com", Password: "h"})
	db.Create(&domain.Event{Title: "Ev", Venue: "V", Capacity: 5, Available: 5, Price: 100, Status: domain.EventStatusActive})

	activo := &domain.Ticket{UserID: 1, EventID: 1, Status: domain.TicketStatusActive, QRCode: "qr"}
	if err := d.Create(activo); err != nil {
		t.Fatalf("Create: %v", err)
	}
	traspasado := &domain.Ticket{UserID: 1, EventID: 1, Status: domain.TicketStatusTransferred, QRCode: "qr2"}
	d.Create(traspasado)

	// FindByID con preload
	porID, err := d.FindByID(activo.ID)
	if err != nil || porID.QRCode != "qr" {
		t.Fatalf("FindByID: %v", err)
	}

	// FindByUserID excluye los traspasados (mis entradas)
	mias, _ := d.FindByUserID(1)
	if len(mias) != 1 {
		t.Errorf("FindByUserID debe excluir traspasados: esperado 1, obtenido %d", len(mias))
	}

	// FindByEventID solo trae activos
	activos, _ := d.FindByEventID(1)
	if len(activos) != 1 {
		t.Errorf("FindByEventID: esperado 1 activo, obtenido %d", len(activos))
	}

	// FindAllByEventID trae todos los estados
	todos, _ := d.FindAllByEventID(1)
	if len(todos) != 2 {
		t.Errorf("FindAllByEventID: esperado 2, obtenido %d", len(todos))
	}

	// Update
	activo.Status = domain.TicketStatusCancelled
	if err := d.Update(activo); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// DeleteTransferredByUserAndEvent borra el registro traspasado del usuario
	if err := d.DeleteTransferredByUserAndEvent(1, 1); err != nil {
		t.Fatalf("DeleteTransferredByUserAndEvent: %v", err)
	}
	if restantes, _ := d.FindAllByEventID(1); len(restantes) != 1 {
		t.Errorf("tras borrar el traspasado debe quedar 1, hay %d", len(restantes))
	}

	if _, err := d.FindByID(99999); err == nil {
		t.Error("FindByID de inexistente debe fallar")
	}
}

func TestPaymentDAO(t *testing.T) {
	db := nuevaDB(t)
	d := NewPaymentDAO(db)

	db.Create(&domain.User{Name: "Cli", Email: "cli@test.com", Password: "h"})
	db.Create(&domain.Event{Title: "Ev", Venue: "V", Capacity: 5, Available: 5, Price: 100, Status: domain.EventStatusActive})
	db.Create(&domain.Ticket{UserID: 1, EventID: 1, Status: domain.TicketStatusActive, QRCode: "qr"})

	p := &domain.Payment{TicketID: 1, UserID: 1, Amount: 100, Method: "credit_card", Status: domain.PaymentStatusApproved}
	if err := d.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	porUser, _ := d.FindByUserID(1)
	if len(porUser) != 1 {
		t.Errorf("FindByUserID: esperado 1, obtenido %d", len(porUser))
	}

	porTicket, err := d.FindByTicketID(1)
	if err != nil || porTicket.Amount != 100 {
		t.Fatalf("FindByTicketID: %v", err)
	}

	if _, err := d.FindByTicketID(99999); err == nil {
		t.Error("FindByTicketID de inexistente debe fallar")
	}
}
