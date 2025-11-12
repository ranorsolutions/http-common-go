package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// --- Helpers ---

func performRequest(method, path string, mw gin.HandlerFunc, config *CORSConfig) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(mw)
	router.GET(path, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// --- Tests ---

func TestCORSMiddleware_DefaultConfig(t *testing.T) {
	w := performRequest("GET", "/test", CORSMiddleware(nil), nil)

	headers := w.Header()

	if got := headers.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected * origin, got %q", got)
	}
	if got := headers.Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Allow-Credentials true, got %q", got)
	}

	// Headers should include known defaults
	allowHeaders := headers.Get("Access-Control-Allow-Headers")
	for _, h := range defaultHeaders {
		if !strings.Contains(allowHeaders, h) {
			t.Errorf("expected header %q in Allow-Headers, got %q", h, allowHeaders)
		}
	}

	// Methods
	allowMethods := headers.Get("Access-Control-Allow-Methods")
	for _, m := range defaultMethods {
		if !strings.Contains(allowMethods, m) {
			t.Errorf("expected method %q in Allow-Methods, got %q", m, allowMethods)
		}
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

func TestCORSMiddleware_CustomConfig(t *testing.T) {
	cfg := &CORSConfig{
		AllowOrigins:     []string{"https://example.com", "https://foo.com"},
		AllowHeaders:     []string{"X-Test", "X-Other"},
		AllowMethods:     []string{"GET", "POST"},
		AllowCredentials: "false",
	}

	w := performRequest("GET", "/test", CORSMiddleware(cfg), cfg)
	h := w.Header()

	if got := h.Get("Access-Control-Allow-Origin"); !strings.Contains(got, "example.com") {
		t.Errorf("expected custom origin list, got %q", got)
	}
	if got := h.Get("Access-Control-Allow-Credentials"); got != "false" {
		t.Errorf("expected Allow-Credentials false, got %q", got)
	}
	if got := h.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Test") {
		t.Errorf("expected custom headers in Allow-Headers, got %q", got)
	}
	if got := h.Get("Access-Control-Allow-Methods"); got != "GET,POST" {
		t.Errorf("expected custom methods in Allow-Methods, got %q", got)
	}
}

func TestCORSMiddleware_OptionsPreflight(t *testing.T) {
	w := performRequest("OPTIONS", "/test", CORSMiddleware(nil), nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", w.Code)
	}

	// Ensure chain was aborted (no body)
	if body := w.Body.String(); body != "" {
		t.Errorf("expected empty body for preflight, got %q", body)
	}
}

func TestCORSMiddleware_CallsNextForNonOptions(t *testing.T) {
	called := false
	mw := func(c *gin.Context) {
		called = true
		c.Next()
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware(nil), mw)

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !called {
		t.Error("expected Next() to be called for non-OPTIONS request")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}
