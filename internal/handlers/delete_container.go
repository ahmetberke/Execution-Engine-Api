package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"execution-engine-api/internal/aws"
	"execution-engine-api/internal/container"
	"execution-engine-api/internal/db"
	auth "execution-engine-api/internal/middlewares"
	"execution-engine-api/internal/storage"
)

func DeleteContainerHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	containerName := "user_container_" + userID

	record, err := db.FindContainerByUserID(userID)
	if record == nil || err != nil {
		http.Error(w, "Failed to fetching container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 1. Container içeriğini tmp klasörüne çıkar
	if err := container.ExtractFilesFromContainer(containerName, userID); err != nil {
		http.Error(w, "Failed to extract files from container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. tmp dizininden S3'e yükle
	if err := aws.UploadUserFilesToS3(userID, record.Path); err != nil {
		http.Error(w, "Failed to upload files to S3: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2.5 MongoDB filemeta güncelle
	if err := storage.SyncFileMetaFromTmp(userID, record.Path); err != nil {
		http.Error(w, "Failed to write to database: "+err.Error(), http.StatusInternalServerError)
	}

	// 3. Docker container sil
	cmd := exec.Command("docker", "rm", "-f", containerName)
	if err := cmd.Run(); err != nil {
		http.Error(w, "Failed to delete container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. tmp klasörünü sil
	tmpDir := fmt.Sprintf("tmp/%s", userID)
	if err := os.RemoveAll(tmpDir); err != nil {
		http.Error(w, "Failed to delete temp dir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. MongoDB'de status = "deleted"
	if err := db.UpdateContainerStatus(userID, "deleted"); err != nil {
		http.Error(w, "Failed to update DB status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Container %s deleted and files synced back to S3", containerName)))
}
