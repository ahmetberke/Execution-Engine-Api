package handlers

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"execution-engine-api/internal/container"
	"execution-engine-api/internal/logger"
	"execution-engine-api/pkg/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func executeCommandOnce(userID, command string, conn *websocket.Conn) {
	defer conn.Close()

	containerName := "user_container_" + userID
	cmdString := fmt.Sprintf("stdbuf -oL -eL bash -c %q", command)
	cmd := exec.Command("docker", "exec", containerName, "bash", "-c", cmdString)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	if err := cmd.Start(); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	// stdout ve stderr ayrı goroutine'lerde yazdırılıyor
	go streamToWebSocket(conn, stdout, "")
	go streamToWebSocket(conn, stderr, "Error: ")

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

func streamToWebSocket(conn *websocket.Conn, reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		conn.WriteMessage(websocket.TextMessage, []byte(prefix+line))
	}
}
