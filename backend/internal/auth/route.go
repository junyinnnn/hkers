package auth

import (
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core/response"
	service "hkers-backend/internal/core/service"
)

// RegisterAuthRoutes registers auth routes on the given router.
func RegisterAuthRoutes(router *gin.Engine, svc *service.Container, jwtManager response.JWTManager) {
	var h service.AuthHandlerInterface = NewHandler(svc.Auth, svc.User, jwtManager)

	// Auth routes under /auth
	auth := router.Group("/auth")
	{
		auth.GET("/login", h.Login)           // Initiates OIDC flow
		auth.GET("/callback", h.Callback)     // Returns JWT token
		auth.POST("/logout", h.Logout)        // Client-side logout with optional OIDC logout URL
		auth.POST("/refresh", h.RefreshToken) // Refresh JWT token
	}
}
