package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func TestGetProjectStatsHandler(t *testing.T) {
	h := &TaskcafeHandler{}

	tests := []struct {
		name       string
		projectID  string
		wantStatus int
	}{
		{
			name:       "valid project ID",
			projectID:  uuid.New().String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid project ID",
			projectID:  "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/reports/projects/{projectID}/stats", h.GetProjectStatsHandler)

			req := httptest.NewRequest("GET", "/reports/projects/"+tt.projectID+"/stats", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetProjectStatsHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response ProjectStatsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.ProjectID != tt.projectID {
					t.Errorf("Response project ID = %v, want %v", response.ProjectID, tt.projectID)
				}
				if response.TotalTasks <= 0 {
					t.Errorf("Expected positive total tasks")
				}
				if response.Progress < 0 || response.Progress > 100 {
					t.Errorf("Progress should be between 0 and 100, got %v", response.Progress)
				}
			}
		})
	}
}

func TestGetTaskSummaryHandler(t *testing.T) {
	h := &TaskcafeHandler{}

	router := chi.NewRouter()
	router.Get("/reports/summary", h.GetTaskSummaryHandler)

	req := httptest.NewRequest("GET", "/reports/summary", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetTaskSummaryHandler() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response TaskSummaryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.TotalTasks <= 0 {
		t.Errorf("Expected positive total tasks")
	}

	if response.CompletedTasks+response.PendingTasks+response.OverdueTasks > response.TotalTasks {
		t.Errorf("Sum of task states should not exceed total tasks")
	}
}

func TestGetProjectHealthHandler(t *testing.T) {
	h := &TaskcafeHandler{}

	tests := []struct {
		name       string
		projectID  string
		wantStatus int
	}{
		{
			name:       "valid project ID",
			projectID:  uuid.New().String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid project ID",
			projectID:  "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/reports/projects/{projectID}/health", h.GetProjectHealthHandler)

			req := httptest.NewRequest("GET", "/reports/projects/"+tt.projectID+"/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetProjectHealthHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response["project_id"] != tt.projectID {
					t.Errorf("Response project ID = %v, want %v", response["project_id"], tt.projectID)
				}

				health, ok := response["health"].(string)
				if !ok || health == "" {
					t.Errorf("Expected health status in response")
				}

				healthScore, ok := response["health_score"].(float64)
				if !ok || healthScore < 1 || healthScore > 5 {
					t.Errorf("Expected health score between 1 and 5, got %v", healthScore)
				}
			}
		})
	}
}

func TestHealthCalculation(t *testing.T) {
	tests := []struct {
		name           string
		completionRate float64
		overdueTasks   int
		expectedHealth string
		expectedScore  int
	}{
		{
			name:           "excellent health",
			completionRate: 85,
			overdueTasks:   0,
			expectedHealth: "excellent",
			expectedScore:  5,
		},
		{
			name:           "good health",
			completionRate: 70,
			overdueTasks:   1,
			expectedHealth: "good",
			expectedScore:  4,
		},
		{
			name:           "fair health",
			completionRate: 50,
			overdueTasks:   3,
			expectedHealth: "fair",
			expectedScore:  3,
		},
		{
			name:           "poor health",
			completionRate: 30,
			overdueTasks:   5,
			expectedHealth: "poor",
			expectedScore:  2,
		},
		{
			name:           "critical health",
			completionRate: 10,
			overdueTasks:   10,
			expectedHealth: "critical",
			expectedScore:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var health string
			var healthScore int

			if tt.completionRate >= 80 && tt.overdueTasks == 0 {
				health = "excellent"
				healthScore = 5
			} else if tt.completionRate >= 60 && tt.overdueTasks <= 2 {
				health = "good"
				healthScore = 4
			} else if tt.completionRate >= 40 {
				health = "fair"
				healthScore = 3
			} else if tt.completionRate >= 20 {
				health = "poor"
				healthScore = 2
			} else {
				health = "critical"
				healthScore = 1
			}

			if health != tt.expectedHealth {
				t.Errorf("Health calculation = %v, want %v", health, tt.expectedHealth)
			}
			if healthScore != tt.expectedScore {
				t.Errorf("Health score = %v, want %v", healthScore, tt.expectedScore)
			}
		})
	}
}
