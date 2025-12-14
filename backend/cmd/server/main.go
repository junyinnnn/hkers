// cmd/server/main.go

package main

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hkers-backend/internal/app"
	"hkers-backend/internal/config"
)

func init() {
	// Register custom types for gob encoding/decoding (used by sessions)
	// Must be registered before any session operations
	gob.Register(map[string]interface{}{})
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode from configuration
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// Bootstrap all application components
	bootstrap, err := app.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}
	defer bootstrap.Database.Close()
	defer bootstrap.Redis.Close()

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server listening on http://%s/", addr)
	if err := http.ListenAndServe(addr, bootstrap.Router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
