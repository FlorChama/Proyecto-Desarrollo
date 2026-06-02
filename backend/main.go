package main

import (
	"fmt"
	"log"
	"os"
	"ticketek-backend/clients"
	"ticketek-backend/controllers"
	"ticketek-backend/dao"
	"ticketek-backend/domain"
	"ticketek-backend/middleware"
	"ticketek-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() *gorm.DB {
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.Event{}, &domain.Ticket{}); err != nil {
		log.Fatalf("Error en migraciones: %v", err)
	}

	log.Println("Base de datos conectada y migrada correctamente")
	return db
}

func main() {
	db := initDB()

	// DAOs
	userDAO := dao.NewUserDAO(db)
	eventDAO := dao.NewEventDAO(db)
	ticketDAO := dao.NewTicketDAO(db)

	// Clients
	emailClient := clients.NewEmailClient()

	// Services
	authService := services.NewAuthService(userDAO)
	eventService := services.NewEventService(eventDAO)
	ticketService := services.NewTicketService(ticketDAO, eventDAO, userDAO, emailClient)

	// Controllers
	authCtrl := controllers.NewAuthController(authService)
	eventCtrl := controllers.NewEventController(eventService)
	ticketCtrl := controllers.NewTicketController(ticketService)

	// Router
	r := gin.Default()

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
			tickets.DELETE("/:id", ticketCtrl.Cancel)
			tickets.POST("/:id/transfer", ticketCtrl.Transfer)
		}

		// Admin (admin autenticado)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.POST("/events", eventCtrl.Create)
			admin.PUT("/events/:id", eventCtrl.Update)
			admin.DELETE("/events/:id", eventCtrl.Cancel)
			admin.GET("/events/:id/report", ticketCtrl.GetEventReport)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor iniciado en el puerto %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
