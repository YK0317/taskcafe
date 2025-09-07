package route

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/jordanknott/taskcafe/internal/db"
	// "github.com/jordanknott/taskcafe/internal/frontend" // Commented out for simple build
	"github.com/jordanknott/taskcafe/internal/utils"
)

// Frontend serves the index.html file (simplified for development)
func (h *TaskcafeHandler) Frontend(w http.ResponseWriter, r *http.Request) {
	// Simple file serving from filesystem
	indexPath := "./frontend/build/index.html"
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		http.Error(w, "Frontend not built. Run: cd frontend && npm run build", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, indexPath)
}

// ProfileImageUpload handles a user uploading a new avatar profile image
func (h *TaskcafeHandler) ProfileImageUpload(w http.ResponseWriter, r *http.Request) {
	log.Info("preparing to upload file")
	userID, ok := r.Context().Value(utils.UserIDKey).(uuid.UUID)
	if !ok {
		log.Error("not a valid uuid")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.WithError(err).Error("issue while uploading file")
		return
	}
	defer file.Close()
	filename := strings.ReplaceAll(handler.Filename, " ", "-")
	encodedFilename := url.QueryEscape(filename)
	log.WithFields(log.Fields{"filename": encodedFilename, "size": handler.Size, "header": handler.Header}).Info("file metadata")

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithError(err).Error("while reading file")
		return
	}
	err = ioutil.WriteFile("uploads/"+filename, fileBytes, 0644)
	if err != nil {
		log.WithError(err).Error("while reading file")
		return
	}

	h.repo.UpdateUserAccountProfileAvatarURL(r.Context(), db.UpdateUserAccountProfileAvatarURLParams{UserID: userID, ProfileAvatarUrl: sql.NullString{String: "/uploads/" + encodedFilename, Valid: true}})
	// return that we have successfully uploaded our file!
	log.Info("file uploaded")
	json.NewEncoder(w).Encode(AvatarUploadResponseData{URL: "/uploads/" + encodedFilename, UserID: userID.String()})

}
