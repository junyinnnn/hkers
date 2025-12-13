package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/auth"
	service "hkers-backend/internal/core/service"
	"hkers-backend/internal/health"
	"hkers-backend/internal/middleware"
	redisconfig "hkers-backend/internal/redis"
	"hkers-backend/internal/user"
)

// NewRouter configures the Gin engine with middleware and route groups.
func NewRouter(sessionSecret string, svc *service.Container) (*gin.Engine, error) {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(middleware.LoadCORSConfig().GetCORSConfig()))

	// Load Redis config for session store
	redisConfig := redisconfig.LoadConfig()

	// Session middleware using Redis (only for OIDC flow state/verifier)
	// Not used for authentication after JWT migration
	store, err := redis.NewStoreWithPool(redisConfig.NewRedisPool(), []byte(sessionSecret))
	if err != nil {
		return nil, fmt.Errorf("create Redis session store: %w", err)
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 1 hour - only needed during OIDC flow
		HttpOnly: true,
		Secure:   os.Getenv("GIN_MODE") == "release", // Secure cookies in production
		SameSite: http.SameSiteLaxMode,
	})
	router.Use(sessions.Sessions("auth-session", store))

	// Load JWT config
	jwtConfig := auth.LoadJWTConfig()

	// Create JWT manager for token-based authentication
	jwtManager := auth.NewJWTManager(jwtConfig.Secret, jwtConfig.Duration)

	// Register route groups
	health.RegisterHealthRoutes(router)
	auth.RegisterAuthRoutes(router, svc, jwtManager)
	user.RegisterUserRoutes(router, svc, jwtManager)

	return router, nil
}
