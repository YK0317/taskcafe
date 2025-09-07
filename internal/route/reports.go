package route

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type reportsResource struct{}

// ProjectStatsResponse represents basic project statistics
type ProjectStatsResponse struct {
	ProjectID      string  `json:"project_id"`
	ProjectName    string  `json:"project_name"`
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	Progress       float64 `json:"progress"`
	OverdueTasks   int     `json:"overdue_tasks"`
}

// TaskSummaryResponse represents task summary statistics
type TaskSummaryResponse struct {
	TotalProjects    int `json:"total_projects"`
	TotalTasks       int `json:"total_tasks"`
	CompletedTasks   int `json:"completed_tasks"`
	PendingTasks     int `json:"pending_tasks"`
	OverdueTasks     int `json:"overdue_tasks"`
}

// GetProjectStatsHandler returns basic statistics for a project
func (h *TaskcafeHandler) GetProjectStatsHandler(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		log.WithError(err).Error("invalid project ID")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid project ID",
		})
		return
	}

	// TODO: In real implementation, you would query the database
	// For demo, return mock data
	stats := ProjectStatsResponse{
		ProjectID:      projectIDStr,
		ProjectName:    "Sample Project",
		TotalTasks:     15,
		CompletedTasks: 8,
		Progress:       53.33, // 8/15 * 100
		OverdueTasks:   2,
	}

	log.WithFields(log.Fields{
		"projectID": projectID,
	}).Info("project stats retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetTaskSummaryHandler returns overall task summary statistics
func (h *TaskcafeHandler) GetTaskSummaryHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: In real implementation, you would query the database for actual stats
	// For demo, return mock data
	summary := TaskSummaryResponse{
		TotalProjects:  5,
		TotalTasks:     42,
		CompletedTasks: 28,
		PendingTasks:   12,
		OverdueTasks:   2,
	}

	log.Info("task summary retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetProjectHealthHandler returns a simple health indicator for a project
func (h *TaskcafeHandler) GetProjectHealthHandler(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		log.WithError(err).Error("invalid project ID")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid project ID",
		})
		return
	}

	// Simple health calculation based on completion rate and overdue tasks
	completionRate := 53.33 // Mock data
	overdueTasks := 2       // Mock data

	var health string
	var healthScore int

	if completionRate >= 80 && overdueTasks == 0 {
		health = "excellent"
		healthScore = 5
	} else if completionRate >= 60 && overdueTasks <= 2 {
		health = "good"
		healthScore = 4
	} else if completionRate >= 40 {
		health = "fair"
		healthScore = 3
	} else if completionRate >= 20 {
		health = "poor"
		healthScore = 2
	} else {
		health = "critical"
		healthScore = 1
	}

	response := map[string]interface{}{
		"project_id":      projectIDStr,
		"health":          health,
		"health_score":    healthScore,
		"completion_rate": completionRate,
		"overdue_tasks":   overdueTasks,
	}

	log.WithFields(log.Fields{
		"projectID": projectID,
		"health":    health,
	}).Info("project health calculated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Routes registers all reports routes
func (rs reportsResource) Routes(taskcafeHandler TaskcafeHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/projects/{projectID}/stats", taskcafeHandler.GetProjectStatsHandler)
	r.Get("/projects/{projectID}/health", taskcafeHandler.GetProjectHealthHandler)
	r.Get("/summary", taskcafeHandler.GetTaskSummaryHandler)
	return r
}
