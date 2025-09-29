package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

// Hero represents a Mobile Legends hero
type Hero struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	Difficulty string `json:"difficulty"`
}

// User represents a user for authentication
type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Config represents the configuration file structure
type Config struct {
	Users []User `yaml:"users"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
}

// Heroes slice to store heroes in memory
var heroes []Hero
var nextID int = 4 // Starting from 4 since we have 3 initial heroes

// Authentication
var (
	validTokens = make(map[string]time.Time)
	tokenMutex  sync.RWMutex
	config      Config
)

// CORS middleware to add CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Authentication middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, "Token required")
			return
		}

		// Check if token is valid
		tokenMutex.RLock()
		_, exists := validTokens[token]
		tokenMutex.RUnlock()

		if !exists {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JSON response helper
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Error response helper
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Load configuration from config.yaml
func loadConfig() error {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	return nil
}

// Clean expired tokens (run in background)
func cleanExpiredTokens() {
	for {
		time.Sleep(30 * time.Minute) // Clean every 30 minutes
		tokenMutex.Lock()
		now := time.Now()
		for token, expiry := range validTokens {
			if now.After(expiry) {
				delete(validTokens, token)
			}
		}
		tokenMutex.Unlock()
	}
}

// POST /api/login - Login endpoint
func login(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate credentials
	var userFound bool
	for _, user := range config.Users {
		if user.Username == loginReq.Username && user.Password == loginReq.Password {
			userFound = true
			break
		}
	}

	if !userFound {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate token
	token := uuid.New().String()
	tokenMutex.Lock()
	validTokens[token] = time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	tokenMutex.Unlock()

	respondWithJSON(w, http.StatusOK, LoginResponse{Token: token})
}

// POST /api/logout - Logout endpoint
func logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Token required")
		return
	}

	// Remove token
	tokenMutex.Lock()
	delete(validTokens, token)
	tokenMutex.Unlock()

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// GET /api/heroes - Get all heroes
func getHeroes(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, heroes)
}

// POST /api/heroes - Create a new hero
func createHero(w http.ResponseWriter, r *http.Request) {
	var newHero Hero
	if err := json.NewDecoder(r.Body).Decode(&newHero); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if newHero.Name == "" || newHero.Role == "" || newHero.Difficulty == "" {
		respondWithError(w, http.StatusBadRequest, "Name, role, and difficulty are required")
		return
	}

	// Set ID and add to heroes slice
	newHero.ID = strconv.Itoa(nextID)
	nextID++
	heroes = append(heroes, newHero)

	respondWithJSON(w, http.StatusCreated, newHero)
}

// PUT /api/heroes/{id} - Update a hero by ID
func updateHero(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updatedHero Hero
	if err := json.NewDecoder(r.Body).Decode(&updatedHero); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if updatedHero.Name == "" || updatedHero.Role == "" || updatedHero.Difficulty == "" {
		respondWithError(w, http.StatusBadRequest, "Name, role, and difficulty are required")
		return
	}

	// Find and update the hero
	for i, hero := range heroes {
		if hero.ID == id {
			updatedHero.ID = id // Ensure ID remains the same
			heroes[i] = updatedHero
			respondWithJSON(w, http.StatusOK, updatedHero)
			return
		}
	}

	respondWithError(w, http.StatusNotFound, "Hero not found")
}

// DELETE /api/heroes/{id} - Delete a hero by ID
func deleteHero(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	for i, hero := range heroes {
		if hero.ID == id {
			// Remove the hero from slice
			heroes = append(heroes[:i], heroes[i+1:]...)
			respondWithJSON(w, http.StatusOK, map[string]string{"message": "Hero deleted successfully"})
			return
		}
	}

	respondWithError(w, http.StatusNotFound, "Hero not found")
}

// Initialize initial heroes data
func initHeroes() {
	heroes = []Hero{
		{ID: "1", Name: "Alucard", Role: "Fighter", Difficulty: "Mudah"},
		{ID: "2", Name: "Miya", Role: "Marksman", Difficulty: "Mudah"},
		{ID: "3", Name: "Fanny", Role: "Assassin", Difficulty: "Sulit"},
	}
}

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize heroes data
	initHeroes()

	// Start token cleanup goroutine
	go cleanExpiredTokens()

	// Create router
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(corsMiddleware)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Authentication routes (no auth required)
	api.HandleFunc("/login", login).Methods("POST")
	api.HandleFunc("/logout", logout).Methods("POST")

	// Heroes routes (auth required)
	api.HandleFunc("/heroes", getHeroes).Methods("GET")
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

	// Start server
	fmt.Println("Server starting on port 8080...")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST   /api/login      - Login")
	fmt.Println("  POST   /api/logout     - Logout")
	fmt.Println("  GET    /api/heroes     - Get all heroes")
	fmt.Println("  POST   /api/heroes     - Create new hero (Auth Required)")
	fmt.Println("  PUT    /api/heroes/{id} - Update hero (Auth Required)")
	fmt.Println("  DELETE /api/heroes/{id} - Delete hero (Auth Required)")

	log.Fatal(http.ListenAndServe(":8080", router))
}
