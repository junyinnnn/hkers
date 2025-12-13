// config/config.go

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Server        ServerConfig
	SessionSecret string
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host string
	Port string
}

// Load reads configuration from environment variables.
// .env file is optional (useful for local development, not needed in Docker)
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	// This allows the app to work in Docker where env vars are set directly
	// Try multiple common locations for .env file
	envPaths := []string{".env", "../.env", "../../.env"}
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded .env file from: %s", path)
			break
		}
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		// Generate a warning but don't fail - useful for development
		log.Printf("WARNING: SESSION_SECRET not set, using default (INSECURE for production)")
		sessionSecret = "default-insecure-secret-change-in-production"
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"), // 0.0.0.0 allows access from outside container
			Port: getEnv("SERVER_PORT", "3000"),
		},
		SessionSecret: sessionSecret,
	}, nil
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
