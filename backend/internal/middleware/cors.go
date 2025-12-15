package middleware

import (
	"time"

	"github.com/gin-contrib/cors"

	"hkers-backend/internal/config"
)

// GetCORSConfig returns the gin-contrib/cors Config based on the centralized CORSConfig.
func GetCORSConfig(corsConfig *config.CORSConfig) cors.Config {
	cfg := cors.Config{
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           time.Duration(corsConfig.MaxAge) * time.Second,
	}

	if corsConfig.AllowAllOrigins {
		cfg.AllowOriginFunc = func(origin string) bool {
			return true
		}
	} else if len(corsConfig.AllowOrigins) > 0 {
		cfg.AllowOrigins = corsConfig.AllowOrigins
	} else {
		// Default: allow all if nothing specified
		cfg.AllowOriginFunc = func(origin string) bool {
			return true
		}
	}

	return cfg
}
