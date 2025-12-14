package app

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"hkers-backend/internal/auth"
	"hkers-backend/internal/config"
	databaseconfig "hkers-backend/internal/config/database"
	redisconfig "hkers-backend/internal/config/redis"
	service "hkers-backend/internal/core/service"
	"hkers-backend/internal/user"
)

// BootstrapResult contains all initialized components needed to run the server
type BootstrapResult struct {
	Database *pgxpool.Pool
	Redis    *redis.Client
	Service  *service.Container
	Router   *gin.Engine
}

// Bootstrap initializes all application components
func Bootstrap(cfg *config.Config) (*BootstrapResult, error) {

	ctx := context.Background()

	// Initialize database connection pool
	pool, err := databaseconfig.InitDB(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	// Initialize Redis client (used by session store)
	redisClient, err := redisconfig.InitRedis(ctx, &cfg.Redis)
	if err != nil {
		pool.Close()
		return nil, err
	}

	// Initialize services
	var authService *auth.Service
	if cfg.Auth.OIDC.Issuer != "" {
		log.Printf("Initializing OIDC service with issuer: %s", cfg.Auth.OIDC.Issuer)
		authService, err = auth.NewService(&cfg.Auth.OIDC)
		if err != nil {
			pool.Close()
			redisClient.Close()
			return nil, err
		}
		log.Printf("OIDC service initialized successfully")
	} else {
		log.Printf("OIDC not configured, skipping OIDC service initialization")
	}

	// Initialize user service
	userService := user.NewService(pool)

	// Create service container
	svc := service.NewContainer(authService, userService)

	// Setup router
	router, err := NewRouter(cfg, svc)
	if err != nil {
		pool.Close()
		redisClient.Close()
		return nil, err
	}

	return &BootstrapResult{
		Database: pool,
		Redis:    redisClient,
		Service:  svc,
		Router:   router,
	}, nil
}
