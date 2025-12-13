package response

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	db "hkers-backend/internal/db/generated"
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

// Container holds all application services.
// Pass this to handlers instead of individual services.
type Container struct {
	Auth AuthServiceInterface
	User UserServiceInterface
	// Add more services as needed:
	// Email *EmailService
}

// NewContainer creates a new service container.
func NewContainer(authSvc AuthServiceInterface, userSvc UserServiceInterface) *Container {
	return &Container{
		Auth: authSvc,
		User: userSvc,
	}
}
