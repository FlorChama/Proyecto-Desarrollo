package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// doRequest ejecuta una petición HTTP contra el router de test.
func doRequest(r *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var buf io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	}
	req, _ := http.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// buyBody devuelve un body válido para comprar una entrada.
func buyBody(eventID uint) map[string]interface{} {
	return map[string]interface{}{
		"event_id":       eventID,
		"payment_method": "credit_card",
		"amount":         1000.0,
		"ticket_type":    "general",
	}
}

// registerUser registra un cliente y devuelve su token e ID.
func registerUser(t *testing.T, r *gin.Engine, name, email, pass string) (string, uint) {
	t.Helper()
	w := doRequest(r, "POST", "/api/auth/register", "", map[string]string{
		"name": name, "email": email, "password": pass,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("registro de %s: esperado 201, obtenido %d (%s)", email, w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Token string `json:"token"`
			User  struct {
				ID uint `json:"ID"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error parseando respuesta de registro: %v", err)
	}
	return resp.Data.Token, resp.Data.User.ID
}

// createEvent crea un evento como admin y devuelve su ID.
func createEvent(t *testing.T, r *gin.Engine, adminToken string, capacity int) uint {
	t.Helper()
	body := map[string]interface{}{
		"title":    "Recital de Prueba",
		"date":     time.Now().Add(48 * time.Hour),
		"venue":    "Estadio Test",
		"capacity": capacity,
		"price":    1000.0,
		"category": "nacional",
	}
	w := doRequest(r, "POST", "/api/admin/events", adminToken, body)
	if w.Code != http.StatusCreated {
		t.Fatalf("crear evento: esperado 201, obtenido %d (%s)", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			ID uint `json:"ID"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}

// ---- Auth ----

func TestRegisterYLogin(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	// Registro OK
	w := doRequest(r, "POST", "/api/auth/register", "", map[string]string{
		"name": "Ana", "email": "ana@test.com", "password": "secret123",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("registro: esperado 201, obtenido %d", w.Code)
	}

	// Email duplicado -> 409
	w = doRequest(r, "POST", "/api/auth/register", "", map[string]string{
		"name": "Ana2", "email": "ana@test.com", "password": "secret123",
	})
	if w.Code != http.StatusConflict {
		t.Errorf("email duplicado: esperado 409, obtenido %d", w.Code)
	}

	// Registro inválido (sin password) -> 400
	w = doRequest(r, "POST", "/api/auth/register", "", map[string]string{"email": "x@test.com"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("registro inválido: esperado 400, obtenido %d", w.Code)
	}

	// Login OK -> 200
	w = doRequest(r, "POST", "/api/auth/login", "", map[string]string{
		"email": "ana@test.com", "password": "secret123",
	})
	if w.Code != http.StatusOK {
		t.Errorf("login: esperado 200, obtenido %d", w.Code)
	}

	// Login con password incorrecta -> 401
	w = doRequest(r, "POST", "/api/auth/login", "", map[string]string{
		"email": "ana@test.com", "password": "incorrecta",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("login incorrecto: esperado 401, obtenido %d", w.Code)
	}
}

// ---- Eventos (público) ----

func TestEventosPublicos(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)
	adminToken, _ := createAdmin(t, db)

	id := createEvent(t, r, adminToken, 100)

	// Listado público sin token -> 200
	w := doRequest(r, "GET", "/api/events", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("listado: esperado 200, obtenido %d", w.Code)
	}

	// Filtro por categoría -> 200
	w = doRequest(r, "GET", "/api/events?category=nacional", "", nil)
	if w.Code != http.StatusOK {
		t.Errorf("filtro categoría: esperado 200, obtenido %d", w.Code)
	}

	// Detalle existente -> 200
	w = doRequest(r, "GET", fmt.Sprintf("/api/events/%d", id), "", nil)
	if w.Code != http.StatusOK {
		t.Errorf("detalle: esperado 200, obtenido %d", w.Code)
	}

	// Detalle inexistente -> 404
	w = doRequest(r, "GET", "/api/events/99999", "", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("detalle inexistente: esperado 404, obtenido %d", w.Code)
	}
}

// ---- Permisos / roles ----

func TestPermisosAdmin(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	clientToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")
	adminToken, _ := createAdmin(t, db)

	body := map[string]interface{}{
		"title": "X", "date": time.Now().Add(24 * time.Hour),
		"venue": "V", "capacity": 10, "price": 500.0,
	}

	// Sin token -> 401
	if w := doRequest(r, "POST", "/api/admin/events", "", body); w.Code != http.StatusUnauthorized {
		t.Errorf("admin sin token: esperado 401, obtenido %d", w.Code)
	}
	// Token de cliente -> 403
	if w := doRequest(r, "POST", "/api/admin/events", clientToken, body); w.Code != http.StatusForbidden {
		t.Errorf("admin con token cliente: esperado 403, obtenido %d", w.Code)
	}
	// Token de admin -> 201
	if w := doRequest(r, "POST", "/api/admin/events", adminToken, body); w.Code != http.StatusCreated {
		t.Errorf("admin con token admin: esperado 201, obtenido %d", w.Code)
	}
}

// ---- Flujo completo: compra, QR, mis entradas, traspaso ----

func TestCompraCancelacionYTraspaso(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	adminToken, _ := createAdmin(t, db)
	eventID := createEvent(t, r, adminToken, 5)

	c1Token, _ := registerUser(t, r, "Cli1", "cli1@test.com", "secret123")
	c2Token, c2ID := registerUser(t, r, "Cli2", "cli2@test.com", "secret123")

	// Comprar sin token -> 401
	if w := doRequest(r, "POST", "/api/tickets", "", buyBody(eventID)); w.Code != http.StatusUnauthorized {
		t.Errorf("compra sin token: esperado 401, obtenido %d", w.Code)
	}

	// Comprar con token -> 201 y genera QR
	w := doRequest(r, "POST", "/api/tickets", c1Token, buyBody(eventID))
	if w.Code != http.StatusCreated {
		t.Fatalf("compra: esperado 201, obtenido %d (%s)", w.Code, w.Body.String())
	}
	var buyResp struct {
		Data struct {
			ID     uint   `json:"ID"`
			QRCode string `json:"qr_code"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &buyResp)
	if buyResp.Data.QRCode == "" {
		t.Error("la compra debe generar un código QR")
	}
	ticketID := buyResp.Data.ID
	originalQR := buyResp.Data.QRCode

	// Mis entradas -> 200 con 1 ticket
	w = doRequest(r, "GET", "/api/tickets/my", c1Token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("mis entradas: esperado 200, obtenido %d", w.Code)
	}
	var myResp struct {
		Data []struct {
			ID uint `json:"ID"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &myResp)
	if len(myResp.Data) != 1 {
		t.Errorf("mis entradas: esperado 1 ticket, obtenido %d", len(myResp.Data))
	}

	// Traspasar a Cli2 -> 200 con QR nuevo y user_id de Cli2
	w = doRequest(r, "POST", fmt.Sprintf("/api/tickets/%d/transfer", ticketID), c1Token,
		map[string]string{"target_email": "cli2@test.com"})
	if w.Code != http.StatusOK {
		t.Fatalf("traspaso: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}
	var transferResp struct {
		Data struct {
			ID     uint   `json:"ID"`
			QRCode string `json:"qr_code"`
			UserID uint   `json:"user_id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &transferResp)
	if transferResp.Data.UserID != c2ID {
		t.Errorf("traspaso: el ticket debe pertenecer a Cli2 (%d), pertenece a %d", c2ID, transferResp.Data.UserID)
	}
	if transferResp.Data.QRCode == originalQR {
		t.Error("el traspaso debe generar un QR nuevo distinto al anterior")
	}

	// El nuevo ticket de Cli2 aparece en sus entradas
	w = doRequest(r, "GET", "/api/tickets/my", c2Token, nil)
	json.Unmarshal(w.Body.Bytes(), &myResp)
	if len(myResp.Data) != 1 {
		t.Errorf("entradas de Cli2 tras traspaso: esperado 1, obtenido %d", len(myResp.Data))
	}

	// Cli2 cancela su nueva entrada -> 200
	newTicketID := transferResp.Data.ID
	w = doRequest(r, "DELETE", fmt.Sprintf("/api/tickets/%d", newTicketID), c2Token, nil)
	if w.Code != http.StatusOK {
		t.Errorf("cancelación: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}
}

// ---- Sin stock ----

func TestCompraSinStock(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	adminToken, _ := createAdmin(t, db)
	eventID := createEvent(t, r, adminToken, 1)

	c1Token, _ := registerUser(t, r, "Cli1", "cli1@test.com", "secret123")
	c2Token, _ := registerUser(t, r, "Cli2", "cli2@test.com", "secret123")

	// Primera compra OK
	if w := doRequest(r, "POST", "/api/tickets", c1Token, buyBody(eventID)); w.Code != http.StatusCreated {
		t.Fatalf("primera compra: esperado 201, obtenido %d", w.Code)
	}
	// Segunda compra sin stock -> 400
	if w := doRequest(r, "POST", "/api/tickets", c2Token, buyBody(eventID)); w.Code != http.StatusBadRequest {
		t.Errorf("compra sin stock: esperado 400, obtenido %d", w.Code)
	}
}

// ---- Admin: actualizar y reporte ----

func TestActualizarEventoYReporte(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	adminToken, _ := createAdmin(t, db)
	eventID := createEvent(t, r, adminToken, 50)

	// Actualizar -> 200
	w := doRequest(r, "PUT", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken,
		map[string]interface{}{"title": "Recital Actualizado", "price": 1500.0})
	if w.Code != http.StatusOK {
		t.Errorf("actualizar evento: esperado 200, obtenido %d", w.Code)
	}

	// Una compra para que el reporte tenga datos
	cToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")
	doRequest(r, "POST", "/api/tickets", cToken, buyBody(eventID))

	// Reporte -> 200
	w = doRequest(r, "GET", fmt.Sprintf("/api/admin/events/%d/report", eventID), adminToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("reporte: esperado 200, obtenido %d", w.Code)
	}
	var rep struct {
		Data struct {
			TotalSold     int `json:"total_sold"`
			TotalCapacity int `json:"total_capacity"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &rep)
	if rep.Data.TotalSold != 1 {
		t.Errorf("reporte: esperado 1 vendida, obtenido %d", rep.Data.TotalSold)
	}

	// Cancelar evento -> 200
	w = doRequest(r, "DELETE", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("cancelar evento: esperado 200, obtenido %d", w.Code)
	}
}

// ---- Reset y cambio de contraseña ----

func TestResetPassword(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	// Registrar usuario
	doRequest(r, "POST", "/api/auth/register", "", map[string]string{
		"name": "User", "email": "user@test.com", "password": "original123",
	})

	// Reset con email inexistente -> 400
	w := doRequest(r, "POST", "/api/auth/reset-password", "", map[string]string{
		"email": "noexiste@test.com", "new_password": "nueva123",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("reset email inexistente: esperado 400, obtenido %d", w.Code)
	}

	// Reset OK -> 200
	w = doRequest(r, "POST", "/api/auth/reset-password", "", map[string]string{
		"email": "user@test.com", "new_password": "nueva123",
	})
	if w.Code != http.StatusOK {
		t.Errorf("reset OK: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}

	// Login con nueva contraseña -> 200
	w = doRequest(r, "POST", "/api/auth/login", "", map[string]string{
		"email": "user@test.com", "password": "nueva123",
	})
	if w.Code != http.StatusOK {
		t.Errorf("login post-reset: esperado 200, obtenido %d", w.Code)
	}
}

func TestChangePassword(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	token, _ := registerUser(t, r, "User", "user@test.com", "original123")

	// Sin token -> 401
	w := doRequest(r, "PUT", "/api/auth/change-password", "", map[string]string{
		"current_password": "original123", "new_password": "nueva123",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("change sin token: esperado 401, obtenido %d", w.Code)
	}

	// Contraseña actual incorrecta -> 400
	w = doRequest(r, "PUT", "/api/auth/change-password", token, map[string]string{
		"current_password": "incorrecta", "new_password": "nueva123",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("change contraseña incorrecta: esperado 400, obtenido %d", w.Code)
	}

	// Cambio OK -> 200
	w = doRequest(r, "PUT", "/api/auth/change-password", token, map[string]string{
		"current_password": "original123", "new_password": "nueva123",
	})
	if w.Code != http.StatusOK {
		t.Errorf("change OK: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}

	// Login con nueva contraseña -> 200
	w = doRequest(r, "POST", "/api/auth/login", "", map[string]string{
		"email": "user@test.com", "password": "nueva123",
	})
	if w.Code != http.StatusOK {
		t.Errorf("login post-change: esperado 200, obtenido %d", w.Code)
	}
}

// ---- Admin: listado completo y pagos ----

func TestGetAllAdminYPagos(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)

	adminToken, _ := createAdmin(t, db)
	createEvent(t, r, adminToken, 10)

	// Listado admin -> 200
	w := doRequest(r, "GET", "/api/admin/events", adminToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("listado admin: esperado 200, obtenido %d", w.Code)
	}

	// Listado admin sin token -> 401
	if w := doRequest(r, "GET", "/api/admin/events", "", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("listado admin sin token: esperado 401, obtenido %d", w.Code)
	}

	// Mis pagos -> 200
	cToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")
	eventID := createEvent(t, r, adminToken, 5)
	doRequest(r, "POST", "/api/tickets", cToken, buyBody(eventID))

	w = doRequest(r, "GET", "/api/tickets/payments", cToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("mis pagos: esperado 200, obtenido %d", w.Code)
	}
}

// ---- Cobertura adicional: errores de binding y filtros ----

func TestCoberturaAdicional(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestServer(db)
	adminToken, _ := createAdmin(t, db)
	cToken, _ := registerUser(t, r, "Cli", "cli@test.com", "secret123")
	eventID := createEvent(t, r, adminToken, 10)

	// GetAll con filtro available -> 200
	if w := doRequest(r, "GET", "/api/events?available=true", "", nil); w.Code != http.StatusOK {
		t.Errorf("filtro available: esperado 200, obtenido %d", w.Code)
	}

	// GetAll con búsqueda -> 200
	if w := doRequest(r, "GET", "/api/events?search=Recital", "", nil); w.Code != http.StatusOK {
		t.Errorf("búsqueda: esperado 200, obtenido %d", w.Code)
	}

	// Create sin campos obligatorios -> 400
	if w := doRequest(r, "POST", "/api/admin/events", adminToken, map[string]string{"title": "X"}); w.Code != http.StatusBadRequest {
		t.Errorf("create sin campos: esperado 400, obtenido %d", w.Code)
	}

	// GetMyTickets sin token -> 401
	if w := doRequest(r, "GET", "/api/tickets/my", "", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("mis entradas sin token: esperado 401, obtenido %d", w.Code)
	}

	// GetMyPayments sin token -> 401
	if w := doRequest(r, "GET", "/api/tickets/payments", "", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("mis pagos sin token: esperado 401, obtenido %d", w.Code)
	}

	// GetMyTickets con token -> 200
	if w := doRequest(r, "GET", "/api/tickets/my", cToken, nil); w.Code != http.StatusOK {
		t.Errorf("mis entradas con token: esperado 200, obtenido %d", w.Code)
	}

	// Update con campo inválido de capacity (negativo es ignorado, igual retorna 200)
	if w := doRequest(r, "PUT", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken,
		map[string]interface{}{"capacity": 20, "vip_price": 500.0}); w.Code != http.StatusOK {
		t.Errorf("update capacity+vip: esperado 200, obtenido %d", w.Code)
	}

	// Update con fecha
	if w := doRequest(r, "PUT", fmt.Sprintf("/api/admin/events/%d", eventID), adminToken,
		map[string]interface{}{"description": "nueva desc", "venue": "nuevo venue", "category": "internacional", "extra_dates": "2026-12-25"}); w.Code != http.StatusOK {
		t.Errorf("update campos adicionales: esperado 200, obtenido %d", w.Code)
	}

	// GetAllAdmin con token cliente -> 403
	if w := doRequest(r, "GET", "/api/admin/events", cToken, nil); w.Code != http.StatusForbidden {
		t.Errorf("admin events con cliente: esperado 403, obtenido %d", w.Code)
	}
}
