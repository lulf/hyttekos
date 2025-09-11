package main

import (
	"log"
	"os"

	"hyttekos/internal/config"
	"hyttekos/internal/database"
	"hyttekos/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(cfg.DatabasePath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start server
	srv := server.New(cfg, db)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}