package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

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
	respondWithJSON(w, code, ErrorResponse{Error: message})
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

	respondWithJSON(w, http.StatusOK, SuccessResponse{Message: "Logged out successfully"})
}

// GET /api/heroes - Get all heroes
// @Summary Get all heroes
// @Description Retrieve all heroes from the database
// @Tags heroes
// @Accept json
// @Produce json
// @Success 200 {array} Hero
// @Router /api/heroes [get]
func getHeroes(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id, name, role, difficulty, created_at, updated_at FROM heroes ORDER BY id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch heroes")
		return
	}
	defer rows.Close()

	var heroes []Hero
	for rows.Next() {
		var hero Hero
		err := rows.Scan(&hero.ID, &hero.Name, &hero.Role, &hero.Difficulty, &hero.CreatedAt, &hero.UpdatedAt)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan hero data")
			return
		}
		heroes = append(heroes, hero)
	}

	if err = rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error iterating heroes")
		return
	}

	respondWithJSON(w, http.StatusOK, heroes)
}

// GET /api/heroes/{id} - Get hero by ID
// @Summary Get hero by ID
// @Description Retrieve a specific hero by ID
// @Tags heroes
// @Accept json
// @Produce json
// @Param id path int true "Hero ID"
// @Success 200 {object} Hero
// @Failure 404 {object} ErrorResponse
// @Router /api/heroes/{id} [get]
func getHeroByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid hero ID")
		return
	}

	var hero Hero
	err = DB.QueryRow("SELECT id, name, role, difficulty, created_at, updated_at FROM heroes WHERE id = $1", id).
		Scan(&hero.ID, &hero.Name, &hero.Role, &hero.Difficulty, &hero.CreatedAt, &hero.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Hero not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch hero")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, hero)
}

// POST /api/heroes - Create a new hero
// @Summary Create a new hero
// @Description Create a new hero in the database
// @Tags heroes
// @Accept json
// @Produce json
// @Param hero body HeroCreateRequest true "Hero data"
// @Success 201 {object} Hero
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/heroes [post]
func createHero(w http.ResponseWriter, r *http.Request) {
	var req HeroCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Role == "" || req.Difficulty == "" {
		respondWithError(w, http.StatusBadRequest, "Name, role, and difficulty are required")
		return
	}

	var hero Hero
	err := DB.QueryRow("INSERT INTO heroes (name, role, difficulty) VALUES ($1, $2, $3) RETURNING id, name, role, difficulty, created_at, updated_at",
		req.Name, req.Role, req.Difficulty).
		Scan(&hero.ID, &hero.Name, &hero.Role, &hero.Difficulty, &hero.CreatedAt, &hero.UpdatedAt)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create hero")
		return
	}

	respondWithJSON(w, http.StatusCreated, hero)
}

// PUT /api/heroes/{id} - Update a hero by ID
// @Summary Update hero by ID
// @Description Update an existing hero by ID
// @Tags heroes
// @Accept json
// @Produce json
// @Param id path int true "Hero ID"
// @Param hero body HeroUpdateRequest true "Hero data"
// @Success 200 {object} Hero
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/heroes/{id} [put]
func updateHero(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid hero ID")
		return
	}

	var req HeroUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Role == "" || req.Difficulty == "" {
		respondWithError(w, http.StatusBadRequest, "Name, role, and difficulty are required")
		return
	}

	var hero Hero
	err = DB.QueryRow("UPDATE heroes SET name = $1, role = $2, difficulty = $3 WHERE id = $4 RETURNING id, name, role, difficulty, created_at, updated_at",
		req.Name, req.Role, req.Difficulty, id).
		Scan(&hero.ID, &hero.Name, &hero.Role, &hero.Difficulty, &hero.CreatedAt, &hero.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Hero not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update hero")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, hero)
}

// DELETE /api/heroes/{id} - Delete a hero by ID
// @Summary Delete hero by ID
// @Description Delete an existing hero by ID
// @Tags heroes
// @Accept json
// @Produce json
// @Param id path int true "Hero ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/heroes/{id} [delete]
func deleteHero(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid hero ID")
		return
	}

	result, err := DB.Exec("DELETE FROM heroes WHERE id = $1", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete hero")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check deletion result")
		return
	}

	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Hero not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
