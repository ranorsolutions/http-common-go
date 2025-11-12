package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ranorsolutions/http-common-go/pkg/log/formatter"
	"github.com/sirupsen/logrus"
)

// testHook is a logrus hook that stores entries for inspection in tests.
type testHook struct {
	entries []*logrus.Entry
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
}

func newTestLogger() (*Logger, *testHook) {
	l, _ := New("test-service", "1.0.0", true)
	h := &testHook{}
	l.Entry.Logger.AddHook(h)
	return l, h
}

// --- New() and Format() tests ---

func TestNewAndFormat(t *testing.T) {
	tt := []struct {
		name   string
		url    string
		method string
	}{
		{name: "GET request", url: "http://localhost:8080/endpoint?param=value", method: "GET"},
		{name: "POST request", url: "http://localhost:8080/endpoint", method: "POST"},
	}

	defaultKeys := []string{"service", "version"}
	requestKeys := []string{"agent", "method", "origin", "resource", "size"}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := New("test-service", "1.0.0", true)

			// Check initial fields
			for _, key := range defaultKeys {
				if _, ok := logger.Entry.Data[key]; !ok {
					t.Fatalf("missing default key %s from entry", key)
				}
			}

			// Create a mock request
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			logger.Format(req)

			// Verify new request-specific fields exist
			for _, key := range requestKeys {
				if _, ok := logger.Entry.Data[key]; !ok {
					t.Errorf("expected key %q in logger fields after Format()", key)
				}
			}
		})
	}
}

// --- Middleware() tests ---

func TestLoggerMiddleware_LogsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	appLogger, _ := New("test-svc", "v1", true)

	r := gin.New()
	r.Use(appLogger.Middleware())
	r.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequestIDPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	appLogger, _ := New("svc", "v1", true)

	r := gin.New()
	r.Use(appLogger.Middleware())
	r.GET("/test", func(c *gin.Context) {
		id := c.GetString("request_id")
		if id == "" {
			t.Error("expected request_id in context")
		}
		c.String(http.StatusOK, id)
	})

	// Case 1: no request ID header (should generate one)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID to be set in response")
	}

	// Case 2: provided header should be echoed
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", "abc-123")
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("X-Request-ID"); got != "abc-123" {
		t.Errorf("expected X-Request-ID header to be preserved, got %q", got)
	}
}

// --- Log level method tests ---

func TestLogLevelMethods(t *testing.T) {
	logger, hook := newTestLogger()

	tests := []struct {
		name   string
		method func(string, ...interface{})
		level  logrus.Level
	}{
		{"Info", logger.Info, logrus.InfoLevel},
		{"Warn", logger.Warn, logrus.WarnLevel},
		{"Error", logger.Error, logrus.ErrorLevel},
		{"Trace", logger.Trace, logrus.TraceLevel},
		{"Debug", logger.Debug, logrus.DebugLevel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := "test " + tc.name
			tc.method(msg)

			found := false
			for _, e := range hook.entries {
				if e.Level == tc.level && e.Message == msg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %s log with message %q not found", tc.name, msg)
			}
		})
	}
}

// --- Fatal() test (override ExitFunc to prevent os.Exit) ---
func TestFatalMethod(t *testing.T) {
	logger, hook := newTestLogger()

	// Override ExitFunc to prevent os.Exit()
	called := false
	exitFunc := logger.Entry.Logger.ExitFunc
	logger.Entry.Logger.ExitFunc = func(int) { called = true }
	defer func() { logger.Entry.Logger.ExitFunc = exitFunc }()

	logger.Fatal("fatal message")

	found := false
	for _, e := range hook.entries {
		if e.Level == logrus.FatalLevel && e.Message == "fatal message" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected Fatal log message not found")
	}
	if !called {
		t.Error("expected ExitFunc to be called")
	}
}

// --- Defensive tests ---

func TestFormatHandlesNilRequest(t *testing.T) {
	logger, _ := newTestLogger()

	// Should not panic
	logger.Format(nil)
}

func TestNewSetsFormatterAndLevel(t *testing.T) {
	logger, _ := New("svc", "v1", true)

	if logger.Entry.Logger.Level != logrus.TraceLevel {
		t.Errorf("expected default level TraceLevel, got %v", logger.Entry.Logger.Level)
	}

	if _, ok := logger.Entry.Logger.Formatter.(*formatter.Formatter); !ok {
		t.Errorf("expected custom formatter to be set")
	}
}

func TestFormatAddsAllExpectedKeys(t *testing.T) {
	logger, _ := New("svc", "v1", true)
	req, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)
	req.Header.Set("User-Agent", "test-agent")

	logger.Format(req)
	keys := []string{"method", "origin", "agent", "size", "resource"}

	for _, k := range keys {
		if _, ok := logger.Entry.Data[k]; !ok {
			t.Errorf("expected key %s missing after Format", k)
		}
	}
}

// Sanity check to ensure default keys never disappear
func TestDefaultKeysPresent(t *testing.T) {
	logger, _ := New("svc", "v1", false)
	for _, k := range []string{"service", "version"} {
		if _, ok := logger.Entry.Data[k]; !ok {
			t.Errorf("missing default field %s", k)
		}
	}
}
