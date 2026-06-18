package tests

import (
	"fmt"
	"os"
	"testing"

	"ticketek-backend/clients"
	"ticketek-backend/controllers"
	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/middleware"
	"ticketek-backend/services"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// getenv devuelve la variable de entorno o un valor por defecto.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// setupTestDB conecta a una base de datos MySQL de test (separada de la real),
// la crea si no existe y deja las tablas limpias en cada test.
// Si no hay MySQL disponible, el test se saltea (Skip) para no fallar el build.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := getenv("TEST_DB_HOST", "127.0.0.1")
	port := getenv("TEST_DB_PORT", "3306")
	user := getenv("TEST_DB_USER", "root")
	pass := getenv("TEST_DB_PASS", "root")
	name := getenv("TEST_DB_NAME", "ticketek_test")

	// Conexión sin base para poder crear la base de test.
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port)
	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		t.Skipf("MySQL no disponible para tests de integración: %v", err)
	}
	if err := rootDB.Exec("CREATE DATABASE IF NOT EXISTS " + name).Error; err != nil {
		t.Skipf("no se pudo crear la base de test: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("no se pudo conectar a la base de test: %v", err)
	}

	// Tablas frescas para aislar cada test.
	_ = db.Migrator().DropTable(&domain.Payment{}, &domain.Ticket{}, &domain.Event{}, &domain.User{})
	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Ticket{}, &domain.Payment{}); err != nil {
		t.Fatalf("error en migraciones de test: %v", err)
	}
	return db
}

// setupTestServer arma el router real (DAOs -> services -> controllers + middleware),
// igual que main.go, para probar la API de punta a punta con httptest.
func setupTestServer(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userDAO := dao.NewUserDAO(db)
	eventDAO := dao.NewEventDAO(db)
	ticketDAO := dao.NewTicketDAO(db)
	paymentDAO := dao.NewPaymentDAO(db)
	emailClient := clients.NewEmailClient() // SMTP vacío: el envío falla en segundo plano, no rompe los tests

	authService := services.NewAuthService(userDAO)
	eventService := services.NewEventService(eventDAO)
	ticketService := services.NewTicketService(ticketDAO, eventDAO, userDAO, paymentDAO, emailClient)

	authCtrl := controllers.NewAuthController(authService)
	eventCtrl := controllers.NewEventController(eventService)
	ticketCtrl := controllers.NewTicketController(ticketService)

	r := gin.New()
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
		}

		events := api.Group("/events")
		{
			events.GET("", eventCtrl.GetAll)
			events.GET("/:id", eventCtrl.GetByID)
		}

		tickets := api.Group("/tickets")
		tickets.Use(middleware.AuthRequired())
		{
			tickets.POST("", ticketCtrl.Buy)
			tickets.GET("/my", ticketCtrl.GetMyTickets)
			tickets.DELETE("/:id", ticketCtrl.Cancel)
			tickets.POST("/:id/transfer", ticketCtrl.Transfer)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.POST("/events", eventCtrl.Create)
			admin.PUT("/events/:id", eventCtrl.Update)
			admin.DELETE("/events/:id", eventCtrl.Cancel)
			admin.GET("/events/:id/report", ticketCtrl.GetEventReport)
		}
	}
	return r
}

// createAdmin crea un usuario administrador directo en la base y devuelve su token.
func createAdmin(t *testing.T, db *gorm.DB) (string, *domain.User) {
	t.Helper()
	admin := &domain.User{
		Name:     "Admin",
		Email:    "admin@test.com",
		Password: utils.HashPassword("admin123"),
		Role:     domain.RoleAdmin,
	}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("error creando admin: %v", err)
	}
	token, err := utils.GenerateToken(admin.ID, admin.Email, domain.RoleAdmin)
	if err != nil {
		t.Fatalf("error generando token de admin: %v", err)
	}
	return token, admin
}
