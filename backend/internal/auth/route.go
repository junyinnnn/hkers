package auth

import (
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core/response"
	"hkers-backend/internal/user"
)

// RegisterAuthRoutes registers auth routes on the given router.
func RegisterAuthRoutes(router *gin.Engine, authSvc ServiceInterface, userSvc user.ServiceInterface, jwtManager response.JWTManager) {
	h := NewHandler(authSvc, userSvc, jwtManager)

	// Auth routes under /auth
	auth := router.Group("/auth")
	{
		auth.GET("/login", h.Login)           // Initiates OIDC flow
		auth.GET("/callback", h.Callback)     // Returns JWT token
		auth.POST("/logout", h.Logout)        // Client-side logout with optional OIDC logout URL
		auth.POST("/refresh", h.RefreshToken) // Refresh JWT token
	}
}
