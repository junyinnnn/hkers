package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core/response"
)

// JWTAuth is a middleware that validates JWT tokens from Authorization header
func JWTAuth(jwtManager response.JWTManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			return
		}

		// Extract token from "Bearer <token>" format
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization header format. Expected: Bearer <token>",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Empty token",
			})
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			return
		}

		// Store claims in context for use in handlers
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("username", claims.Username)
		ctx.Set("oidc_sub", claims.OIDCSub)
		ctx.Set("is_active", claims.IsActive)

		ctx.Next()
	}
}

// GetUserIDFromContext retrieves the authenticated user ID from the context
func GetUserIDFromContext(ctx *gin.Context) (int32, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int32)
	return id, ok
}

// GetEmailFromContext retrieves the authenticated user email from the context
func GetEmailFromContext(ctx *gin.Context) (string, bool) {
	email, exists := ctx.Get("email")
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}

// GetUsernameFromContext retrieves the authenticated username from the context
func GetUsernameFromContext(ctx *gin.Context) (string, bool) {
	username, exists := ctx.Get("username")
	if !exists {
		return "", false
	}
	u, ok := username.(string)
	return u, ok
}
