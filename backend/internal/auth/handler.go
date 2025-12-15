package auth

import (
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"hkers-backend/internal/core/response"
	db "hkers-backend/internal/sqlc/generated"
	"hkers-backend/internal/user"
)

// Handler handles authentication-related HTTP requests.
type Handler struct {
	authService ServiceInterface
	userService user.ServiceInterface
	jwtManager  response.JWTManager
}

// NewHandler creates a new auth Handler instance.
func NewHandler(authService ServiceInterface, userService user.ServiceInterface, jwtManager response.JWTManager) HandlerInterface {
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
		response.Error(ctx, http.StatusServiceUnavailable, "OIDC authentication is not configured. Please configure OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL environment variables.")
		return
	}

	state, err := h.authService.GenerateState()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to generate state")
		return
	}

	codeVerifier, codeChallenge, err := h.authService.GeneratePKCE()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to generate PKCE verifier")
		return
	}

	// Save state in session for CSRF protection
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Set("code_verifier", codeVerifier)
	if err := session.Save(); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to save session")
		return
	}

	// Redirect to OIDC authorization URL
	ctx.Redirect(http.StatusTemporaryRedirect, h.authService.GetAuthURLWithPKCE(state, codeChallenge))
}

// Callback handles the OAuth2 callback from the OIDC provider.
// GET /auth/callback
func (h *Handler) Callback(ctx *gin.Context) {
	if h.authService == nil {
		response.Error(ctx, http.StatusServiceUnavailable, "OIDC authentication is not configured. Please configure OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL environment variables.")
		return
	}

	session := sessions.Default(ctx)

	// Verify state parameter to prevent CSRF
	if ctx.Query("state") != session.Get("state") {
		response.Error(ctx, http.StatusBadRequest, "Invalid state parameter")
		return
	}

	verifier, ok := session.Get("code_verifier").(string)
	if !ok || verifier == "" {
		response.Error(ctx, http.StatusBadRequest, "Missing PKCE verifier")
		return
	}

	// Exchange authorization code for tokens
	token, err := h.authService.ExchangeCodeWithPKCE(ctx.Request.Context(), ctx.Query("code"), verifier)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Failed to exchange authorization code")
		return
	}

	// Verify the ID token
	idToken, _, verifyErr := h.authService.VerifyIDToken(ctx.Request.Context(), token)
	if verifyErr != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to verify ID token")
		return
	}

	// Extract user profile from claims
	profile, profileErr := h.authService.ExtractClaims(idToken)
	if profileErr != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to extract claims")
		return
	}

	// Get subject identifier (unique user ID from the OIDC provider)
	oidcSub, ok := profile["sub"].(string)
	if !ok || oidcSub == "" {
		response.Error(ctx, http.StatusInternalServerError, "Invalid OIDC token: missing sub claim")
		return
	}

	// Check if user is allowed to login (must exist in database and be active)
	var dbUser *db.User
	if h.userService != nil {
		var validateErr error
		dbUser, validateErr = h.userService.ValidateOIDCLogin(ctx.Request.Context(), oidcSub)
		if validateErr != nil {
			if errors.Is(validateErr, user.ErrUserNotActive) {
				// User exists but is not activated - pending approval
				response.Error(ctx, http.StatusForbidden, "Your account is pending approval. Please contact an administrator.")
				return
			}
			if errors.Is(validateErr, user.ErrUserNotAllowed) {
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
					response.Error(ctx, http.StatusInternalServerError, "Failed to register user")
					return
				}

				if isNew {
					response.Error(ctx, http.StatusForbidden, "Your account has been registered and is pending approval. Please contact an administrator.")
				} else {
					response.Error(ctx, http.StatusForbidden, "Your account is not active. Please contact an administrator.")
				}
				return
			}
			// Other database errors
			response.Error(ctx, http.StatusInternalServerError, "Failed to validate user")
			return
		}
	} else {
		response.Error(ctx, http.StatusInternalServerError, "User service not configured")
		return
	}

	// Generate JWT token for the authenticated user
	jwtToken, jwtErr := h.jwtManager.GenerateToken(
		dbUser.ID,
		dbUser.Email.String,
		dbUser.OidcSub,
		dbUser.Username,
		dbUser.IsActive.Bool,
	)
	if jwtErr != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	// Clear temporary OIDC session data (state, verifier)
	session.Delete("state")
	session.Delete("code_verifier")
	if saveErr := session.Save(); saveErr != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to clear session")
		return
	}

	// Return JWT token and user info in response
	response.Success(ctx, http.StatusOK, gin.H{
		"access_token": jwtToken,
		"token_type":   "Bearer",
		"expires_in":   86400 * 7, // 7 days in seconds
		"user": gin.H{
			"id":           dbUser.ID,
			"email":        dbUser.Email.String,
			"username":     dbUser.Username,
			"oidc_sub":     dbUser.OidcSub,
			"is_active":    dbUser.IsActive,
			"trust_points": dbUser.TrustPoints,
			"created_at":   dbUser.CreatedAt,
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
		response.Success(ctx, http.StatusOK, gin.H{
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
		response.Error(ctx, http.StatusInternalServerError, "Failed to build logout URL")
		return
	}

	if ok {
		response.Success(ctx, http.StatusOK, gin.H{
			"message":    "Logged out successfully",
			"logout_url": logoutURL,
		})
	} else {
		response.Success(ctx, http.StatusOK, gin.H{
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
		response.Error(ctx, http.StatusUnauthorized, "Authorization header required")
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		response.Error(ctx, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}
	oldToken := authHeader[len(bearerPrefix):]

	// Refresh the token
	newToken, err := h.jwtManager.RefreshToken(oldToken)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Failed to refresh token")
		return
	}

	// Return new token
	response.Success(ctx, http.StatusOK, gin.H{
		"access_token": newToken,
		"token_type":   "Bearer",
		"expires_in":   86400 * 7, // 7 days in seconds
	})
}
