// Package main Event Ticketing System API
//
// This is a REST API for an Event Ticketing System built with Go and PostgreSQL.
//
// Terms Of Service: http://swagger.io/terms/
//
// Schemes: http, https
// Host: localhost:8000
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// SecurityDefinitions:
// Bearer:
//   type: apiKey
//   name: Authorization
//   in: header
//   description: "Enter the token in the format: Bearer {token}"
//
// swagger:meta
package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"event-ticketing-system/internal/database"
	"event-ticketing-system/internal/handlers"
	"event-ticketing-system/internal/middleware"
	"event-ticketing-system/internal/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it:", err)
	}

	// Initialize Gin router
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSMiddleware())

	// Initialize database connection
	db := database.InitDB()
	defer db.Close()

	// Auto-migrate the schema
	db.AutoMigrate(&models.User{}, &models.Event{}, &models.Ticket{}, &models.AttendanceLog{})

	// Middleware to inject database into context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Setup routes
	setupRoutes(r, db)

	// Swagger endpoint - Dynamic URL configuration
	swaggerURL := ginSwagger.URL(getSwaggerURL())
	log.Printf("Swagger documentation URL: %s", getSwaggerURL())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))

	// Additional endpoint for Swagger UI compatibility
	r.StaticFile("/docs/swagger.json", getSwaggerFilePath())

	// Get port from environment variable or default to 8000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(r.Run(":" + port))
}

// setupRoutes configures all API routes
func setupRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	eventHandler := handlers.NewEventHandler(db)
	ticketHandler := handlers.NewTicketHandler(db)

	// Public routes
	public := r.Group("/api")
	{
		// Authentication routes
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
		public.POST("/logout", authHandler.Logout)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		// Event routes (public for browsing, protected for creation)
		protected.GET("/events", eventHandler.GetEvents)
		protected.GET("/events/:id", eventHandler.GetEvent)

		// Ticket routes
		protected.POST("/events/:id/purchase", ticketHandler.PurchaseTicket)
		protected.GET("/tickets", ticketHandler.GetTickets)
		protected.GET("/tickets/:id", ticketHandler.GetTicket)
	}

	// Admin routes
	admin := r.Group("/api")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.AdminAuth())
	{
		// Event management routes
		admin.POST("/events", eventHandler.CreateEvent)
		admin.PUT("/events/:id", eventHandler.UpdateEvent)
		admin.DELETE("/events/:id", eventHandler.DeleteEvent)

		// Ticket validation routes
		admin.POST("/tickets/:id/validate", ticketHandler.ValidateTicket)

		// Attendee management routes
		admin.GET("/events/:id/attendees", ticketHandler.GetEventAttendees)
		admin.GET("/events/:id/attendees/export", ticketHandler.ExportAttendees)
	}
}

// getSwaggerURL returns the dynamic swagger URL based on environment variables or calculated paths
func getSwaggerURL() string {
	// Method 1: Environment Variable (highest priority)
	if swaggerURL := os.Getenv("SWAGGER_URL"); swaggerURL != "" {
		return swaggerURL
	}

	// Method 2: Dynamic path calculation based on executable location
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		// Go up from cmd/server to project root, then to docs
		projectRoot := filepath.Dir(filepath.Dir(execDir))
		swaggerPath := filepath.Join(projectRoot, "docs", "swagger.json")

		// Check if file exists
		if _, err := os.Stat(swaggerPath); err == nil {
			return fmt.Sprintf("file:///%s", filepath.ToSlash(swaggerPath))
		}
	}

	// Method 3: Relative path as fallback (original behavior)
	// This works when running from project root: go run ./cmd/server
	return "../../docs/swagger.json"
}

// getSwaggerFilePath returns the file path for swagger.json, handling URL parsing
func getSwaggerFilePath() string {
	// Method 1: Environment Variable (highest priority)
	if swaggerURL := os.Getenv("SWAGGER_URL"); swaggerURL != "" {
		// If it's a full URL, parse it and return just the path component
		if strings.HasPrefix(swaggerURL, "http://") || strings.HasPrefix(swaggerURL, "https://") {
			if u, err := url.Parse(swaggerURL); err == nil && u.Path != "" {
				return u.Path
			}
		}
		// If it's already a path, return as-is
		return swaggerURL
	}

	// Method 2: Dynamic path calculation based on executable location
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		// Go up from cmd/server to project root, then to docs
		projectRoot := filepath.Dir(filepath.Dir(execDir))
		swaggerPath := filepath.Join(projectRoot, "docs", "swagger.json")

		// Check if file exists
		if _, err := os.Stat(swaggerPath); err == nil {
			return swaggerPath
		}
	}

	// Method 3: Relative path as fallback (original behavior)
	// This works when running from project root: go run ./cmd/server
	return "../../docs/swagger.json"
}