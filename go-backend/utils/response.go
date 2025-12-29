package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondWithError sends a JSON error response with appropriate status code
func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// RespondWithSuccess sends a JSON success response
func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := SuccessResponse{
		Success: true,
		Message: message,
	}
	if data != nil {
		response.Data = data
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(response)
}

// RespondWithJSON sends a JSON response with custom structure
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(payload)
}

// RespondWithInternalError logs the error and sends a generic error response
func RespondWithInternalError(w http.ResponseWriter, err error, context string) {
	log.Printf("%s error: %v", context, err)
	RespondWithError(w, http.StatusInternalServerError,
		"An unexpected error occurred. Please try again later.")
}

// RespondWithValidationError sends a 400 Bad Request with validation message
func RespondWithValidationError(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusBadRequest, message)
}

// RespondWithNotFound sends a 404 Not Found response
func RespondWithNotFound(w http.ResponseWriter, resource string) {
	RespondWithError(w, http.StatusNotFound, resource+" not found")
}

// RespondWithUnauthorized sends a 401 Unauthorized response
func RespondWithUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	RespondWithError(w, http.StatusUnauthorized, message)
}

// RespondWithForbidden sends a 403 Forbidden response
func RespondWithForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	RespondWithError(w, http.StatusForbidden, message)
}

// RespondWithConflict sends a 409 Conflict response
func RespondWithConflict(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusConflict, message)
}

// RespondWithMethodNotAllowed sends a 405 Method Not Allowed response
func RespondWithMethodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed. Use "+allowed)
}
