package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		message        string
		wantStatusCode int
		wantSuccess    bool
		wantError      string
	}{
		{
			name:           "bad request error",
			statusCode:     http.StatusBadRequest,
			message:        "Invalid input",
			wantStatusCode: http.StatusBadRequest,
			wantSuccess:    false,
			wantError:      "Invalid input",
		},
		{
			name:           "unauthorized error",
			statusCode:     http.StatusUnauthorized,
			message:        "Unauthorized access",
			wantStatusCode: http.StatusUnauthorized,
			wantSuccess:    false,
			wantError:      "Unauthorized access",
		},
		{
			name:           "internal server error",
			statusCode:     http.StatusInternalServerError,
			message:        "Something went wrong",
			wantStatusCode: http.StatusInternalServerError,
			wantSuccess:    false,
			wantError:      "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondWithError(w, tt.statusCode, tt.message)

			if w.Code != tt.wantStatusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatusCode)
			}

			var response ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success != tt.wantSuccess {
				t.Errorf("success = %v, want %v", response.Success, tt.wantSuccess)
			}

			if response.Error != tt.wantError {
				t.Errorf("error = %q, want %q", response.Error, tt.wantError)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}
		})
	}
}

func TestRespondWithSuccess(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		message        string
		data           interface{}
		wantStatusCode int
		wantSuccess    bool
		wantMessage    string
	}{
		{
			name:           "success with data",
			statusCode:     http.StatusOK,
			message:        "Operation successful",
			data:           map[string]string{"id": "123"},
			wantStatusCode: http.StatusOK,
			wantSuccess:    true,
			wantMessage:    "Operation successful",
		},
		{
			name:           "success without data",
			statusCode:     http.StatusOK,
			message:        "Deleted successfully",
			data:           nil,
			wantStatusCode: http.StatusOK,
			wantSuccess:    true,
			wantMessage:    "Deleted successfully",
		},
		{
			name:           "created status",
			statusCode:     http.StatusCreated,
			message:        "Resource created",
			data:           map[string]int{"id": 456},
			wantStatusCode: http.StatusCreated,
			wantSuccess:    true,
			wantMessage:    "Resource created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondWithSuccess(w, tt.statusCode, tt.message, tt.data)

			if w.Code != tt.wantStatusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatusCode)
			}

			var response SuccessResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success != tt.wantSuccess {
				t.Errorf("success = %v, want %v", response.Success, tt.wantSuccess)
			}

			if response.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", response.Message, tt.wantMessage)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}
		})
	}
}

func TestRespondWithValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithValidationError(w, "Email is required")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Success != false {
		t.Errorf("success = %v, want false", response.Success)
	}

	if response.Error != "Email is required" {
		t.Errorf("error = %q, want %q", response.Error, "Email is required")
	}
}

func TestRespondWithNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithNotFound(w, "User")

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Error != "User not found" {
		t.Errorf("error = %q, want %q", response.Error, "User not found")
	}
}

func TestRespondWithUnauthorized(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "with custom message",
			message: "Invalid token",
			want:    "Invalid token",
		},
		{
			name:    "with empty message",
			message: "",
			want:    "Unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondWithUnauthorized(w, tt.message)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
			}

			var response ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Error != tt.want {
				t.Errorf("error = %q, want %q", response.Error, tt.want)
			}
		})
	}
}

func TestRespondWithConflict(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithConflict(w, "Email already exists")

	if w.Code != http.StatusConflict {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusConflict)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Error != "Email already exists" {
		t.Errorf("error = %q, want %q", response.Error, "Email already exists")
	}
}

func TestRespondWithMethodNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithMethodNotAllowed(w, "POST")

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	allowHeader := w.Header().Get("Allow")
	if allowHeader != "POST" {
		t.Errorf("Allow header = %q, want %q", allowHeader, "POST")
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	expectedError := "Method not allowed. Use POST"
	if response.Error != expectedError {
		t.Errorf("error = %q, want %q", response.Error, expectedError)
	}
}

func TestRespondWithInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := errors.New("database connection failed")

	RespondWithInternalError(w, testErr, "Test operation")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	expectedError := "An unexpected error occurred. Please try again later."
	if response.Error != expectedError {
		t.Errorf("error = %q, want %q", response.Error, expectedError)
	}

	// Note: We can't easily test that the error was logged, but we verify the response is generic
}
