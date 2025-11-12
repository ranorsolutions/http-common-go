package response

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestNewResponse(t *testing.T) {
	r := NewResponse(200, "ok", map[string]string{"a": "b"})
	if r.Status != 200 || r.Message != "ok" {
		t.Fatal("invalid response fields")
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	r := NewPaginatedResponse(200, 2, "ok", "next", "prev", []int{1, 2})
	p, ok := r.Content.(*PaginatedResponse)
	if !ok {
		t.Fatal("expected Content to be *PaginatedResponse")
	}
	if p.Count != 2 || p.Next != "next" {
		t.Fatal("invalid pagination fields")
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	err := WriteJSON(rec, 201, "created", map[string]string{"id": "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp Response
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if resp.Message != "created" {
		t.Errorf("expected message 'created', got %q", resp.Message)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type: application/json")
	}
}
