package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/services"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// banco arma un router con los controllers reales sobre SQLite. Un middleware de
// test toma el id de usuario del header X-User para simular la autenticación sin
// depender del middleware real (que se prueba en su propio paquete).
func banco(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

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

	authCtrl := NewAuthController(services.NewAuthService(userDAO))
	eventCtrl := NewEventController(services.NewEventService(eventDAO))
	ticketCtrl := NewTicketController(services.NewTicketService(ticketDAO, eventDAO, userDAO, paymentDAO))

	r := gin.New()
	// Middleware de test: si viene X-User, lo cargamos en el contexto.
	r.Use(func(c *gin.Context) {
		if h := c.GetHeader("X-User"); h != "" {
			id, _ := strconv.ParseUint(h, 10, 32)
			c.Set("user_id", uint(id))
		}
		c.Next()
	})

	r.POST("/auth/register", authCtrl.Register)
	r.POST("/auth/login", authCtrl.Login)
	r.POST("/auth/reset-password", authCtrl.ResetPassword)
	r.PUT("/auth/change-password", authCtrl.ChangePassword)

	r.GET("/events", eventCtrl.GetAll)
	r.GET("/events/:id", eventCtrl.GetByID)
	r.GET("/admin/events", eventCtrl.GetAllAdmin)
	r.POST("/admin/events", eventCtrl.Create)
	r.POST("/admin/events/upload", eventCtrl.UploadImage)
	r.PUT("/admin/events/:id", eventCtrl.Update)
	r.DELETE("/admin/events/:id", eventCtrl.Cancel)
	r.GET("/admin/events/:id/report", ticketCtrl.GetEventReport)

	r.POST("/tickets", ticketCtrl.Buy)
	r.GET("/tickets/my", ticketCtrl.GetMyTickets)
	r.GET("/tickets/payments", ticketCtrl.GetMyPayments)
	r.DELETE("/tickets/:id", ticketCtrl.Cancel)
	r.POST("/tickets/:id/transfer", ticketCtrl.Transfer)

	return r, db
}

func req(r *gin.Engine, method, path, user string, body interface{}) *httptest.ResponseRecorder {
	var buf io.Reader
	if body != nil {
		switch b := body.(type) {
		case string: // JSON crudo (para probar binding inválido)
			buf = bytes.NewBufferString(b)
		default:
			j, _ := json.Marshal(b)
			buf = bytes.NewBuffer(j)
		}
	}
	httpReq, _ := http.NewRequest(method, path, buf)
	httpReq.Header.Set("Content-Type", "application/json")
	if user != "" {
		httpReq.Header.Set("X-User", user)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)
	return w
}

