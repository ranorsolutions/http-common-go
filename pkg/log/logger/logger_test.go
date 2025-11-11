package logger

import (
	"golang.org/x/exp/slices"
	"net/http"
	"testing"
)

func TestFormat(t *testing.T) {
	// msg := "Test Message"
	tt := []struct {
		name   string
		url    string
		method string
		status int
		err    string
	}{
		{name: "GET request", url: "localhost:8080/endpoint?param=value", method: "GET"},
		{name: "POST request", url: "localhost:8080/endpoint", method: "POST"},
	}

	defaultKeys := []string{"service", "version"}
	requestKeys := []string{"agent", "method", "origin", "resource", "size"}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := New("test", "1.0.0", true)

			// Check that the formatter has not already been added
			for key := range logger.Entry.Data {
				if !slices.Contains(defaultKeys, key) {
					t.Fatalf("missing default key %s from entry", key)
				}

				if slices.Contains(requestKeys, key) {
					t.Fatalf("request key %s already added to entry", key)
				}
			}

			// Create a request to be formatted
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			// Format the logger with the request object
			logger.Format(req)

			// Test that the entry has been updated
			for key := range logger.Entry.Data {
				if !slices.Contains(append(defaultKeys, requestKeys...), key) {
					t.Fatalf("request key %s not found on entry", key)
				}
			}
		})
	}
}
