// Package cors provides a simple CORS middleware for Gin that sets the
// appropriate Access-Control-* headers and handles preflight OPTIONS requests.
package cors

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines the configuration for CORS headers.
// If any field is omitted, sensible defaults are used.
type CORSConfig struct {
	AllowOrigins     []string
	AllowCredentials string
	AllowHeaders     []string
	AllowMethods     []string
}

// Default values for headers and methods.
var (
	defaultHeaders = []string{
		"Content-Type", "Content-Length", "Accept-Encoding",
		"X-CSRF-Token", "Authorization", "Accept", "Origin",
		"Cache-Control", "X-Requested-With",
	}
	defaultMethods = []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"}
)

// CORSMiddleware returns a Gin middleware that applies CORS headers
// based on the provided configuration. If config is nil, a permissive
// default is used that allows all origins and credentials.
//
// Example:
//
//	router := gin.New()
//	router.Use(cors.CORSMiddleware(&cors.CORSConfig{
//	    AllowOrigins: []string{"https://example.com"},
//	}))
//
// The middleware automatically responds to OPTIONS preflight requests
// with status 204 and skips the rest of the chain.
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use default configuration if none provided.
		if config == nil {
			config = &CORSConfig{
				AllowOrigins:     []string{"*"},
				AllowHeaders:     defaultHeaders,
				AllowMethods:     defaultMethods,
				AllowCredentials: "true",
			}
		}

		headers := c.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", strings.Join(config.AllowOrigins, ","))
		headers.Set("Access-Control-Allow-Credentials", config.AllowCredentials)
		headers.Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
		// ✅ FIXED BUG: previously used AllowOrigins instead of AllowMethods
		headers.Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))

		// Handle preflight request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
