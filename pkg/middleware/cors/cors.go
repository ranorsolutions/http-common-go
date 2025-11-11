package cors

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowCredentials string
	AllowHeaders     []string
	AllowMethods     []string
}

var (
	defaultHeaders = []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"}
	defaultMethods = []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"}
)

// CORSMiddleware -- Adds support for CORS headers
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config == nil {
			config = &CORSConfig{
				AllowOrigins:     []string{"*"},
				AllowHeaders:     defaultHeaders,
				AllowMethods:     defaultMethods,
				AllowCredentials: "true",
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(config.AllowOrigins, ","))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", config.AllowCredentials)
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowOrigins, ","))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
