package aws

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadUserFilesToS3 uploads only new or updated files from tmp/userID to S3
func UploadUserFilesToS3(userID string, initPath string) error {
	s3Client, err := getS3Client()
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}

	localDir := fmt.Sprintf("tmp/%s", userID)

	err = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// relPath: tmp/enes/workspace/enes/test2.py → workspace/enes/test2.py
		relPath := path[len(localDir)+1:]

		// Remove workspace/userID prefix if present
		trimmedPath := relPath
		if strings.HasPrefix(relPath, "workspace/"+userID+"/") {
			trimmedPath = strings.TrimPrefix(relPath, "workspace/"+userID+"/")
		}

		// Prepend initPath if exists
		finalPath := trimmedPath
		if initPath != "" {
			finalPath = filepath.Join(initPath, trimmedPath)
		}

		// Final S3 key: user_<userID>/finalPath
		s3Key := fmt.Sprintf("user_%s/%s", userID, finalPath)

		// Compute file hash
		hash, err := computeSHA256(path)
		if err != nil {
			return fmt.Errorf("hash error for %s: %w", path, err)
		}

		// Check existing object metadata
		head, err := s3Client.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(s3Key),
		})

		if err == nil && head.Metadata != nil {
			if s3Hash, ok := head.Metadata["Hash"]; ok && *s3Hash == hash {
				log.Printf("Skipped unchanged: %s", s3Key)
				return nil
			}
		}

		// Upload file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open error: %w", err)
		}
		defer file.Close()

		log.Println("Uploading:", s3Key)

		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(s3Key),
			Body:   file,
			Metadata: map[string]*string{
				"Hash": aws.String(hash),
			},
		})
		if err != nil {
			return fmt.Errorf("upload error: %w", err)
		}

		return nil
	})

	return err
}

// computeSHA256 returns the hex SHA-256 hash of a file
func computeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
