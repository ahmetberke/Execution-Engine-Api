package handlers

import (
	"encoding/json"
	"net/http"

	"execution-engine-api/internal/container"
	"execution-engine-api/internal/db"
	auth "execution-engine-api/internal/middlewares"
)

func GetContainerStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	// 1. Mongo'dan kaydı al
	record, err := db.FindContainerByUserID(userID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if record == nil {
		http.Error(w, "No container found for user", http.StatusNotFound)
		return
	}

	// 2. Docker durumu kontrol et
	actualRunning := container.IsContainerTrulyRunning(userID)

	// 3. Mongo ile Docker çelişirse -> Mongo'yu güncelle
	if actualRunning && record.Status != "running" {
		_ = db.UpdateContainerStatus(userID, "running")
		record.Status = "running"
	} else if !actualRunning && record.Status == "running" {
		_ = db.UpdateContainerStatus(userID, "stopped")
		record.Status = "stopped"
	}

	// 4. Cevabı döndür
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}
