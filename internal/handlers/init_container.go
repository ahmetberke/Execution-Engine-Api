package handlers

import (
	"encoding/json"
	"net/http"

	"execution-engine-api/internal/container"
	"execution-engine-api/pkg/models"
)

func InitContainerHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ContainerInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	err := container.CreateContainerWithPath(req.UserID, req.RootDir)
	if err != nil {
		http.Error(w, "Failed to create container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Container created and files synced"))
}
