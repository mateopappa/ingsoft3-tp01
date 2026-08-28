package http

import (
	"encoding/json"
	"net/http"
)

// JSON sends a JSON response with status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// Error sends a standardized JSON error response.
func Error(w http.ResponseWriter, status int, message, code string) {
	JSON(w, status, map[string]string{
		"error": message,
		"code":  code,
	})
}
