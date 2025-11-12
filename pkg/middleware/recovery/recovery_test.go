package recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranorsolutions/http-common-go/pkg/log/logger"
	"github.com/sirupsen/logrus"
)

// testHook captures logrus entries for assertions.
type testHook struct {
	entries []*logrus.Entry
}

func (h *testHook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *testHook) Fire(e *logrus.Entry) error {
	h.entries = append(h.entries, e)
	return nil
}

func TestRecovery_DefaultJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(Middleware(nil)) // defaults

	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if got := w.Header().Get("X-Recovered-From"); got != "panic" {
		t.Fatalf("expected X-Recovered-From header to be set, got %q", got)
	}
	// Body should include masked message
	if !contains(w.Body.String(), "internal server error") {
		t.Fatalf("expected masked error message, got %q", w.Body.String())
	}
}

func TestRecovery_PlainText_NoStack_CustomMessage_And_Callback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	cfg := &RecoveryConfig{
		IncludeStack:     false,
		ResponseJSON:     false,
		MaskErrorMessage: "oops",
		OnPanic: func(c *gin.Context, r any) {
			called = true
		},
	}

	r := gin.New()
	r.Use(Middleware(cfg))

	r.GET("/panic", func(c *gin.Context) {
		panic("kapow")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected OnPanic callback to be invoked")
	}
	if w.Body.String() != `"oops"` { // Gin JSON encodes plain values; middleware used AbortWithStatusJSON
		t.Fatalf("expected plain masked message JSON-encoded, got %q", w.Body.String())
	}
}

func TestRecovery_UsesInjectedLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create our package logger and attach a hook
	l, _ := logger.New("svc", "v1", true)
	h := &testHook{}
	l.Entry.Logger.AddHook(h)

	r := gin.New()
	// inject logger into context (simulating your logger middleware)
	r.Use(func(c *gin.Context) {
		c.Set("logger", l)
		c.Next()
	})
	r.Use(Middleware(&RecoveryConfig{
		IncludeStack: true,
	}))

	r.GET("/panic", func(c *gin.Context) {
		panic("boom with stack")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// Ensure an error-level log entry was produced
	found := false
	for _, e := range h.entries {
		if e.Level == logrus.ErrorLevel && contains(e.Message, "panic recovered") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error log entry containing 'panic recovered'")
	}
}

// contains is a tiny helper to avoid importing strings in multiple spots.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	// simple substring search
outer:
	for i := 0; i+len(sub) <= len(s); i++ {
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
