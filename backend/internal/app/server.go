package app

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/auth"
	"hkers-backend/internal/config"
	redisconfig "hkers-backend/internal/config/redis"
	"hkers-backend/internal/health"
	"hkers-backend/internal/middleware"
	"hkers-backend/internal/user"
)

// NewRouter configures the Gin engine with middleware and route groups.
func NewRouter(cfg *config.Config, authSvc auth.ServiceInterface, userSvc user.ServiceInterface) (*gin.Engine, error) {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(middleware.GetCORSConfig(&cfg.CORS)))

	// Session middleware using Redis (only for OIDC flow state/verifier)
	// Not used for authentication after JWT migration
	store, err := redis.NewStoreWithPool(redisconfig.NewRedisPool(&cfg.Redis), []byte(cfg.Server.SessionSecret))
	if err != nil {
		return nil, fmt.Errorf("create Redis session store: %w", err)
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 1 hour - only needed during OIDC flow
		HttpOnly: true,
		Secure:   cfg.Server.GinMode == "release", // Secure cookies in production
		SameSite: http.SameSiteLaxMode,
	})
	router.Use(sessions.Sessions("auth-session", store))

	// Create JWT manager for token-based authentication
	jwtManager := auth.NewJWTManager(cfg.Auth.JWT.Secret, cfg.Auth.JWT.Duration)

	// Register route groups
	health.RegisterHealthRoutes(router)
	auth.RegisterAuthRoutes(router, authSvc, userSvc, jwtManager)
	user.RegisterUserRoutes(router, jwtManager)

	return router, nil
}
