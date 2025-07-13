package handlers

import (
	"encoding/json"
	"net/http"

	"execution-engine-api/internal/container"
	"execution-engine-api/internal/db"
	auth "execution-engine-api/internal/middlewares"
	"execution-engine-api/internal/redis"
	"execution-engine-api/pkg/models"
)

func InitContainerHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(auth.UserIDKey).(string)

	var req models.ContainerInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// MongoDB'den container kaydını kontrol et
	existing, err := db.FindContainerByUserID(userID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if existing != nil && existing.Status == "running" {
		// ⬇️ Mevcut kayıt varsa ve "running" ise bile path güncellenmeli
		if err := db.UpdateContainer(userID, "running", req.RootDir); err != nil {
			http.Error(w, "Failed to update container path: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Container already running. Path updated."))
		return
	}

	// Docker container oluştur ve senkronize et
	err = container.CreateContainerWithPath(userID, req.RootDir)
	if err != nil {
		http.Error(w, "Failed to create container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TTL başlat
	if err := redis.SetContainerTTL(userID); err != nil {
		http.Error(w, "Container created but failed to set TTL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// MongoDB kaydını oluştur veya güncelle
	if existing == nil {
		newRecord := &models.ContainerRecord{
			UserID:        userID,
			ContainerName: "user_container_" + userID,
			Path:          req.RootDir,
			Status:        "running",
		}
		err = db.InsertContainer(newRecord)
	} else {
		err = db.UpdateContainer(userID, "running", req.RootDir)
	}

	if err != nil {
		http.Error(w, "Failed to persist container state: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Container created, files synced, and state persisted."))
}
