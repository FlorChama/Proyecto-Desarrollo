package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ticketek-backend/domain"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
)

func routerConAuth() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/privado", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetUint("user_id")})
	})
	r.GET("/admin", AuthRequired(), AdminRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func pedir(r *gin.Engine, path, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthRequired(t *testing.T) {
	r := routerConAuth()

	// Sin token -> 401
	if w := pedir(r, "/privado", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("sin token: esperado 401, obtenido %d", w.Code)
	}
	// Token inválido -> 401
	if w := pedir(r, "/privado", "basura"); w.Code != http.StatusUnauthorized {
		t.Errorf("token inválido: esperado 401, obtenido %d", w.Code)
	}
	// Token válido -> 200
	token, _ := utils.GenerateToken(7, "u@test.com", domain.RoleClient)
	if w := pedir(r, "/privado", token); w.Code != http.StatusOK {
		t.Errorf("token válido: esperado 200, obtenido %d", w.Code)
	}
}

func TestAdminRequired(t *testing.T) {
	r := routerConAuth()

	// Cliente -> 403
	tokenCliente, _ := utils.GenerateToken(1, "cli@test.com", domain.RoleClient)
	if w := pedir(r, "/admin", tokenCliente); w.Code != http.StatusForbidden {
		t.Errorf("cliente en /admin: esperado 403, obtenido %d", w.Code)
	}
	// Admin -> 200
	tokenAdmin, _ := utils.GenerateToken(2, "adm@test.com", domain.RoleAdmin)
	if w := pedir(r, "/admin", tokenAdmin); w.Code != http.StatusOK {
		t.Errorf("admin en /admin: esperado 200, obtenido %d", w.Code)
	}
}
