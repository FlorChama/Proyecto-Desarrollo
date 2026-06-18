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

// TestReporteCuentaCanceladas verifica que el reporte distingue entradas
// vendidas de canceladas.
func TestReporteCuentaCanceladas(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)
	adminToken, _ := createAdmin(t, db)
	eventID := createEvent(t, r, adminToken, 10)

	c1, _ := registerUser(t, r, "C1", "c1@test.com", "secret123")
	c2, _ := registerUser(t, r, "C2", "c2@test.com", "secret123")

	// Dos compras
	w := doRequest(r, "POST", "/api/tickets", c1, buyBody(eventID))
	var b struct {
		Data struct {
			ID uint `json:"ID"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &b)
	doRequest(r, "POST", "/api/tickets", c2, buyBody(eventID))

	// C1 cancela su entrada
	doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", b.Data.ID), c1, nil)

	// Reporte: 1 vendida, 1 cancelada
	w = doRequest(r, "GET", fmt.Sprintf("/api/admin/events/%d/report", eventID), adminToken, nil)
	var rep struct {
		Data struct {
			TotalSold      int `json:"total_sold"`
			TotalCancelled int `json:"total_cancelled"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &rep)
	if rep.Data.TotalSold != 1 {
		t.Errorf("total_sold: esperado 1, obtenido %d", rep.Data.TotalSold)
	}
	if rep.Data.TotalCancelled != 1 {
		t.Errorf("total_cancelled: esperado 1, obtenido %d", rep.Data.TotalCancelled)
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
	badBuy := map[string]interface{}{
		"event_id": uint(99999), "payment_method": "credit_card", "amount": 1000.0,
	}
	if w := doRequest(r, "POST", "/api/tickets", cToken, badBuy); w.Code != http.StatusNotFound {
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

	// Comprar ticket válido para probar errores de tickets
	w := doRequest(r, "POST", "/api/tickets", cToken, buyBody(eventID))
	if w.Code != http.StatusCreated {
		t.Fatalf("compra para tests: esperado 201, obtenido %d (%s)", w.Code, w.Body.String())
	}
	var buy struct {
		Data struct {
			ID uint `json:"ID"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &buy)
	tid := buy.Data.ID

	// Cli2 intenta cancelar la entrada de Cli -> 403
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", tid), c2Token, nil); w.Code != http.StatusForbidden {
		t.Errorf("cancelar entrada ajena: esperado 403, obtenido %d", w.Code)
	}

	// Transferir a uno mismo -> 400
	if w := doRequest(r, "POST", fmt.Sprintf("/api/tickets/%d/transfer", tid), cToken,
		map[string]string{"target_email": "cli@test.com"}); w.Code != http.StatusBadRequest {
		t.Errorf("transferir a uno mismo: esperado 400, obtenido %d", w.Code)
	}

	// Transferir a email inexistente -> 404
	if w := doRequest(r, "POST", fmt.Sprintf("/api/tickets/%d/transfer", tid), cToken,
		map[string]string{"target_email": "noexiste@test.com"}); w.Code != http.StatusNotFound {
		t.Errorf("transferir a inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Cancelar OK
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", tid), cToken, nil); w.Code != http.StatusOK {
		t.Errorf("cancelar entrada propia: esperado 200, obtenido %d", w.Code)
	}
	// Cancelar de nuevo -> 400 (ya cancelada)
	if w := doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", tid), cToken, nil); w.Code != http.StatusBadRequest {
		t.Errorf("cancelar dos veces: esperado 400, obtenido %d", w.Code)
	}
}
