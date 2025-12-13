package response

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTManager defines the interface for JWT token management
type JWTManager interface {
	GenerateToken(userID int32, email, oidcSub, username string, isActive bool) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(oldToken string) (string, error)
}

// JWTClaims represents the claims in our JWT token
type JWTClaims struct {
	UserID   int32  `json:"user_id"`   // Database user ID
	Email    string `json:"email"`     // User email
	OIDCSub  string `json:"oidc_sub"`  // OIDC subject identifier
	Username string `json:"username"`  // Username
	IsActive bool   `json:"is_active"` // Account active status
	jwt.RegisteredClaims
}

// Response represents a standard API response envelope.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success sends a successful JSON response.
func Success(ctx *gin.Context, statusCode int, data interface{}) {
	ctx.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// Error sends an error JSON response.
func Error(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}
