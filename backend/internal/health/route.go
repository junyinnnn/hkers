package health

import (
	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers base/public routes on the given router.
func RegisterHealthRoutes(router *gin.Engine) {
	router.GET("/", Handler)
	router.GET("/health", Handler)
	router.HEAD("/health", Handler)
}