// idDe extrae data.ID de una respuesta exitosa.
func idDe(w *httptest.ResponseRecorder) uint {
	var resp struct {
		Data struct {
			ID uint `json:"ID"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}

// crearAdminYEvento crea un usuario y un evento directo en la base.
func sembrar(t *testing.T, db *gorm.DB) (uint, uint) {
	t.Helper()
	u := &domain.User{Name: "Cli", Email: "cli@test.com", Password: utils.HashPassword("secret123"), Role: domain.RoleClient}
	db.Create(u)
	ev := &domain.Event{Title: "Recital", Date: time.Now().Add(48 * time.Hour), Venue: "Estadio",
		Capacity: 5, Available: 5, Category: "nacional", Price: 1000, Status: domain.EventStatusActive}
	db.Create(ev)
	return u.ID, ev.ID
}

// ---- Auth ----

func TestAuthController(t *testing.T) {
	r, _ := banco(t)

	// Register OK
	if w := req(r, "POST", "/auth/register", "", map[string]string{"name": "Ana", "email": "ana@test.com", "password": "secret123"}); w.Code != http.StatusCreated {
		t.Errorf("register: esperado 201, obtenido %d", w.Code)
	}
	// Register JSON inválido -> 400
	if w := req(r, "POST", "/auth/register", "", "{esto no es json"); w.Code != http.StatusBadRequest {
		t.Errorf("register json inválido: esperado 400, obtenido %d", w.Code)
	}
	// Register duplicado -> 409
	if w := req(r, "POST", "/auth/register", "", map[string]string{"name": "Ana", "email": "ana@test.com", "password": "secret123"}); w.Code != http.StatusConflict {
		t.Errorf("register duplicado: esperado 409, obtenido %d", w.Code)
	}

	// Login OK
	if w := req(r, "POST", "/auth/login", "", map[string]string{"email": "ana@test.com", "password": "secret123"}); w.Code != http.StatusOK {
		t.Errorf("login: esperado 200, obtenido %d", w.Code)
	}
	// Login json inválido -> 400
	if w := req(r, "POST", "/auth/login", "", "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("login json inválido: esperado 400, obtenido %d", w.Code)
	}
	// Login incorrecto -> 401
	if w := req(r, "POST", "/auth/login", "", map[string]string{"email": "ana@test.com", "password": "mala"}); w.Code != http.StatusUnauthorized {
		t.Errorf("login incorrecto: esperado 401, obtenido %d", w.Code)
	}

	// ResetPassword OK
	if w := req(r, "POST", "/auth/reset-password", "", map[string]string{"email": "ana@test.com", "new_password": "nueva123"}); w.Code != http.StatusOK {
		t.Errorf("reset: esperado 200, obtenido %d", w.Code)
	}
	// ResetPassword json inválido -> 400
	if w := req(r, "POST", "/auth/reset-password", "", "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("reset json inválido: esperado 400, obtenido %d", w.Code)
	}
	// ResetPassword email inexistente -> 400
	if w := req(r, "POST", "/auth/reset-password", "", map[string]string{"email": "nadie@test.com", "new_password": "nueva123"}); w.Code != http.StatusBadRequest {
		t.Errorf("reset inexistente: esperado 400, obtenido %d", w.Code)
	}

	// ChangePassword OK
	uid := idDeUsuario(r)
	if w := req(r, "PUT", "/auth/change-password", uid, map[string]string{"current_password": "nueva123", "new_password": "otra1234"}); w.Code != http.StatusOK {
		t.Errorf("change: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}
	// ChangePassword json inválido -> 400
	if w := req(r, "PUT", "/auth/change-password", uid, "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("change json inválido: esperado 400, obtenido %d", w.Code)
	}
	// ChangePassword con actual incorrecta -> 400
	if w := req(r, "PUT", "/auth/change-password", uid, map[string]string{"current_password": "mala", "new_password": "otra1234"}); w.Code != http.StatusBadRequest {
		t.Errorf("change incorrecta: esperado 400, obtenido %d", w.Code)
	}
}

// idDeUsuario registra a Ana en limpio... acá reusamos el login para obtener el ID.
func idDeUsuario(r *gin.Engine) string {
	w := req(r, "POST", "/auth/login", "", map[string]string{"email": "ana@test.com", "password": "nueva123"})
	var resp struct {
		Data struct {
			User struct {
				ID uint `json:"ID"`
			} `json:"user"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return strconv.FormatUint(uint64(resp.Data.User.ID), 10)
}

// ---- Event ----

func TestEventController(t *testing.T) {
	r, _ := banco(t)

	// Create OK
	w := req(r, "POST", "/admin/events", "", map[string]interface{}{
		"title": "Show", "date": time.Now().Add(24 * time.Hour), "venue": "V", "capacity": 10, "price": 1000.0, "category": "nacional",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: esperado 201, obtenido %d (%s)", w.Code, w.Body.String())
	}
	eventID := idDe(w)

	// Create json inválido -> 400
	if w := req(r, "POST", "/admin/events", "", "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("create json inválido: esperado 400, obtenido %d", w.Code)
	}
	// Create sin campos requeridos -> 400
	if w := req(r, "POST", "/admin/events", "", map[string]string{"title": "X"}); w.Code != http.StatusBadRequest {
		t.Errorf("create incompleto: esperado 400, obtenido %d", w.Code)
	}

	// GetAll OK
	if w := req(r, "GET", "/events", "", nil); w.Code != http.StatusOK {
		t.Errorf("getall: esperado 200, obtenido %d", w.Code)
	}
	// GetAll con query inválida (available no booleano) -> 400
	if w := req(r, "GET", "/events?available=siono", "", nil); w.Code != http.StatusBadRequest {
		t.Errorf("getall query inválida: esperado 400, obtenido %d", w.Code)
	}
	// GetAllAdmin OK
	if w := req(r, "GET", "/admin/events", "", nil); w.Code != http.StatusOK {
		t.Errorf("getalladmin: esperado 200, obtenido %d", w.Code)
	}

	// GetByID OK
	if w := req(r, "GET", fmt.Sprintf("/events/%d", eventID), "", nil); w.Code != http.StatusOK {
		t.Errorf("getbyid: esperado 200, obtenido %d", w.Code)
	}
	// GetByID id inválido -> 400
	if w := req(r, "GET", "/events/abc", "", nil); w.Code != http.StatusBadRequest {
		t.Errorf("getbyid inválido: esperado 400, obtenido %d", w.Code)
	}
	// GetByID inexistente -> 404
	if w := req(r, "GET", "/events/99999", "", nil); w.Code != http.StatusNotFound {
		t.Errorf("getbyid inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Update OK
	if w := req(r, "PUT", fmt.Sprintf("/admin/events/%d", eventID), "", map[string]interface{}{"title": "Editado"}); w.Code != http.StatusOK {
		t.Errorf("update: esperado 200, obtenido %d", w.Code)
	}
	// Update id inválido -> 400
	if w := req(r, "PUT", "/admin/events/abc", "", map[string]string{"title": "x"}); w.Code != http.StatusBadRequest {
		t.Errorf("update id inválido: esperado 400, obtenido %d", w.Code)
	}
	// Update json inválido -> 400
	if w := req(r, "PUT", fmt.Sprintf("/admin/events/%d", eventID), "", "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("update json inválido: esperado 400, obtenido %d", w.Code)
	}
	// Update inexistente -> 404
	if w := req(r, "PUT", "/admin/events/99999", "", map[string]string{"title": "x"}); w.Code != http.StatusNotFound {
		t.Errorf("update inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Cancel id inválido -> 400
	if w := req(r, "DELETE", "/admin/events/abc", "", nil); w.Code != http.StatusBadRequest {
		t.Errorf("cancel id inválido: esperado 400, obtenido %d", w.Code)
	}
	// Cancel inexistente -> 404
	if w := req(r, "DELETE", "/admin/events/99999", "", nil); w.Code != http.StatusNotFound {
		t.Errorf("cancel inexistente: esperado 404, obtenido %d", w.Code)
	}
	// Cancel OK
	if w := req(r, "DELETE", fmt.Sprintf("/admin/events/%d", eventID), "", nil); w.Code != http.StatusOK {
		t.Errorf("cancel: esperado 200, obtenido %d", w.Code)
	}
}

func TestUploadImage(t *testing.T) {
	r, _ := banco(t)

	// Sin archivo -> 400
	if w := req(r, "POST", "/admin/events/upload", "", nil); w.Code != http.StatusBadRequest {
		t.Errorf("upload sin archivo: esperado 400, obtenido %d", w.Code)
	}

	// Con archivo -> 200 (corremos dentro de un dir temporal para no ensuciar el repo)
	cwd, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(cwd)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("image", "foto.png")
	fw.Write([]byte("contenido-de-prueba"))
	mw.Close()

	httpReq, _ := http.NewRequest("POST", "/admin/events/upload", &buf)
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)
	if w.Code != http.StatusOK {
		t.Errorf("upload con archivo: esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}
}

// ---- Ticket ----

func TestTicketController(t *testing.T) {
	r, db := banco(t)
	uid, eventID := sembrar(t, db)
	user := strconv.FormatUint(uint64(uid), 10)

	// Buy OK
	w := req(r, "POST", "/tickets", user, map[string]interface{}{"event_id": eventID, "payment_method": "credit_card", "ticket_type": "general"})
	if w.Code != http.StatusCreated {
		t.Fatalf("buy: esperado 201, obtenido %d (%s)", w.Code, w.Body.String())
	}
	ticketID := idDe(w)

	// Buy json inválido -> 400
	if w := req(r, "POST", "/tickets", user, "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("buy json inválido: esperado 400, obtenido %d", w.Code)
	}
	// Buy evento inexistente -> 404
	if w := req(r, "POST", "/tickets", user, map[string]interface{}{"event_id": 99999, "payment_method": "credit_card"}); w.Code != http.StatusNotFound {
		t.Errorf("buy inexistente: esperado 404, obtenido %d", w.Code)
	}

	// GetMyTickets OK
	if w := req(r, "GET", "/tickets/my", user, nil); w.Code != http.StatusOK {
		t.Errorf("my tickets: esperado 200, obtenido %d", w.Code)
	}
	// GetMyPayments OK
	if w := req(r, "GET", "/tickets/payments", user, nil); w.Code != http.StatusOK {
		t.Errorf("my payments: esperado 200, obtenido %d", w.Code)
	}

	// Report id inválido -> 400
	if w := req(r, "GET", "/admin/events/abc/report", "", nil); w.Code != http.StatusBadRequest {
		t.Errorf("report id inválido: esperado 400, obtenido %d", w.Code)
	}
	// Report inexistente -> 404
	if w := req(r, "GET", "/admin/events/99999/report", "", nil); w.Code != http.StatusNotFound {
		t.Errorf("report inexistente: esperado 404, obtenido %d", w.Code)
	}
	// Report OK
	if w := req(r, "GET", fmt.Sprintf("/admin/events/%d/report", eventID), "", nil); w.Code != http.StatusOK {
		t.Errorf("report: esperado 200, obtenido %d", w.Code)
	}

	// Transfer id inválido -> 400
	if w := req(r, "POST", "/tickets/abc/transfer", user, map[string]string{"target_email": "x@test.com"}); w.Code != http.StatusBadRequest {
		t.Errorf("transfer id inválido: esperado 400, obtenido %d", w.Code)
	}
	// Transfer json inválido -> 400
	if w := req(r, "POST", fmt.Sprintf("/tickets/%d/transfer", ticketID), user, "{x"); w.Code != http.StatusBadRequest {
		t.Errorf("transfer json inválido: esperado 400, obtenido %d", w.Code)
	}
	// Transfer a email inexistente -> 404
	if w := req(r, "POST", fmt.Sprintf("/tickets/%d/transfer", ticketID), user, map[string]string{"target_email": "nadie@test.com"}); w.Code != http.StatusNotFound {
		t.Errorf("transfer inexistente: esperado 404, obtenido %d", w.Code)
	}

	// Cancel id inválido -> 400
	if w := req(r, "DELETE", "/tickets/abc", user, nil); w.Code != http.StatusBadRequest {
		t.Errorf("cancel id inválido: esperado 400, obtenido %d", w.Code)
	}
	// Cancel OK
	if w := req(r, "DELETE", fmt.Sprintf("/tickets/%d", ticketID), user, nil); w.Code != http.StatusOK {
		t.Errorf("cancel: esperado 200, obtenido %d", w.Code)
	}
}
