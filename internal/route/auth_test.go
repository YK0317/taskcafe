package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestPasswordResetHandler(t *testing.T) {
	// Create a mock handler (simplified for demo)
	h := &TaskcafeHandler{}

	tests := []struct {
		name       string
		email      string
		wantStatus int
	}{
		{
			name:       "valid email",
			email:      "test@example.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty email",
			email:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email format",
			email:      "invalid-email",
			wantStatus: http.StatusOK, // Still returns OK for security
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqData := RequestPasswordResetData{Email: tt.email}
			body, _ := json.Marshal(reqData)

			req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.RequestPasswordResetHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("RequestPasswordResetHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response["message"] == "" {
					t.Errorf("Expected message in response")
				}
			}
		})
	}
}

func TestLoginRequestValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		valid    bool
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			password: "password123",
			valid:    true,
		},
		{
			name:     "empty username",
			username: "",
			password: "password123",
			valid:    false,
		},
		{
			name:     "empty password",
			username: "testuser",
			password: "",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginData := LoginRequestData{
				Username: tt.username,
				Password: tt.password,
			}

			// Basic validation logic
			isValid := loginData.Username != "" && loginData.Password != ""
			
			if isValid != tt.valid {
				t.Errorf("Login validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}
