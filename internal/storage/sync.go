package storage

import (
	"execution-engine-api/internal/db"
	"execution-engine-api/pkg/models"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// storage/sync.go
func SyncFileMetaFromTmp(userID string, initPath string) error {
	tmpDir := fmt.Sprintf("tmp/%s", userID)

	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(path, tmpDir+"/")

		// workspace/user_<id>/ klasörünü çıkar
		prefixToTrim := fmt.Sprintf("workspace/%s/", userID)
		relPath = strings.TrimPrefix(relPath, prefixToTrim)

		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = "" // kök klasördeyse boş string ver
		}

		if initPath != "" {
			dir = initPath + dir
		}

		name := filepath.Base(relPath)
		mimeType := mime.TypeByExtension(filepath.Ext(name))

		log.Println("Uploading to Mongodb:", relPath, "Dir:", dir, "user:", "user_"+userID)

		meta := models.FileMeta{
			UserID:    "user_" + userID,
			Name:      name,
			Path:      dir,
			Type:      detectFileType(name),
			MimeType:  mimeType,
			CreatedAt: time.Now(),
		}

		// Mongo'ya insert veya upsert
		db.UpsertFileMeta(meta)
		return nil
	})

	return err
}

func detectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return "image"
	case ".pdf":
		return "pdf"
	case ".txt":
		return "text"
	case ".zip", ".rar":
		return "archive"
	case ".go", ".js", ".py":
		return "code"
	default:
		return "file"
	}
}
