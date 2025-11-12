package context

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// --- GinContextFromContext() tests ---

func TestGinContextFromContext_Success(t *testing.T) {
	// Create a Gin context with a test recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Store it in a Go context using the middleware key
	ctx := context.WithValue(context.Background(), ginContextKey, c)

	// Retrieve via helper
	retrieved := GinContextFromContext(ctx)
	if retrieved == nil {
		t.Fatal("expected non-nil gin.Context from context")
	}
	if retrieved != c {
		t.Fatal("expected retrieved gin.Context to match stored one")
	}
}

func TestGinContextFromContext_MissingKey(t *testing.T) {
	ctx := context.Background()
	got := GinContextFromContext(ctx)
	if got != nil {
		t.Fatal("expected nil when key is missing")
	}
}

func TestGinContextFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ginContextKey, "not-a-gin-context")
	got := GinContextFromContext(ctx)
	if got != nil {
		t.Fatal("expected nil when stored value is wrong type")
	}
}

func TestGinContextFromContext_NilContext(t *testing.T) {
	got := GinContextFromContext(nil)
	if got != nil {
		t.Fatal("expected nil when context is nil")
	}
}

// --- GinContextToContextMiddleware() tests ---

func TestGinContextToContextMiddleware_SetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(GinContextToContextMiddleware())

	called := false
	router.GET("/test", func(c *gin.Context) {
		called = true
		// retrieve gin.Context from Go context
		retrieved := GinContextFromContext(c.Request.Context())
		if retrieved == nil {
			t.Fatal("expected to retrieve gin.Context from request.Context")
		}
		if retrieved != c {
			t.Fatal("expected retrieved gin.Context to be identical to current context")
		}
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !called {
		t.Fatal("expected route handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

func TestGinContextToContextMiddleware_ChainsNext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var sequence bytes.Buffer

	router.Use(func(c *gin.Context) {
		sequence.WriteString("A")
		c.Next()
		sequence.WriteString("C")
	})

	router.Use(GinContextToContextMiddleware())

	router.GET("/seq", func(c *gin.Context) {
		sequence.WriteString("B")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/seq", nil)
	router.ServeHTTP(w, req)

	expected := "ABC"
	if sequence.String() != expected {
		t.Fatalf("expected middleware sequence %s, got %s", expected, sequence.String())
	}
}
