package handlers

import (
	"encoding/json"
	"net/http"
	"execution-engine-api/internal/container"
	"execution-engine-api/internal/executor"
	"execution-engine-api/pkg/models"
	"execution-engine-api/internal/logger"

)

func CommandHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CommandRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.Log.Warn("Invalid request payload")
		return
	}

	err = container.EnsureContainer(req.UserID)
	if err != nil {
		http.Error(w, "Error ensuring container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := executor.ExecuteCommandInContainer(req.UserID, req.Command)
	if err != nil {
		response := models.CommandResponse{UserID: req.UserID, Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.CommandResponse{UserID: req.UserID, Output: output}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
