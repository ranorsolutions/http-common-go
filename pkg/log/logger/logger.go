package logger

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dot0s/http-common-go/pkg/log/formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	Entry *logrus.Entry
}

// New -- Sets the application-wide logger
func New(name, version string, forceColors bool) (*Logger, error) {
	// Create a new logrus instance
	log := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.TraceLevel,
		Formatter: &formatter.Formatter{
			ForceColors:     forceColors,
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		},
	}

	// Return the new Logrus Instance for this service
	logger := &Logger{
		Entry: log.WithFields(logrus.Fields{
			"service": name,
			"version": version,
		}),
	}

	return logger, nil
}

// FormatLog -- Applies fields to the log output
func (l *Logger) Format(r *http.Request) {
	// Log all incoming messages
	l.Entry = l.Entry.WithFields(logrus.Fields{
		"method":   r.Method,
		"origin":   r.RemoteAddr,
		"agent":    r.Header.Get("User-Agent"),
		"size":     r.ContentLength,
		"resource": r.URL.Path,
	})
}

// Logger -- Attaches logging to all routes
func (l *Logger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attach the logger to the request obect
		c.Set("logger", l)

		// Add debug logs
		l.Format(c.Copy().Request)
		l.Entry.Debug("Request Received")

		c.Next()
	}

}

/**
 * Begin logging functions
 */
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Entry.Info(fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.Entry.Error(fmt.Sprintf(msg, args...))
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Entry.Warn(fmt.Sprintf(msg, args...))
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Entry.Fatal(fmt.Sprintf(msg, args...))
}

func (l *Logger) Trace(msg string, args ...interface{}) {
	l.Entry.Trace(fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Entry.Debug(fmt.Sprintf(msg, args...))
}
