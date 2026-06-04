package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestCasosBordeIDsInvalidos verifica que los endpoints respondan 400 ante IDs
// no numéricos en la URL.
func TestCasosBordeIDsInvalidos(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)
	adminToken, _ := createAdmin(t, db)
	cToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")

	casos := []struct {
		method, path, token string
		body                interface{}
	}{
		{"GET", "/api/events/abc", "", nil},
		{"DELETE", "/api/tickets/abc", cToken, nil},
		{"POST", "/api/tickets/abc/transfer", cToken, map[string]string{"target_email": "x@test.com"}},
		{"PUT", "/api/admin/events/abc", adminToken, map[string]string{"title": "x"}},
		{"DELETE", "/api/admin/events/abc", adminToken, nil},
		{"GET", "/api/admin/events/abc/report", adminToken, nil},
	}
	for _, c := range casos {
		if w := doRequest(r, c.method, c.path, c.token, c.body); w.Code != http.StatusBadRequest {
			t.Errorf("%s %s: esperado 400, obtenido %d", c.method, c.path, w.Code)
		}
	}
}

// TestCasosBordeNegocio cubre las validaciones de negocio en los services.
func TestCasosBordeNegocio(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)
	adminToken, _ := createAdmin(t, db)
	cToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")
	c2Token, _ := registerUser(t, r, "Cli2", "cli2@test.com", "secret123")
	eventID := createEvent(t, r, adminToken, 5)

	// Comprar evento inexistente -> 404
	if w := doRequest(r, "POST", "/api/tickets", cToken, map[string]uint{"event_id": 99999}); w.Code != http.StatusNotFound {
		t.Errorf("compra evento inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Login sin password -> 400
	if w := doRequest(r, "POST", "/api/auth/login", "", map[string]string{"email": "x@y.com"}); w.Code != http.StatusBadRequest {
		t.Errorf("login sin password: esperado 400, obtenido %d", w.Code)
	}

	// Reporte de evento inexistente -> 404
	if w := doRequest(r, "GET", "/api/admin/events/99999/report", adminToken, nil); w.Code != http.StatusNotFound {
		t.Errorf("reporte inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Update de evento inexistente -> 404
	if w := doRequest(r, "PUT", "/api/admin/events/99999", adminToken, map[string]string{"title": "x"}); w.Code != http.StatusNotFound {
		t.Errorf("update inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Cancelar evento inexistente -> 404
	if w := doRequest(r, "DELETE", "/api/admin/events/99999", adminToken, nil); w.Code != http.StatusNotFound {
		t.Errorf("cancelar evento inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Compramos un ticket con Cli para probar errores de tickets
	w := doRequest(r, "POST", "/api/tickets", cToken, map[string]uint{"event_id": eventID})
	var buy struct {
		Data struct {
			ID uint `json:"ID"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &buy)
	tid := buy.Data.ID

	// Cli2 intenta cancelar la entrada de Cli -> 403 (no es el dueño)
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", tid), c2Token, nil); w.Code != http.StatusForbidden {
		t.Errorf("cancelar entrada ajena: esperado 403, obtenido %d", w.Code)
	}

	// Transferir a uno mismo -> 400
	if w := doRequest(r, "POST", fmt.Sprintf("/api/tickets/%d/transfer", tid), cToken,
		map[string]string{"target_email": "cli@test.com"}); w.Code != http.StatusBadRequest {
		t.Errorf("transferir a uno mismo: esperado 400, obtenido %d", w.Code)
	}

	// Transferir a email inexistente -> 404 (usuario destino no encontrado)
	if w := doRequest(r, "POST", fmt.Sprintf("/api/tickets/%d/transfer", tid), cToken,
		map[string]string{"target_email": "noexiste@test.com"}); w.Code != http.StatusNotFound {
		t.Errorf("transferir a inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Cancelar OK y luego intentar cancelar de nuevo -> 400
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", tid), cToken, nil); w.Code != http.StatusOK {
		t.Errorf("cancelar entrada propia: esperado 200, obtenido %d", w.Code)
	}
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", tid), cToken, nil); w.Code != http.StatusBadRequest {
		t.Errorf("cancelar dos veces: esperado 400, obtenido %d", w.Code)
	}
}
