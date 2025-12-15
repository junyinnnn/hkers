package user

import (
	"context"

	"github.com/gin-gonic/gin"

	db "hkers-backend/internal/sqlc/generated"
)

// ServiceInterface defines the interface for user services
type ServiceInterface interface {
	ValidateOIDCLogin(ctx context.Context, oidcSub string) (*db.User, error)
	GetOrCreateOIDCUser(ctx context.Context, oidcSub, nickname, email string) (*db.User, bool, error)
}

// HandlerInterface defines the interface for user HTTP handlers
type HandlerInterface interface {
	GetProfile(ctx *gin.Context)
}
