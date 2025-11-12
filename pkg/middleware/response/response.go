// Package response provides simple helpers for writing standardized JSON HTTP responses.
package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response defines the standard JSON response envelope.
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Content interface{} `json:"content"`
}

// ResponseLogger wraps gin.ResponseWriter to capture status codes
// while preserving full compatibility with Gin's writer interface.
type ResponseLogger struct {
	gin.ResponseWriter
	statusCode int
}

// NewWriter wraps a gin.ResponseWriter for logging and status tracking.
func NewWriter(w gin.ResponseWriter) *ResponseLogger {
	return &ResponseLogger{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader records and forwards the status code.
func (w *ResponseLogger) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Status returns the most recently written HTTP status code.
func (w *ResponseLogger) Status() int {
	return w.statusCode
}

// PaginatedResponse defines the schema for paginated API results.
type PaginatedResponse struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Results  interface{} `json:"results"`
}

// NewResponse creates a simple JSON response envelope.
func NewResponse(status int, message string, content interface{}) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Content: content,
	}
}

// NewPaginatedResponse creates a paginated JSON response envelope.
func NewPaginatedResponse(status, count int, message, next, prev string, results interface{}) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Content: &PaginatedResponse{
			Count:    count,
			Next:     next,
			Previous: prev,
			Results:  results,
		},
	}
}

// WriteJSON writes a JSON response to the http.ResponseWriter.
func WriteJSON(w http.ResponseWriter, statusCode int, message string, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(NewResponse(statusCode, message, data))
}
