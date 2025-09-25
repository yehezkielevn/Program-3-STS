package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Hero represents a Mobile Legends hero
type Hero struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	Difficulty string `json:"difficulty"`
}

// Heroes slice to store heroes in memory
var heroes []Hero
var nextID int = 4 // Starting from 4 since we have 3 initial heroes

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
	// Initialize heroes data
	initHeroes()

	// Create router
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(corsMiddleware)

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/heroes", getHeroes).Methods("GET")
	api.HandleFunc("/heroes", createHero).Methods("POST")
	api.HandleFunc("/heroes/{id}", updateHero).Methods("PUT")
	api.HandleFunc("/heroes/{id}", deleteHero).Methods("DELETE")

	// Handle OPTIONS requests for all routes
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
	fmt.Println("  GET    /api/heroes     - Get all heroes")
	fmt.Println("  POST   /api/heroes     - Create new hero")
	fmt.Println("  PUT    /api/heroes/{id} - Update hero")
	fmt.Println("  DELETE /api/heroes/{id} - Delete hero")

	log.Fatal(http.ListenAndServe(":8080", router))
}
