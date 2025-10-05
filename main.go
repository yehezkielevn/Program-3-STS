package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "mobile-legends-api/docs"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Mobile Legends Heroes API
// @version 1.0
// @description REST API for managing Mobile Legends heroes with PostgreSQL database
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	if err := InitDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer DB.Close()

	// Create tables and insert initial data
	if err := CreateTables(); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	if err := InsertInitialData(); err != nil {
		log.Fatalf("Error inserting initial data: %v", err)
	}

	// Start token cleanup goroutine
	go cleanExpiredTokens()

	// Create router
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(corsMiddleware)

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Authentication routes (no auth required)
	api.HandleFunc("/login", login).Methods("POST")
	api.HandleFunc("/logout", logout).Methods("POST")

	// Heroes routes
	api.HandleFunc("/heroes", getHeroes).Methods("GET")
	api.HandleFunc("/heroes/{id}", getHeroByID).Methods("GET")
	api.HandleFunc("/heroes", authMiddleware(http.HandlerFunc(createHero)).ServeHTTP).Methods("POST")
	api.HandleFunc("/heroes/{id}", authMiddleware(http.HandlerFunc(updateHero)).ServeHTTP).Methods("PUT")
	api.HandleFunc("/heroes/{id}", authMiddleware(http.HandlerFunc(deleteHero)).ServeHTTP).Methods("DELETE")

	// Handle OPTIONS requests for all routes
	api.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}).Methods("OPTIONS")

	api.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}).Methods("OPTIONS")

	api.HandleFunc("/heroes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}).Methods("OPTIONS")

	api.HandleFunc("/heroes/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}).Methods("OPTIONS")

	// Get server port from environment
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  POST   /api/login      - Login")
	fmt.Println("  POST   /api/logout     - Logout")
	fmt.Println("  GET    /api/heroes     - Get all heroes")
	fmt.Println("  GET    /api/heroes/{id} - Get hero by ID")
	fmt.Println("  POST   /api/heroes     - Create new hero (Auth Required)")
	fmt.Println("  PUT    /api/heroes/{id} - Update hero (Auth Required)")
	fmt.Println("  DELETE /api/heroes/{id} - Delete hero (Auth Required)")
	fmt.Printf("  Swagger UI: http://localhost:%s/swagger/\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
