package routes

import (
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core"
	coreauth "hkers-backend/internal/core/auth"
	"hkers-backend/internal/http/handlers/user"
	"hkers-backend/internal/http/middleware"
)

// RegisterUserRoutes registers user routes on the given router.
func RegisterUserRoutes(router *gin.Engine, svc *core.Container, jwtManager *coreauth.JWTManager) {
	_ = svc // Will be used when user service is added
	h := user.NewHandler()

	// API routes - require JWT authentication
	api := router.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtManager))
	{
		api.GET("/me", h.GetProfile)
	}
}
