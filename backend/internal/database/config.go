package database

import (
	"os"
)

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// LoadConfig loads database configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		User:     getEnv("POSTGRES_USER", "pguser"),
		Password: getEnv("POSTGRES_PASSWORD", "pgpassword"),
		Name:     getEnv("POSTGRES_DB", "pgdb"),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}
}

// GetDSN returns the PostgreSQL connection string (lib/pq format).
func (d *Config) GetDSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

// GetConnString returns the PostgreSQL connection string (pgx format).
func (d *Config) GetConnString() string {
	return "postgres://" + d.User + ":" + d.Password +
		"@" + d.Host + ":" + d.Port +
		"/" + d.Name + "?sslmode=" + d.SSLMode
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
