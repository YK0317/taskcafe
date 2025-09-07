package route

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type priorityResource struct{}

// SetTaskPriorityRequest is the request data for setting task priority
type SetTaskPriorityRequest struct {
	Priority string `json:"priority"`
}

// SetTaskPriorityResponse is the response data for setting task priority
type SetTaskPriorityResponse struct {
	TaskID   string `json:"task_id"`
	Priority string `json:"priority"`
	Message  string `json:"message"`
}

// GetTaskPriorityResponse is the response data for getting task priority
type GetTaskPriorityResponse struct {
	TaskID   string `json:"task_id"`
	Priority string `json:"priority"`
}

// SetTaskPriorityHandler sets the priority for a task
func (h *TaskcafeHandler) SetTaskPriorityHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		log.WithError(err).Error("invalid task ID")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid task ID",
		})
		return
	}

	var requestData SetTaskPriorityRequest
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.WithError(err).Error("issue decoding set priority request")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format",
		})
		return
	}

	// Validate priority
	validPriorities := map[string]bool{
		"high":   true,
		"medium": true,
		"low":    true,
	}

	if !validPriorities[requestData.Priority] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid priority. Use: high, medium, low",
		})
		return
	}

	// TODO: In real implementation, you would update the database
	// For now, we'll just simulate success
	log.WithFields(log.Fields{
		"taskID":   taskID,
		"priority": requestData.Priority,
	}).Info("task priority updated")

	response := SetTaskPriorityResponse{
		TaskID:   taskIDStr,
		Priority: requestData.Priority,
		Message:  "Priority updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTaskPriorityHandler gets the priority for a task
func (h *TaskcafeHandler) GetTaskPriorityHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		log.WithError(err).Error("invalid task ID")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid task ID",
		})
		return
	}

	// TODO: In real implementation, you would query the database
	// For demo, return mock data
	response := GetTaskPriorityResponse{
		TaskID:   taskIDStr,
		Priority: "medium", // Mock default priority
	}

	log.WithFields(log.Fields{
		"taskID": taskID,
	}).Info("task priority retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Routes registers all priority routes
func (rs priorityResource) Routes(taskcafeHandler TaskcafeHandler) chi.Router {
	r := chi.NewRouter()
	r.Put("/{taskID}/priority", taskcafeHandler.SetTaskPriorityHandler)
	r.Get("/{taskID}/priority", taskcafeHandler.GetTaskPriorityHandler)
	return r
}
