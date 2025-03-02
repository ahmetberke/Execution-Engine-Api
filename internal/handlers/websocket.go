package handlers

import (
	"bufio"
	"net/http"
	"os/exec"

	"github.com/gorilla/websocket"
	"execution-engine-api/internal/container"
	"execution-engine-api/pkg/models"
	"execution-engine-api/internal/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func executeCommandOnce(userID, command string, conn *websocket.Conn) {
	defer conn.Close() // Komut tamamlandığında bağlantıyı kapat

	containerName := "user_container_" + userID
	cmd := exec.Command("docker", "exec", containerName, "bash", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		logger.Log.Warn(err.Error())
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		logger.Log.Warn(err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		logger.Log.Warn(err.Error())
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		message := scanner.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			logger.Log.Warn(err.Error())
			break
		}
	}

	scannerErr := bufio.NewScanner(stderr)
	for scannerErr.Scan() {
		message := scannerErr.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Error: "+message)); err != nil {
			logger.Log.Warn(message)
			break
		}
	}

	cmd.Wait()
	conn.WriteMessage(websocket.TextMessage, []byte("ExecutionFinished"))
}

// WSHandler: Sadece tek bir komut çalıştırıp WebSocket'i kapatan yeni handler
func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Warn("WebSocket upgrade failed: " + err.Error())
		return
	}

	var req models.CommandRequest
	err = conn.ReadJSON(&req)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error reading request: "+err.Error()))
		return
	}

	err = container.EnsureContainer(req.UserID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error ensuring container: "+err.Error()))
		return
	}

	// Her komut için ayrı WebSocket bağlantısı açıp işleyelim
	go executeCommandOnce(req.UserID, req.Command, conn)
}
