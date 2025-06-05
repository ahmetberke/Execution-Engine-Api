// handlers/delete_container.go
package handlers

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gorilla/mux"
)

func DeleteContainerHandler(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	containerName := "user_container_" + userID

	cmd := exec.Command("docker", "rm", "-f", containerName)
	if err := cmd.Run(); err != nil {
		http.Error(w, "Failed to delete container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Container %s deleted", containerName)))
}
