// cmd/server/main.go

package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"hkers-backend/internal/app"
)

func init() {
	// Register custom types for gob encoding/decoding (used by sessions)
	// Must be registered before any session operations
	gob.Register(map[string]interface{}{})
}

func main() {
	// Set Gin mode from environment
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	// Load configuration
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Bootstrap all application components
	bootstrap, err := app.Bootstrap(cfg.SessionSecret)
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
