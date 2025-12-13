package auth

import (
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	coreauth "hkers-backend/internal/core/auth"
	coreuser "hkers-backend/internal/core/user"
	db "hkers-backend/internal/db/generated"
	"hkers-backend/internal/core"
)

// Handler handles authentication-related HTTP requests.
type Handler struct {
	authService *coreauth.Service
	userService *coreuser.Service
	jwtManager  *coreauth.JWTManager
}

// NewHandler creates a new auth Handler instance.
func NewHandler(authService *coreauth.Service, userService *coreuser.Service, jwtManager *coreauth.JWTManager) *Handler {
	return &Handler{
		authService: authService,
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// Login initiates the OAuth2 login flow.
// GET /auth/login
func (h *Handler) Login(ctx *gin.Context) {
	if h.authService == nil {
		core.Error(ctx, http.StatusServiceUnavailable, "OIDC authentication is not configured. Please configure OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL environment variables.")
		return
	}

	state, err := h.authService.GenerateState()
	if err != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to generate state")
		return
	}

	codeVerifier, codeChallenge, err := h.authService.GeneratePKCE()
	if err != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to generate PKCE verifier")
		return
	}

	// Save state in session for CSRF protection
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Set("code_verifier", codeVerifier)
	if err := session.Save(); err != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to save session")
		return
	}

	// Redirect to OIDC authorization URL
	ctx.Redirect(http.StatusTemporaryRedirect, h.authService.GetAuthURLWithPKCE(state, codeChallenge))
}

// Callback handles the OAuth2 callback from the OIDC provider.
// GET /auth/callback
func (h *Handler) Callback(ctx *gin.Context) {
	if h.authService == nil {
		core.Error(ctx, http.StatusServiceUnavailable, "OIDC authentication is not configured. Please configure OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL environment variables.")
		return
	}

	session := sessions.Default(ctx)

	// Verify state parameter to prevent CSRF
	if ctx.Query("state") != session.Get("state") {
		core.Error(ctx, http.StatusBadRequest, "Invalid state parameter")
		return
	}

	verifier, ok := session.Get("code_verifier").(string)
	if !ok || verifier == "" {
		core.Error(ctx, http.StatusBadRequest, "Missing PKCE verifier")
		return
	}

	// Exchange authorization code for tokens
	token, err := h.authService.ExchangeCodeWithPKCE(ctx.Request.Context(), ctx.Query("code"), verifier)
	if err != nil {
		core.Error(ctx, http.StatusUnauthorized, "Failed to exchange authorization code")
		return
	}

	// Verify the ID token
	idToken, _, verifyErr := h.authService.VerifyIDToken(ctx.Request.Context(), token)
	if verifyErr != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to verify ID token")
		return
	}

	// Extract user profile from claims
	profile, profileErr := h.authService.ExtractClaims(idToken)
	if profileErr != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to extract claims")
		return
	}

	// Get subject identifier (unique user ID from the OIDC provider)
	oidcSub, ok := profile["sub"].(string)
	if !ok || oidcSub == "" {
		core.Error(ctx, http.StatusInternalServerError, "Invalid OIDC token: missing sub claim")
		return
	}

	// Check if user is allowed to login (must exist in database and be active)
	var user *db.User
	if h.userService != nil {
		var validateErr error
		user, validateErr = h.userService.ValidateOIDCLogin(ctx.Request.Context(), oidcSub)
		if validateErr != nil {
			if errors.Is(validateErr, coreuser.ErrUserNotActive) {
				// User exists but is not activated - pending approval
				core.Error(ctx, http.StatusForbidden, "Your account is pending approval. Please contact an administrator.")
				return
			}
			if errors.Is(validateErr, coreuser.ErrUserNotAllowed) {
				// User doesn't exist in our system
				// Option 1: Auto-create as inactive (requires admin approval)
				email, _ := profile["email"].(string)
				nickname, _ := profile["nickname"].(string)
				if nickname == "" {
					nickname, _ = profile["name"].(string)
				}
				if nickname == "" {
					nickname = oidcSub // fallback to sub as username
				}

				_, isNew, createErr := h.userService.GetOrCreateOIDCUser(ctx.Request.Context(), oidcSub, nickname, email)
				if createErr != nil {
					core.Error(ctx, http.StatusInternalServerError, "Failed to register user")
					return
				}

				if isNew {
					core.Error(ctx, http.StatusForbidden, "Your account has been registered and is pending approval. Please contact an administrator.")
				} else {
					core.Error(ctx, http.StatusForbidden, "Your account is not active. Please contact an administrator.")
				}
				return
			}
			// Other database errors
			core.Error(ctx, http.StatusInternalServerError, "Failed to validate user")
			return
		}
	} else {
		core.Error(ctx, http.StatusInternalServerError, "User service not configured")
		return
	}

	// Generate JWT token for the authenticated user
	jwtToken, jwtErr := h.jwtManager.GenerateToken(
		user.ID,
		user.Email.String,
		user.OidcSub,
		user.Username,
		user.IsActive.Bool,
	)
	if jwtErr != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	// Clear temporary OIDC session data (state, verifier)
	session.Delete("state")
	session.Delete("code_verifier")
	if saveErr := session.Save(); saveErr != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to clear session")
		return
	}

	// Return JWT token and user info in response
	core.Success(ctx, http.StatusOK, gin.H{
		"access_token": jwtToken,
		"token_type":   "Bearer",
		"expires_in":   86400 * 7, // 7 days in seconds
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email.String,
			"username":     user.Username,
			"oidc_sub":     user.OidcSub,
			"is_active":    user.IsActive,
			"trust_points": user.TrustPoints,
			"created_at":   user.CreatedAt,
		},
	})
}

// Logout handles user logout.
// POST /auth/logout
func (h *Handler) Logout(ctx *gin.Context) {
	// For JWT-based auth, logout is handled client-side (delete token)
	// But we can still provide OIDC provider logout URL if needed

	// If OIDC is not configured, just return success
	if h.authService == nil {
		core.Success(ctx, http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
		return
	}

	// Build return URL (prefer configured post-logout redirect)
	returnToURL := h.authService.PostLogoutRedirect()
	if returnToURL == "" {
		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}
		returnToURL = scheme + "://" + ctx.Request.Host
	}

	// Get provider end-session URL (if configured)
	// Note: We can't get id_token from session anymore, so OIDC logout might be limited
	logoutURL, ok, err := h.authService.GetEndSessionURL(returnToURL, "")
	if err != nil {
		core.Error(ctx, http.StatusInternalServerError, "Failed to build logout URL")
		return
	}

	if ok {
		core.Success(ctx, http.StatusOK, gin.H{
			"message":    "Logged out successfully",
			"logout_url": logoutURL,
		})
	} else {
		core.Success(ctx, http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
	}
}

// RefreshToken handles JWT token refresh.
// POST /auth/refresh
func (h *Handler) RefreshToken(ctx *gin.Context) {
	// Get token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		core.Error(ctx, http.StatusUnauthorized, "Authorization header required")
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		core.Error(ctx, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}
	oldToken := authHeader[len(bearerPrefix):]

	// Refresh the token
	newToken, err := h.jwtManager.RefreshToken(oldToken)
	if err != nil {
		core.Error(ctx, http.StatusUnauthorized, "Failed to refresh token")
		return
	}

	// Return new token
	core.Success(ctx, http.StatusOK, gin.H{
		"access_token": newToken,
		"token_type":   "Bearer",
		"expires_in":   86400 * 7, // 7 days in seconds
	})
}
