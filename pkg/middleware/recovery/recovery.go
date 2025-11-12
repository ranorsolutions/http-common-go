// Package recovery provides a Gin middleware that recovers from panics,
// logs the error (optionally including a stack trace), and returns a safe
// 500 response to the client. If a logger compatible with this package's
// Errorf interface is present in the Gin context under the key "logger",
// it will be used; otherwise the middleware falls back to the standard log.
package recovery

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// loggerIface is the minimal contract we need from a logger.
// Your pkg/log/logger.Logger satisfies this via its Error method.
// We use Errorf-style signature for formatted logging.
type loggerIface interface {
	Error(msg string, args ...interface{})
}

// RecoveryConfig controls the behavior of the recovery middleware.
type RecoveryConfig struct {
	// IncludeStack controls whether a stack trace is captured and logged.
	IncludeStack bool

	// ResponseJSON controls whether the response is JSON (true) or plain text (false).
	ResponseJSON bool

	// MaskErrorMessage is the body text returned to the client (never the raw panic).
	// Defaults to "internal server error" when empty.
	MaskErrorMessage string

	// OnPanic, if provided, is invoked after the panic is recovered but before the response is sent.
	// Use this to add metrics or custom tracing.
	OnPanic func(c *gin.Context, recovered any)
}

// DefaultConfig returns a permissive, production-safe configuration.
func DefaultConfig() *RecoveryConfig {
	return &RecoveryConfig{
		IncludeStack:     true,
		ResponseJSON:     true,
		MaskErrorMessage: "internal server error",
	}
}

// Middleware returns a Gin middleware that recovers from panics,
// logs, and returns a 500 with a safe message.
func Middleware(cfg *RecoveryConfig) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.MaskErrorMessage == "" {
		cfg.MaskErrorMessage = "internal server error"
	}

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Pick a logger if present
				var lg loggerIface
				if v, ok := c.Get("logger"); ok {
					if typed, ok2 := v.(loggerIface); ok2 {
						lg = typed
					}
				}

				// Log panic + optional stack
				if cfg.IncludeStack {
					stack := debug.Stack()
					if lg != nil {
						lg.Error("panic recovered: %v\n%s", r, string(stack))
					} else {
						log.Printf("panic recovered: %v\n%s", r, string(stack))
					}
				} else {
					if lg != nil {
						lg.Error("panic recovered: %v", r)
					} else {
						log.Printf("panic recovered: %v", r)
					}
				}

				// User callback
				if cfg.OnPanic != nil {
					cfg.OnPanic(c, r)
				}

				// Mark as recovered and reply
				c.Header("X-Recovered-From", "panic")
				if cfg.ResponseJSON {
					// Safe JSON payload
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": cfg.MaskErrorMessage,
					})
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, cfg.MaskErrorMessage)
				}
			}
		}()

		c.Next()
	}
}
