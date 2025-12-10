package routes

import (
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core"
	coreauth "hkers-backend/internal/core/auth"
	authhandler "hkers-backend/internal/http/handlers/auth"
)

// RegisterAuthRoutes registers auth routes on the given router.
func RegisterAuthRoutes(router *gin.Engine, svc *core.Container, jwtManager *coreauth.JWTManager) {
	h := authhandler.NewHandler(svc.Auth, svc.User, jwtManager)

	// Auth routes under /auth
	auth := router.Group("/auth")
	{
		auth.GET("/login", h.Login)       // Initiates OIDC flow
		auth.GET("/callback", h.Callback) // Returns JWT token
		auth.POST("/logout", h.Logout)    // Client-side logout with optional OIDC logout URL
		auth.POST("/refresh", h.RefreshToken) // Refresh JWT token
	}
}
