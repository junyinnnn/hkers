package user

import (
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core/response"
	"hkers-backend/internal/middleware"
)

// RegisterUserRoutes registers user routes on the given router.
func RegisterUserRoutes(router *gin.Engine, jwtManager response.JWTManager) {
	h := NewHandler()

	// API routes - require JWT authentication
	api := router.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtManager))
	{
		api.GET("/me", h.GetProfile)
	}
}
