package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func TestSetTaskPriorityHandler(t *testing.T) {
	h := &TaskcafeHandler{}

	tests := []struct {
		name       string
		taskID     string
		priority   string
		wantStatus int
	}{
		{
			name:       "valid high priority",
			taskID:     uuid.New().String(),
			priority:   "high",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid medium priority",
			taskID:     uuid.New().String(),
			priority:   "medium",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid low priority",
			taskID:     uuid.New().String(),
			priority:   "low",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid priority",
			taskID:     uuid.New().String(),
			priority:   "urgent",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid task ID",
			taskID:     "invalid-uuid",
			priority:   "high",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqData := SetTaskPriorityRequest{Priority: tt.priority}
			body, _ := json.Marshal(reqData)

			router := chi.NewRouter()
			router.Put("/tasks/{taskID}/priority", h.SetTaskPriorityHandler)

			req := httptest.NewRequest("PUT", "/tasks/"+tt.taskID+"/priority", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SetTaskPriorityHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response SetTaskPriorityResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.Priority != tt.priority {
					t.Errorf("Response priority = %v, want %v", response.Priority, tt.priority)
				}
				if response.TaskID != tt.taskID {
					t.Errorf("Response task ID = %v, want %v", response.TaskID, tt.taskID)
				}
			}
		})
	}
}

func TestGetTaskPriorityHandler(t *testing.T) {
	h := &TaskcafeHandler{}

	tests := []struct {
		name       string
		taskID     string
		wantStatus int
	}{
		{
			name:       "valid task ID",
			taskID:     uuid.New().String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid task ID",
			taskID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/tasks/{taskID}/priority", h.GetTaskPriorityHandler)

			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID+"/priority", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetTaskPriorityHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response GetTaskPriorityResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.TaskID != tt.taskID {
					t.Errorf("Response task ID = %v, want %v", response.TaskID, tt.taskID)
				}
				if response.Priority == "" {
					t.Errorf("Expected priority in response")
				}
			}
		})
	}
}

func TestPriorityValidation(t *testing.T) {
	validPriorities := map[string]bool{
		"high":   true,
		"medium": true,
		"low":    true,
	}

	tests := []struct {
		priority string
		valid    bool
	}{
		{"high", true},
		{"medium", true},
		{"low", true},
		{"urgent", false},
		{"critical", false},
		{"", false},
		{"HIGH", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			isValid := validPriorities[tt.priority]
			if isValid != tt.valid {
				t.Errorf("Priority validation for %q = %v, want %v", tt.priority, isValid, tt.valid)
			}
		})
	}
}
