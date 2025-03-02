package executor

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ExecuteCommandInContainer(userID, command string) (string, error) {
	containerName := "user_container_" + userID
	cmd := exec.Command("docker", "exec", containerName, "bash", "-c", command)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf(stderr.String())
	}
	return out.String(), nil
}
