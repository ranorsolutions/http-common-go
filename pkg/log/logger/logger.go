// Package logger provides a structured logging utility built on top of Logrus,
// with optional Gin middleware integration and contextual request logging.
package logger

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranorsolutions/http-common-go/pkg/log/formatter"
	"github.com/ranorsolutions/http-common-go/pkg/middleware/response"
	"github.com/sirupsen/logrus"
)

// Logger wraps a Logrus entry and provides convenience helpers for
// application-wide logging and request-scoped logging via Gin middleware.
type Logger struct {
	Entry *logrus.Entry
}

// New initializes a new Logger instance configured with the provided service
// name and version. It sets up a Logrus instance with a custom formatter.
//
// Example:
//
//	log, _ := logger.New("user-service", "1.0.0", true)
//	log.Info("service started")
func New(name, version string, forceColors bool) (*Logger, error) {
	log := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.TraceLevel,
		Hooks: make(logrus.LevelHooks), // ✅ prevents nil map panic
		Formatter: &formatter.Formatter{
			ForceColors:     forceColors,
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		},
	}

	return &Logger{
		Entry: log.WithFields(logrus.Fields{
			"service": name,
			"version": version,
		}),
	}, nil
}

// Format attaches standard HTTP request fields to the logger entry for
// contextual logging of incoming requests.
func (l *Logger) Format(r *http.Request) {
	if r == nil {
		return
	}

	l.Entry = l.Entry.WithFields(logrus.Fields{
		"method":   r.Method,
		"origin":   r.RemoteAddr,
		"agent":    r.Header.Get("User-Agent"),
		"size":     r.ContentLength,
		"resource": r.URL.Path,
	})
}

// Middleware returns a Gin middleware that wraps requests with
// structured logging and adds automatic request ID correlation.
//
// It logs:
//   - request_id
//   - method, path, client IP
//   - status code and latency
func (log *Logger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// -------------------------------------------------------------------
		// 1. Handle request ID
		reqID := c.Request.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Writer.Header().Set("X-Request-ID", reqID)

		// -------------------------------------------------------------------
		// 2. Handle W3C Trace Context (traceparent)
		traceParent := c.Request.Header.Get("traceparent")
		var traceID, spanID string

		if traceParent == "" {
			// Generate new trace/span IDs
			traceID = strings.ReplaceAll(uuid.New().String(), "-", "")
			traceID = traceID[:32] // 16 bytes, hex-encoded
			spanID = strings.ReplaceAll(uuid.New().String(), "-", "")
			spanID = spanID[:16]
			traceParent = "00-" + traceID + "-" + spanID + "-01"
		} else {
			// Parse traceparent header
			parts := strings.Split(traceParent, "-")
			if len(parts) >= 4 {
				traceID = parts[1]
				spanID = parts[2]
			} else {
				traceID = strings.ReplaceAll(uuid.New().String(), "-", "")[:32]
				spanID = strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
				traceParent = "00-" + traceID + "-" + spanID + "-01"
			}
		}

		// Always include trace headers in response for propagation
		c.Writer.Header().Set("traceparent", traceParent)
		if state := c.Request.Header.Get("tracestate"); state != "" {
			c.Writer.Header().Set("tracestate", state)
		}

		// -------------------------------------------------------------------
		// 3. Prepare writer and contextual logger
		rw := response.NewWriter(c.Writer)
		c.Writer = rw

		reqLogger := log.Entry.WithFields(map[string]interface{}{
			"request_id": reqID,
			"trace_id":   traceID,
			"span_id":    spanID,
		})

		// Save to Gin context
		c.Set("logger", log)
		c.Set("request_id", reqID)
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)
		c.Set("logger_entry", reqLogger)

		// -------------------------------------------------------------------
		// 4. Log start of request
		reqLogger.WithFields(map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}).Debug("Request Received")

		// Process the request
		c.Next()

		// -------------------------------------------------------------------
		// 5. Log completion
		duration := time.Since(start)
		status := rw.Status()

		reqLogger.WithFields(map[string]interface{}{
			"status":   status,
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"clientIP": c.ClientIP(),
			"latency":  duration.String(),
		}).Info("request completed")
	}
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Entry.Info(fmt.Sprintf(msg, args...))
}

// Error logs a message at error level.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Entry.Error(fmt.Sprintf(msg, args...))
}

// Warn logs a message at warning level.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Entry.Warn(fmt.Sprintf(msg, args...))
}

// Fatal logs a message at fatal level and exits the program.
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Entry.Fatal(fmt.Sprintf(msg, args...))
}

// Trace logs a message at trace level.
func (l *Logger) Trace(msg string, args ...interface{}) {
	l.Entry.Trace(fmt.Sprintf(msg, args...))
}

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Entry.Debug(fmt.Sprintf(msg, args...))
}
