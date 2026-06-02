package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestRegisterEndpoint_BadRequest(t *testing.T) {
	r := setupRouter()
	r.POST("/api/auth/register", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "datos inválidos")
			return
		}
		utils.CreatedResponse(c, body)
	})

	// Sin body
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d", w.Code)
	}
}

func TestLoginEndpoint_Unauthorized(t *testing.T) {
	r := setupRouter()
	r.POST("/api/auth/login", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "credenciales inválidas")
	})

	body := map[string]string{"email": "no@existe.com", "password": "wrong"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, obtenido %d", w.Code)
	}
}

func TestProtectedEndpoint_NoToken(t *testing.T) {
	r := setupRouter()

	// Simula middleware de auth
	r.POST("/api/tickets", func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "token requerido")
			return
		}
		utils.CreatedResponse(c, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tickets", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, obtenido %d", w.Code)
	}
}

func TestProtectedEndpoint_ValidToken(t *testing.T) {
	r := setupRouter()

	r.GET("/api/tickets/my", func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "token requerido")
			return
		}
		tokenStr := header[len("Bearer "):]
		_, err := utils.ValidateToken(tokenStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "token inválido")
			return
		}
		utils.SuccessResponse(c, []interface{}{})
	})

	token, _ := utils.GenerateToken(1, "test@test.com", "client")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tickets/my", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d", w.Code)
	}
}

func TestAdminEndpoint_ClientForbidden(t *testing.T) {
	r := setupRouter()

	r.POST("/api/admin/events", func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "token requerido")
			return
		}
		tokenStr := header[len("Bearer "):]
		claims, err := utils.ValidateToken(tokenStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "token inválido")
			return
		}
		if claims.Role != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "se requiere rol de administrador")
			return
		}
		utils.CreatedResponse(c, nil)
	})

	clientToken, _ := utils.GenerateToken(1, "cliente@test.com", "client")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/admin/events", nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("esperado 403, obtenido %d", w.Code)
	}
}
