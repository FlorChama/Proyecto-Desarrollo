package main

import (
	"fmt"
	"log"
	"os"
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

// buildDSN arma la cadena de conexión a MySQL leyendo las variables de entorno
// y aplicando los valores por defecto. Se separó de initDB para poder testear
// los valores por defecto sin necesidad de conectarse a una base real.
func buildDSN() string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if host == "" {
		host = "localhost"
	}
	if user == "" {
		user = "root"
	}
	if pass == "" {
		pass = "root"
	}
	if name == "" {
		name = "ticketek_db"
	}
	if port == "" {
		port = "3306"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)
}

func initDB() *gorm.DB {
	dsn := buildDSN()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Ticket{}, &domain.Payment{}); err != nil {
		log.Fatalf("Error en migraciones: %v", err)
	}

	log.Println("Base de datos conectada y migrada correctamente")
	return db
}

// seedAdmin crea un usuario administrador inicial si todavía no existe ninguno.
// El rol Administrador se provisiona acá (no desde el registro público), de modo
// que nadie pueda auto-asignarse permisos de admin al registrarse.
func seedAdmin(db *gorm.DB) {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@tickethub.com"
	}
	pass := os.Getenv("ADMIN_PASSWORD")
	if pass == "" {
		pass = "admin123"
	}

	var count int64
	db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&count)
	if count > 0 {
		return
	}

	admin := &domain.User{
		Name:     "Administrador",
		Email:    email,
		Password: utils.HashPassword(pass),
		Role:     domain.RoleAdmin,
	}
	if err := db.Create(admin).Error; err != nil {
		log.Printf("No se pudo crear el administrador inicial: %v", err)
		return
	}
	log.Printf("Administrador inicial creado: %s", email)
}

// SetupRouter arma el router con todos los DAOs, services, controllers y rutas a
// partir de una conexión a la base. Se separó de main() para poder construir el
// mismo router en los tests sin levantar el servidor ni depender de MySQL.
func SetupRouter(db *gorm.DB) *gin.Engine {
	// DAOs
	userDAO := dao.NewUserDAO(db)
	eventDAO := dao.NewEventDAO(db)
	ticketDAO := dao.NewTicketDAO(db)
	paymentDAO := dao.NewPaymentDAO(db)

	// Services
	authService := services.NewAuthService(userDAO)
	eventService := services.NewEventService(eventDAO)
	ticketService := services.NewTicketService(ticketDAO, eventDAO, userDAO, paymentDAO)

	// Controllers
	authCtrl := controllers.NewAuthController(authService)
	eventCtrl := controllers.NewEventController(eventService)
	ticketCtrl := controllers.NewTicketController(ticketService)

	// Router
	r := gin.Default()
	r.Static("/uploads", "./uploads")

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		// Auth (público)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
			auth.POST("/reset-password", authCtrl.ResetPassword)
			auth.PUT("/change-password", middleware.AuthRequired(), authCtrl.ChangePassword)
		}

		// Eventos (público)
		events := api.Group("/events")
		{
			events.GET("", eventCtrl.GetAll)
			events.GET("/:id", eventCtrl.GetByID)
		}

		// Tickets (cliente autenticado)
		tickets := api.Group("/tickets")
		tickets.Use(middleware.AuthRequired())
		{
			tickets.POST("", ticketCtrl.Buy)
			tickets.GET("/my", ticketCtrl.GetMyTickets)
			tickets.GET("/payments", ticketCtrl.GetMyPayments)
			tickets.DELETE("/:id", ticketCtrl.Cancel)
			tickets.POST("/:id/transfer", ticketCtrl.Transfer)
		}

		// Admin (admin autenticado)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.GET("/events", eventCtrl.GetAllAdmin)
			admin.POST("/events/upload", eventCtrl.UploadImage)
			admin.POST("/events", eventCtrl.Create)
			admin.PUT("/events/:id", eventCtrl.Update)
			admin.DELETE("/events/:id", eventCtrl.Cancel)
			admin.GET("/events/:id/report", ticketCtrl.GetEventReport)
		}
	}

	return r
}

func main() {
	db := initDB()
	seedAdmin(db)

	r := SetupRouter(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor iniciado en el puerto %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
