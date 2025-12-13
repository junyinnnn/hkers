package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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

// InitDB creates and verifies a pgx pool based on configuration.
func InitDB(ctx context.Context) (*pgxpool.Pool, error) {
	config := LoadConfig()
	log.Printf("Connecting to database at %s:%s", config.Host, config.Port)

	pool, err := pgxpool.New(ctx, config.GetConnString())
	if err != nil {
		return nil, err
	}

	// Verify connectivity early
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	log.Printf("Database connection established successfully")
	return pool, nil
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
