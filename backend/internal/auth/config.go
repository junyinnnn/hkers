package auth

import (
	"os"
	"strings"
	"time"
)

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret   string
	Duration time.Duration
}

// OIDCConfig holds OpenID Connect configuration.
type OIDCConfig struct {
	Issuer                string
	ClientID              string
	ClientSecret          string
	RedirectURL           string
	Scopes                []string
	EndSessionURL         string
	PostLogoutRedirectURL string
}

// LoadJWTConfig loads JWT configuration from environment variables.
func LoadJWTConfig() JWTConfig {
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtDurationStr := getEnv("JWT_DURATION", "168h") // Default 7 days
	jwtDuration, err := time.ParseDuration(jwtDurationStr)
	if err != nil {
		jwtDuration = 7 * 24 * time.Hour // fallback to 7 days
	}

	return JWTConfig{
		Secret:   jwtSecret,
		Duration: jwtDuration,
	}
}

// LoadOIDCConfig loads OIDC configuration from environment variables.
func LoadOIDCConfig() OIDCConfig {
	oidcIssuer := os.Getenv("OIDC_ISSUER")
	oidcClientID := os.Getenv("OIDC_CLIENT_ID")
	oidcClientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	oidcRedirectURL := os.Getenv("OIDC_REDIRECT_URL")

	oidcScopes := strings.TrimSpace(os.Getenv("OIDC_SCOPES"))
	if oidcScopes == "" {
		oidcScopes = "openid,profile,email"
	}
	rawScopes := strings.Split(oidcScopes, ",")
	for i := range rawScopes {
		rawScopes[i] = strings.TrimSpace(rawScopes[i])
	}

	endSessionURL := strings.TrimSpace(os.Getenv("OIDC_END_SESSION_URL"))
	postLogoutRedirect := strings.TrimSpace(os.Getenv("OIDC_POST_LOGOUT_REDIRECT_URL"))

	return OIDCConfig{
		Issuer:                oidcIssuer,
		ClientID:              oidcClientID,
		ClientSecret:          oidcClientSecret,
		RedirectURL:           oidcRedirectURL,
		Scopes:                rawScopes,
		EndSessionURL:         endSessionURL,
		PostLogoutRedirectURL: postLogoutRedirect,
	}
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
