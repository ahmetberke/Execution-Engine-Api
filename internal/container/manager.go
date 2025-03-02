package container

import (
	"bytes"
	"os/exec"
	"strings"
)

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
	return cmd.Run()
}
