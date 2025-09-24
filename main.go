// Package main Event Ticketing System API
//
// This is a REST API for an Event Ticketing System built with Go and PostgreSQL.
//
// Terms Of Service: http://swagger.io/terms/
//
// Schemes: http, https
// Host: localhost:8080
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
	"log"
	"os"

	"event-ticketing-system/docs"
	"event-ticketing-system/internal/database"
	"event-ticketing-system/internal/middleware"
	"event-ticketing-system/internal/models"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	// Initialize Gin router
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSMiddleware())

	// Swagger host configuration
	docs.SwaggerInfo.Host = "localhost:8080"
	if port := os.Getenv("PORT"); port != "" {
		docs.SwaggerInfo.Host = "localhost:" + port
	}

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

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler, "/swagger/"))

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(r.Run(":" + port))
}