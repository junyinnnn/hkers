package user

import (
	"github.com/gin-gonic/gin"

	response "hkers-backend/internal/core"
	"hkers-backend/internal/middleware"
)

// RegisterUserRoutes registers user routes on the given router.
func RegisterUserRoutes(router *gin.Engine, svc *response.Container, jwtManager response.JWTManager) {
	_ = svc // Will be used when user service is added
	h := NewHandler()

	// API routes - require JWT authentication
	api := router.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtManager))
	{
		api.GET("/me", h.GetProfile)
	}
}
