package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// ServiceInterface defines the interface for authentication services
type ServiceInterface interface {
	GenerateState() (string, error)
	GeneratePKCE() (string, string, error)
	GetAuthURLWithPKCE(state, codeChallenge string) string
	ExchangeCodeWithPKCE(ctx context.Context, code, verifier string) (*oauth2.Token, error)
	VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, string, error)
	ExtractClaims(token *oidc.IDToken) (map[string]interface{}, error)
	PostLogoutRedirect() string
	GetEndSessionURL(returnTo, idToken string) (string, bool, error)
}

// HandlerInterface defines the interface for authentication HTTP handlers
type HandlerInterface interface {
	Login(ctx *gin.Context)
	Callback(ctx *gin.Context)
	Logout(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)
}
