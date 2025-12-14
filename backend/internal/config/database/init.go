package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"hkers-backend/internal/config"
)

// InitDB creates and verifies a pgx pool based on configuration.
func InitDB(ctx context.Context, dbConfig *config.DatabaseConfig) (*pgxpool.Pool, error) {
	log.Printf("Connecting to database at %s:%s", dbConfig.Host, dbConfig.Port)

	pool, err := pgxpool.New(ctx, dbConfig.GetConnString())
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
