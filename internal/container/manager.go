package container

import (
	"bytes"
	"execution-engine-api/internal/aws"
	"execution-engine-api/internal/logger"
	"fmt"
	"os/exec"
	"strings"
)

// Kullanıcının dosyalarını alıp konteyner içine kopyalar
func SyncFilesToContainer(userID string) error {
	tmpDir := fmt.Sprintf("tmp/%s", userID)
	containerName := "user_container_" + userID

	// AWS'den dosyaları çek
	err := aws.SyncUserFiles(userID)
	if err != nil {
		return fmt.Errorf("failed to sync files from AWS: %w", err)
	}

	// Host makinede dosyaların gerçekten indirildiğini kontrol et
	fmt.Println("Checking files on host before copying:")
	cmd := exec.Command("ls", "-lah", tmpDir)
	output, _ := cmd.Output()
	fmt.Println(string(output)) // Konsola yazdır

	// Dosyaları konteyner içine kopyala
	cmd = exec.Command("docker", "cp", tmpDir, containerName+":/workspace")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy files to container: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("Files copied to container %s", containerName))
	return nil
}

func containerExists(containerName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	containers := strings.Split(out.String(), "\n")
	for _, name := range containers {
		if name == containerName {
			return true
		}
	}
	return false
}

func containerRunning(containerName string) bool {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	containers := strings.Split(out.String(), "\n")
	for _, name := range containers {
		if name == containerName {
			return true
		}
	}
	return false
}

func EnsureContainer(userID string) error {
	containerName := "user_container_" + userID

	if containerRunning(containerName) {
		return nil
	}

	if containerExists(containerName) {
		cmd := exec.Command("docker", "start", containerName)
		return cmd.Run()
	}

	cmd := exec.Command("docker", "run", "-dit", "--name", containerName, "custom-ubuntu-python", "bash")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %s, stderr: %s", err.Error(), stderr.String())
	}

	// Kullanıcı dosyalarını konteynere senkronize et
	err := SyncFilesToContainer(userID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Could not sync files for user %s: %s", userID, err.Error()))
	}

	return nil
}

func CreateContainerWithPath(userID, rootDir string) error {
	containerName := "user_container_" + userID

	if containerRunning(containerName) {
		return nil
	}

	if containerExists(containerName) {
		cmd := exec.Command("docker", "start", containerName)
		return cmd.Run()
	}

	cmd := exec.Command("docker", "run", "-dit", "--name", containerName, "custom-ubuntu-python", "bash")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %s, stderr: %s", err.Error(), stderr.String())
	}

	// Belirli klasörü senkronize et
	err := SyncSpecificPath(userID, rootDir)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Could not sync %s for user %s: %s", rootDir, userID, err.Error()))
	}

	return nil
}

func SyncSpecificPath(userID, rootDir string) error {
	localPath := fmt.Sprintf("tmp/%s", userID)
	containerName := "user_container_" + userID

	err := aws.SyncUserSubPath(userID, rootDir) // => s3://bucket/userId/rootDir → tmp/userId/
	if err != nil {
		return fmt.Errorf("failed to sync path from AWS: %w", err)
	}

	cmd := exec.Command("docker", "cp", localPath, containerName+":/workspace")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy files to container: %w", err)
	}

	return nil
}
