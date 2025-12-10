package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hkers-backend/internal/http/middleware"
	"hkers-backend/internal/http/response"
)

// Handler handles user-related HTTP requests.
type Handler struct {
	// Add user service dependency here when needed
	// userService *services.UserService
}

// NewHandler creates a new user Handler instance.
func NewHandler() *Handler {
	return &Handler{}
}

// GetProfile returns the authenticated user's profile.
// GET /user or GET /api/v1/me
// Note: JWT middleware has already validated the token and extracted claims
func (h *Handler) GetProfile(ctx *gin.Context) {
	// Get user info from JWT claims (set by JWT middleware)
	userID, _ := middleware.GetUserIDFromContext(ctx)
	email, _ := middleware.GetEmailFromContext(ctx)
	username, _ := middleware.GetUsernameFromContext(ctx)
	oidcSub, _ := ctx.Get("oidc_sub")
	isActive, _ := ctx.Get("is_active")

	// Return user profile from JWT claims
	response.Success(ctx, http.StatusOK, gin.H{
		"id":        userID,
		"email":     email,
		"username":  username,
		"oidc_sub":  oidcSub,
		"is_active": isActive,
	})
}
