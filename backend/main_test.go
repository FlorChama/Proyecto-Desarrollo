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
