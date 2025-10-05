package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Database connection pool
var DB *sql.DB

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// InitDB initializes database connection with connection pooling
func InitDB() error {
	config := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "heroes_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// CreateTables creates the heroes table if it doesn't exist
func CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS heroes (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		role VARCHAR(100) NOT NULL,
		difficulty VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create trigger to update updated_at column
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	DROP TRIGGER IF EXISTS update_heroes_updated_at ON heroes;
	CREATE TRIGGER update_heroes_updated_at
		BEFORE UPDATE ON heroes
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Tables created successfully")
	return nil
}

// InsertInitialData inserts initial heroes data
func InsertInitialData() error {
	// Check if data already exists
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM heroes").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %v", err)
	}

	if count > 0 {
		log.Println("Initial data already exists, skipping insertion")
		return nil
	}

	heroes := []struct {
		name       string
		role       string
		difficulty string
	}{
		{"Alucard", "Fighter", "Mudah"},
		{"Miya", "Marksman", "Mudah"},
		{"Fanny", "Assassin", "Sulit"},
	}

	query := "INSERT INTO heroes (name, role, difficulty) VALUES ($1, $2, $3)"
	for _, hero := range heroes {
		_, err := DB.Exec(query, hero.name, hero.role, hero.difficulty)
		if err != nil {
			return fmt.Errorf("failed to insert hero %s: %v", hero.name, err)
		}
	}

	log.Println("Initial data inserted successfully")
	return nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
