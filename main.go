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
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"event-ticketing-system/internal/database"
	"event-ticketing-system/internal/handlers"
	"event-ticketing-system/internal/middleware"
	"event-ticketing-system/internal/models"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it:", err)
	}

	// Initialize Gorilla Mux router
	r := mux.NewRouter()

	// Initialize database connection
	db := database.InitDB()
	if db != nil {
		defer db.Close()

		// Auto-migrate the schema
		db.AutoMigrate(&models.User{}, &models.Event{}, &models.Ticket{}, &models.AttendanceLog{})
	} else {
		log.Println("Warning: Database connection is not available. API endpoints requiring database will not work.")
	}

	// Add CORS middleware
	r.Use(middleware.CORSMiddleware)

	// Middleware to inject database into context
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), "db", db)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	// Setup routes
	setupRoutes(r, db)

	// Swagger JSON endpoint - serve dynamically from SWAGGER_URL environment variable
	swaggerFilePath := getSwaggerFilePath()
	if swaggerFilePath == "" {
		log.Fatal("SWAGGER_URL environment variable is required but not set")
	}
	r.Path("/docs/swagger.json").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swaggerFilePath)
	}))

	// Swagger UI routes
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			// Serve Swagger UI HTML page
			w.Header().Set("Content-Type", "text/html")
			html := `<!DOCTYPE html>
<html lang="en">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>Event Ticketing System API - Swagger UI</title>
	   <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
	   <style>
	       html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
	       *, *:before, *:after { box-sizing: inherit; }
	       body { margin:0; background: #fafafa; }
	   </style>
</head>
<body>
	   <div id="swagger-ui"></div>
	   <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
	   <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
	   <script>
	       window.onload = function() {
	           const ui = SwaggerUIBundle({
	               url: '/docs/swagger.json',
	               dom_id: '#swagger-ui',
	               deepLinking: true,
	               presets: [
	                   SwaggerUIBundle.presets.apis,
	                   SwaggerUIStandalonePreset
	               ],
	               plugins: [
	                   SwaggerUIBundle.plugins.DownloadUrl
	               ],
	               layout: "StandaloneLayout"
	           });
	       };
	   </script>
</body>
</html>`
			w.Write([]byte(html))
		} else {
			// For other assets, redirect to CDN
			http.Redirect(w, r, "https://unpkg.com/swagger-ui-dist@4.15.5"+r.URL.Path, http.StatusMovedPermanently)
		}
	})))

	// Redirect root path to Swagger UI
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})

	// Get port from environment variable or default to 8000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger JSON available at http://localhost:%s/docs/swagger.json", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// setupRoutes configures all API routes
func setupRoutes(r *mux.Router, db *gorm.DB) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	eventHandler := handlers.NewEventHandler(db)
	ticketHandler := handlers.NewTicketHandler(db)

	// Public routes
	public := r.PathPrefix("/api").Subrouter()
	{
		// Authentication routes
		public.HandleFunc("/register", authHandler.Register).Methods("POST")
		public.HandleFunc("/login", authHandler.Login).Methods("POST")
		public.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	}

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.JWTAuth)
	{
		// Event routes (public for browsing, protected for creation)
		protected.HandleFunc("/events", eventHandler.GetEvents).Methods("GET")
		protected.HandleFunc("/events/{id}", eventHandler.GetEvent).Methods("GET")

		// Ticket routes
		protected.HandleFunc("/events/{id}/purchase", ticketHandler.PurchaseTicket).Methods("POST")
		protected.HandleFunc("/tickets", ticketHandler.GetTickets).Methods("GET")
		protected.HandleFunc("/tickets/{id}", ticketHandler.GetTicket).Methods("GET")
	}

	// Admin routes
	admin := r.PathPrefix("/api").Subrouter()
	admin.Use(middleware.JWTAuth)
	admin.Use(middleware.AdminAuth)
	{
		// Event management routes
		admin.HandleFunc("/events", eventHandler.CreateEvent).Methods("POST")
		admin.HandleFunc("/events/{id}", eventHandler.UpdateEvent).Methods("PUT")
		admin.HandleFunc("/events/{id}", eventHandler.DeleteEvent).Methods("DELETE")

		// Ticket validation routes
		admin.HandleFunc("/tickets/{id}/validate", ticketHandler.ValidateTicket).Methods("POST")

		// Attendee management routes
		admin.HandleFunc("/events/{id}/attendees", ticketHandler.GetEventAttendees).Methods("GET")
		admin.HandleFunc("/events/{id}/attendees/export", ticketHandler.ExportAttendees).Methods("GET")
	}
}

// getSwaggerFilePath returns the full file path for swagger.json based on SWAGGER_URL environment variable
func getSwaggerFilePath() string {
	// Get SWAGGER_URL from environment variable
	swaggerURL := os.Getenv("SWAGGER_URL")
	if swaggerURL == "" {
		log.Fatal("SWAGGER_URL environment variable is not set")
		return ""
	}

	// If it's a full URL, parse it and return just the path component
	if strings.HasPrefix(swaggerURL, "http://") || strings.HasPrefix(swaggerURL, "https://") {
		if u, err := url.Parse(swaggerURL); err == nil && u.Path != "" {
			return u.Path
		} else {
			log.Fatalf("Invalid SWAGGER_URL format: %s, error: %v", swaggerURL, err)
			return ""
		}
	}

	// If it's already a file path, validate it exists
	if _, err := os.Stat(swaggerURL); err != nil {
		log.Fatalf("Swagger file not found at path: %s, error: %v", swaggerURL, err)
		return ""
	}

	return swaggerURL
}