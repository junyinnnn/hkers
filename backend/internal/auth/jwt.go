package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when the JWT token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the JWT token has expired
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the claims in our JWT token
type JWTClaims struct {
	UserID   int32  `json:"user_id"`   // Database user ID
	Email    string `json:"email"`     // User email
	OIDCSub  string `json:"oidc_sub"`  // OIDC subject identifier
	Username string `json:"username"`  // Username
	IsActive bool   `json:"is_active"` // Account active status
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a new JWT token for a user
func (m *JWTManager) GenerateToken(userID int32, email, oidcSub, username string, isActive bool) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		OIDCSub:  oidcSub,
		Username: username,
		IsActive: isActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if user account is still active
	if !claims.IsActive {
		return nil, errors.New("user account is not active")
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (m *JWTManager) RefreshToken(tokenString string) (string, error) {
	// First validate the existing token
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh even if expired, but not if invalid
		if !errors.Is(err, ErrExpiredToken) {
			return "", err
		}
		// For expired tokens, parse without validation to get claims
		token, parseErr := jwt.ParseWithClaims(
			tokenString,
			&JWTClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(m.secretKey), nil
			},
			jwt.WithoutClaimsValidation(),
		)
		if parseErr != nil {
			return "", parseErr
		}
		claims, _ = token.Claims.(*JWTClaims)
	}

	// Generate new token with same user data but new expiration
	return m.GenerateToken(claims.UserID, claims.Email, claims.OIDCSub, claims.Username, claims.IsActive)
}
