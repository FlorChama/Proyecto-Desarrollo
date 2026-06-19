package tests

import (
	"testing"

	"ticketek-backend/clients"
	"ticketek-backend/controllers"
	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/middleware"
	"ticketek-backend/services"
	"ticketek-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB crea una base de datos SQLite en archivo temporal, aislada por test.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dir := t.TempDir()
	dbPath := dir + "/test.db"

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("error abriendo SQLite: %v", err)
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Ticket{}, &domain.Payment{}); err != nil {
		t.Fatalf("error en migraciones de test: %v", err)
	}

	// Cerramos la conexión al terminar el test para que el archivo temporal de
	// SQLite pueda eliminarse (en Windows queda bloqueado si sigue abierto).
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

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
			auth.POST("/reset-password", authCtrl.ResetPassword)
			auth.PUT("/change-password", middleware.AuthRequired(), authCtrl.ChangePassword)
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

		tickets.GET("/payments", ticketCtrl.GetMyPayments)

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.GET("/events", eventCtrl.GetAllAdmin)
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
