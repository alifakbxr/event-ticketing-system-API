package database

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func InitDB() *gorm.DB {
	var dsn string


	// Check if DATABASE_URL is provided (for Neon database)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		// Use the provided DATABASE_URL directly
		dsn = databaseURL
		log.Println("Using DATABASE_URL for database connection")
	} else {
		// Fallback to individual environment variables
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "password")
		dbname := getEnv("DB_NAME", "event_ticketing")

		// Create database connection string with SSL mode for Neon compatibility
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
			host, port, user, password, dbname)
		log.Println("Using individual environment variables for database connection")
	}

	// Connect to database
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test the connection
	if err := db.DB().Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected successfully")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}