package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ticketek-backend/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dbDeTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
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

func TestSeedAdmin(t *testing.T) {
	db := dbDeTest(t)

	// Primera vez: crea el admin
	seedAdmin(db)
	var count int64
	db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&count)
	if count != 1 {
		t.Fatalf("debe haber 1 admin, hay %d", count)
	}

	// Segunda vez: ya existe, no duplica
	seedAdmin(db)
	db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&count)
	if count != 1 {
		t.Errorf("seedAdmin no debe duplicar el admin, hay %d", count)
	}
}

func TestSeedAdminConErrorDeBD(t *testing.T) {
	// Base sin migrar: la tabla users no existe, así que db.Create falla.
	// seedAdmin debe registrar el error y no romper (no crea ningún admin).
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/sinmigrar.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("error abriendo SQLite: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

	seedAdmin(db) // no debe entrar en pánico aunque falle el insert
}

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := dbDeTest(t)

	r := SetupRouter(db)
	if r == nil {
		t.Fatal("SetupRouter no debe devolver nil")
	}

	// El listado público de eventos responde 200 -> el router quedó cableado.
	httpReq, _ := http.NewRequest("GET", "/api/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)
	if w.Code != http.StatusOK {
		t.Errorf("GET /api/events: esperado 200, obtenido %d", w.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := dbDeTest(t)
	r := SetupRouter(db)

	// Una petición OPTIONS (preflight de CORS) debe cortar con 204.
	httpReq, _ := http.NewRequest("OPTIONS", "/api/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)
	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS /api/events: esperado 204, obtenido %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("falta el header Access-Control-Allow-Origin")
	}
}

func TestBuildDSNPorDefecto(t *testing.T) {
	// Sin variables de entorno: se usan todos los valores por defecto.
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASS", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_PORT", "")

	dsn := buildDSN()
	esperado := "root:root@tcp(localhost:3306)/ticketek_db?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != esperado {
		t.Errorf("DSN por defecto incorrecto:\n  esperado: %s\n  obtenido: %s", esperado, dsn)
	}
}

func TestBuildDSNConVariables(t *testing.T) {
	// Con variables seteadas: el DSN debe reflejarlas.
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_USER", "usuario")
	t.Setenv("DB_PASS", "clave")
	t.Setenv("DB_NAME", "mibase")
	t.Setenv("DB_PORT", "3310")

	dsn := buildDSN()
	esperado := "usuario:clave@tcp(db:3310)/mibase?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != esperado {
		t.Errorf("DSN con variables incorrecto:\n  esperado: %s\n  obtenido: %s", esperado, dsn)
	}
}
