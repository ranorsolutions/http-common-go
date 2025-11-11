package response

import (
	"encoding/json"
	"net/http"
)

type (
	// Response -- HTTP JSON Response Schema
	Response struct {
		Status  int         `json:"status"`
		Message string      `json:"message"`
		Content interface{} `json:"content"`
	}

	// PaginatedResponse -- Paginated JSON Response Schema
	PaginatedResponse struct {
		Count    int         `json:"count"`
		Next     string      `json:"next"`
		Previous string      `json:"previous"`
		Results  interface{} `json:"results"`
	}

	// ResponseLogger -- Logging Wrapped HTTP Writer
	ResponseLogger struct {
		http.ResponseWriter
		statusCode int
	}
)

// NewResponse -- The Response Struct Factory Function
func NewResponse(status int, message string, content interface{}) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Content: content,
	}
}

// NewPaginatedResponse -- The Response Struct Factory Function
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

// ResponseWriter -- Writes results to the http.ResponseWriter object
func ResponseWriter(res http.ResponseWriter, statusCode int, message string, data interface{}) error {
	// Set the HTTP Status Code
	res.WriteHeader(statusCode)

	// Set the CORS Headers
	res.Header().Set("Access-Control-Allow-Origin", "*")

	// Apply the Model factory funciton to set the results
	results := NewResponse(statusCode, message, data)

	// Encode the results into the response
	err := json.NewEncoder(res).Encode(results)

	// Catch any errors
	return err
}

// NewWriter -- Applies the logger to the response writer
func NewWriter(w http.ResponseWriter) *ResponseLogger {
	return &ResponseLogger{w, http.StatusOK}
}

// WriteHeader -- Method to set the HTTP status to the header and logger
func (w *ResponseLogger) WriteHeader(code int) {
	// Apply the HTTP status to the wrapped writer
	w.statusCode = code

	// Apply the HTTP status to the response
	w.ResponseWriter.WriteHeader(code)
}
