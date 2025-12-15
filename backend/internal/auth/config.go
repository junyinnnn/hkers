package auth

// This file previously contained LoadJWTConfig() and LoadOIDCConfig() functions.
// These have been removed in favor of the centralized configuration system.
// Use config.Config from internal/config/config.go instead.
//
// Example:
//   cfg, _ := config.Load()
//   jwtConfig := cfg.Auth.JWT
//   oidcConfig := cfg.Auth.OIDC
