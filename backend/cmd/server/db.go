package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"hkers-backend/internal/database"
	redisconfig "hkers-backend/internal/redis"
)

// initDB creates and verifies a pgx pool based on configuration.
func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	dbConfig := database.LoadConfig()
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

// initRedis creates a Redis client and verifies connectivity with a PING.
func initRedis(ctx context.Context) (*redis.Client, error) {
	redisConfig := redisconfig.LoadConfig()
	opts := &redis.Options{
		Addr:      redisConfig.GetAddr(),
		Username:  redisConfig.Username,
		Password:  redisConfig.Password,
		DB:        redisConfig.DB,
		TLSConfig: redisConfig.GetTLSConfig(),
	}

	log.Printf("Connecting to Redis at %s", opts.Addr)
	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	log.Printf("Redis connection established successfully")
	return client, nil
}
