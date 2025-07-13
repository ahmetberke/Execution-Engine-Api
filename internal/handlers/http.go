package handlers

import (
	"encoding/json"
	"execution-engine-api/internal/container"
	"execution-engine-api/internal/executor"
	"execution-engine-api/internal/logger"
	auth "execution-engine-api/internal/middlewares"
	"execution-engine-api/internal/redis"
	"execution-engine-api/pkg/models"
	"net/http"
)

func CommandHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(auth.UserIDKey).(string)

	var req models.CommandRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.Log.Warn("Invalid request payload")
		return
	}

	err = container.EnsureContainer(userID)
	if err != nil {
		http.Error(w, "Error ensuring container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := executor.ExecuteCommandInContainer(userID, req.Command)
	if err != nil {
		response := models.CommandResponse{UserID: userID, Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	} else {
		redis.SetContainerTTL(userID) // ✅ TTL resetle
	}

	response := models.CommandResponse{UserID: userID, Output: output}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
