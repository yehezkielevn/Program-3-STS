package main

import (
	"time"
)

// Hero represents a Mobile Legends hero with database fields
type Hero struct {
	ID         int       `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Role       string    `json:"role" db:"role"`
	Difficulty string    `json:"difficulty" db:"difficulty"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// HeroCreateRequest represents request for creating a new hero
type HeroCreateRequest struct {
	Name       string `json:"name" validate:"required"`
	Role       string `json:"role" validate:"required"`
	Difficulty string `json:"difficulty" validate:"required"`
}

// HeroUpdateRequest represents request for updating a hero
type HeroUpdateRequest struct {
	Name       string `json:"name" validate:"required"`
	Role       string `json:"role" validate:"required"`
	Difficulty string `json:"difficulty" validate:"required"`
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
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}