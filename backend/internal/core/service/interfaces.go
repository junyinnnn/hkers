package service

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	db "hkers-backend/internal/sqlc/generated"
)

// AuthServiceInterface defines the interface for authentication services
type AuthServiceInterface interface {
	GenerateState() (string, error)
	GeneratePKCE() (string, string, error)
	GetAuthURLWithPKCE(state, codeChallenge string) string
	ExchangeCodeWithPKCE(ctx context.Context, code, verifier string) (*oauth2.Token, error)
	VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, string, error)
	ExtractClaims(token *oidc.IDToken) (map[string]interface{}, error)
	PostLogoutRedirect() string
	GetEndSessionURL(returnTo, idToken string) (string, bool, error)
}

// UserServiceInterface defines the interface for user services
type UserServiceInterface interface {
	ValidateOIDCLogin(ctx context.Context, oidcSub string) (*db.User, error)
	GetOrCreateOIDCUser(ctx context.Context, oidcSub, nickname, email string) (*db.User, bool, error)
}

// AuthHandlerInterface defines the interface for authentication HTTP handlers
type AuthHandlerInterface interface {
	Login(ctx *gin.Context)
	Callback(ctx *gin.Context)
	Logout(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)
}

// UserHandlerInterface defines the interface for user HTTP handlers
type UserHandlerInterface interface {
	GetProfile(ctx *gin.Context)
}
