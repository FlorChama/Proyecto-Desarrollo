package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ticketek-backend/domain"

	"github.com/gin-gonic/gin"
)

// ---- hash ----

func TestHashYCheckPassword(t *testing.T) {
	password := "miPassword123"
	hash := HashPassword(password)

	if hash == password {
		t.Error("el hash no debe ser igual a la contraseña original")
	}
	if hash == "" {
		t.Error("el hash no debe estar vacío")
	}
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword debe dar true con la contraseña correcta")
	}
	if CheckPassword("otraClave", hash) {
		t.Error("CheckPassword debe dar false con la contraseña incorrecta")
	}
}

// ---- jwt ----

func TestGenerateYValidateToken(t *testing.T) {
	token, err := GenerateToken(1, "test@test.com", domain.RoleClient)
	if err != nil {
		t.Fatalf("GenerateToken no debe fallar: %v", err)
	}
	if token == "" {
		t.Fatal("el token no debe estar vacío")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken no debe fallar: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("UserID esperado 1, obtenido %d", claims.UserID)
	}
	if claims.Role != domain.RoleClient {
		t.Errorf("Role esperado %s, obtenido %s", domain.RoleClient, claims.Role)
	}
}

func TestValidateTokenInvalido(t *testing.T) {
	if _, err := ValidateToken("token.invalido.abc"); err == nil {
		t.Error("ValidateToken debe fallar con un token inválido")
	}
}

// ---- qr ----

func TestGenerateQRCodeYData(t *testing.T) {
	data := GenerateTicketQRData(1, 2, 3)
	if data != "TICKET-1-USER-2-EVENT-3" {
		t.Errorf("data del QR inesperada: %s", data)
	}

	qr, err := GenerateQRCode(data)
	if err != nil {
		t.Fatalf("GenerateQRCode no debe fallar: %v", err)
	}
	if len(qr) == 0 {
		t.Error("el QR no debe estar vacío")
	}
}

// ---- response ----

func TestResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// SuccessResponse -> 200
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	SuccessResponse(c, gin.H{"ok": true})
	if w.Code != http.StatusOK {
		t.Errorf("SuccessResponse: esperado 200, obtenido %d", w.Code)
	}

	// CreatedResponse -> 201
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	CreatedResponse(c, gin.H{"ok": true})
	if w.Code != http.StatusCreated {
		t.Errorf("CreatedResponse: esperado 201, obtenido %d", w.Code)
	}

	// ErrorResponse -> el status que se le pase
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	ErrorResponse(c, http.StatusBadRequest, "ups")
	if w.Code != http.StatusBadRequest {
		t.Errorf("ErrorResponse: esperado 400, obtenido %d", w.Code)
	}
}

func TestHandleServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	casos := []struct {
		err  error
		code int
	}{
		{domain.ErrEventoNoEncontrado, http.StatusNotFound},
		{domain.ErrEntradaNoEncontrada, http.StatusNotFound},
		{domain.ErrUsuarioNoEncontrado, http.StatusNotFound},
		{domain.ErrSinPermiso, http.StatusForbidden},
		{errInesperado{}, http.StatusBadRequest},
	}
	for _, caso := range casos {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		HandleServiceError(c, caso.err)
		if w.Code != caso.code {
			t.Errorf("HandleServiceError(%v): esperado %d, obtenido %d", caso.err, caso.code, w.Code)
		}
	}
}

type errInesperado struct{}

func (errInesperado) Error() string { return "error de negocio cualquiera" }
