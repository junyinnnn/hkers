package middleware

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
)

// CORSConfig holds CORS-related configuration.
type CORSConfig struct {
	AllowOrigins     []string
	AllowAllOrigins  bool
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// LoadCORSConfig loads CORS configuration from environment variables.
func LoadCORSConfig() CORSConfig {
	// Allow all origins by default (can be restricted via CORS_ALLOW_ORIGINS)
	allowAllOrigins := getEnv("CORS_ALLOW_ALL_ORIGINS", "true") == "true"

	var allowOrigins []string
	if !allowAllOrigins {
		originsStr := getEnv("CORS_ALLOW_ORIGINS", "")
		if originsStr != "" {
			allowOrigins = strings.Split(originsStr, ",")
			// Trim whitespace
			for i := range allowOrigins {
				allowOrigins[i] = strings.TrimSpace(allowOrigins[i])
			}
		}
	}

	// Allow methods
	methodsStr := getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
	allowMethods := strings.Split(methodsStr, ",")
	for i := range allowMethods {
		allowMethods[i] = strings.TrimSpace(allowMethods[i])
	}

	// Allow headers
	headersStr := getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization")
	allowHeaders := strings.Split(headersStr, ",")
	for i := range allowHeaders {
		allowHeaders[i] = strings.TrimSpace(allowHeaders[i])
	}

	// Expose headers
	exposeStr := getEnv("CORS_EXPOSE_HEADERS", "Content-Length")
	exposeHeaders := strings.Split(exposeStr, ",")
	for i := range exposeHeaders {
		exposeHeaders[i] = strings.TrimSpace(exposeHeaders[i])
	}

	// Allow credentials
	allowCredentials := getEnv("CORS_ALLOW_CREDENTIALS", "true") == "true"

	// Max age (in seconds, default 12 hours)
	maxAgeStr := getEnv("CORS_MAX_AGE", "43200")
	maxAge, _ := strconv.Atoi(maxAgeStr)
	if maxAge <= 0 {
		maxAge = 43200 // 12 hours default
	}

	return CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowAllOrigins:  allowAllOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
	}
}

// GetCORSConfig returns the gin-contrib/cors Config based on the CORSConfig.
func (c CORSConfig) GetCORSConfig() cors.Config {
	config := cors.Config{
		AllowMethods:     c.AllowMethods,
		AllowHeaders:     c.AllowHeaders,
		ExposeHeaders:    c.ExposeHeaders,
		AllowCredentials: c.AllowCredentials,
		MaxAge:           time.Duration(c.MaxAge) * time.Second,
	}

	if c.AllowAllOrigins {
		config.AllowOriginFunc = func(origin string) bool {
			return true
		}
	} else if len(c.AllowOrigins) > 0 {
		config.AllowOrigins = c.AllowOrigins
	} else {
		// Default: allow all if nothing specified
		config.AllowOriginFunc = func(origin string) bool {
			return true
		}
	}

	return config
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
