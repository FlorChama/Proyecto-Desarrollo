package tests

import (
	"fmt"
	"net/http"
	"testing"
)

// TestFlujosAdicionales cubre caminos felices que faltaban en los otros tests:
// mis entradas, mis pagos, cambio y reseteo de contraseña, listado con filtros
// y edición de un evento por el admin.
func TestFlujosAdicionales(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	adminToken, _ := createAdmin(t, db)
	userToken, _ := registerUser(t, r, "Leo", "leo@test.com", "secret123")
	eventID := createEvent(t, r, adminToken, 10)

	// Comprar una entrada para tener datos en "mis entradas" y "mis pagos"
	w := doRequest(r, "POST", "/api/tickets", userToken, buyBody(eventID))
	if w.Code != http.StatusCreated {
		t.Fatalf("compra: esperado 201, obtenido %d (%s)", w.Code, w.Body.String())
	}

	// Mis entradas -> 200
	w = doRequest(r, "GET", "/api/tickets/my", userToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("mis entradas: esperado 200, obtenido %d", w.Code)
	}

	// Mis pagos -> 200
	w = doRequest(r, "GET", "/api/tickets/payments", userToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("mis pagos: esperado 200, obtenido %d", w.Code)
	}

	// Cambio de contraseña con la actual correcta -> 200
	w = doRequest(r, "PUT", "/api/auth/change-password", userToken, map[string]string{
		"current_password": "secret123", "new_password": "nuevaclave1",
	})
	if w.Code != http.StatusOK {
		t.Errorf("cambio de contraseña: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}

	// Cambio con la actual incorrecta -> 400
	w = doRequest(r, "PUT", "/api/auth/change-password", userToken, map[string]string{
		"current_password": "incorrecta", "new_password": "otraclave1",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("cambio con clave incorrecta: esperado 400, obtenido %d", w.Code)
	}

	// Reseteo de contraseña -> 200
	w = doRequest(r, "POST", "/api/auth/reset-password", "", map[string]string{
		"email": "leo@test.com", "new_password": "reseteada1",
	})
	if w.Code != http.StatusOK {
		t.Errorf("reseteo: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}

	// Listado de eventos con filtros -> 200
	w = doRequest(r, "GET", "/api/events?category=nacional&search=Prueba&available=true", "", nil)
	if w.Code != http.StatusOK {
		t.Errorf("eventos con filtros: esperado 200, obtenido %d", w.Code)
	}

	// Edición de un evento por el admin -> 200
	w = doRequest(r, "PUT", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken, map[string]interface{}{
		"title": "Recital Editado", "capacity": 20,
	})
	if w.Code != http.StatusOK {
		t.Errorf("editar evento: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}
}

// TestRamasDeError cubre las ramas de error de los services que no estaban
// testeadas: login inválido, compra sin stock, VIP sin precio, evento cancelado.
func TestRamasDeError(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)
	adminToken, _ := createAdmin(t, db)
	cToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")

	// Login con contraseña incorrecta -> 401
	if w := doRequest(r, "POST", "/api/auth/login", "", map[string]string{
		"email": "cli@test.com", "password": "malaclave"}); w.Code != http.StatusUnauthorized {
		t.Errorf("login mal: esperado 401, obtenido %d", w.Code)
	}
	// Login de usuario inexistente -> 401
	if w := doRequest(r, "POST", "/api/auth/login", "", map[string]string{
		"email": "nadie@test.com", "password": "secret123"}); w.Code != http.StatusUnauthorized {
		t.Errorf("login inexistente: esperado 401, obtenido %d", w.Code)
	}

	// Evento con un solo lugar
	eventID := createEvent(t, r, adminToken, 1)

	// Comprar VIP en un evento sin precio VIP -> 400
	vip := map[string]interface{}{"event_id": eventID, "payment_method": "credit_card", "ticket_type": "vip"}
	if w := doRequest(r, "POST", "/api/tickets", cToken, vip); w.Code != http.StatusBadRequest {
		t.Errorf("compra VIP sin precio: esperado 400, obtenido %d", w.Code)
	}

	// Primera compra OK (agota el stock)
	if w := doRequest(r, "POST", "/api/tickets", cToken, buyBody(eventID)); w.Code != http.StatusCreated {
		t.Fatalf("compra: esperado 201, obtenido %d", w.Code)
	}
	// Segunda compra sin stock -> 400
	if w := doRequest(r, "POST", "/api/tickets", cToken, buyBody(eventID)); w.Code != http.StatusBadRequest {
		t.Errorf("compra sin stock: esperado 400, obtenido %d", w.Code)
	}

	// Transferir un ticket inexistente -> 404
	if w := doRequest(r, "POST", "/api/tickets/99999/transfer", cToken,
		map[string]string{"target_email": "cli@test.com"}); w.Code != http.StatusNotFound {
		t.Errorf("transferir inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Cancelar el evento
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken, nil); w.Code != http.StatusOK {
		t.Fatalf("cancelar evento: esperado 200, obtenido %d", w.Code)
	}
	// Comprar en un evento cancelado -> 400
	if w := doRequest(r, "POST", "/api/tickets", cToken, buyBody(eventID)); w.Code != http.StatusBadRequest {
		t.Errorf("compra en evento cancelado: esperado 400, obtenido %d", w.Code)
	}
	// Editar un evento cancelado -> 400
	if w := doRequest(r, "PUT", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken,
		map[string]string{"title": "Nuevo"}); w.Code != http.StatusBadRequest {
		t.Errorf("editar evento cancelado: esperado 400, obtenido %d", w.Code)
	}
}
