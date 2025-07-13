package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"execution-engine-api/internal/db"
	"execution-engine-api/internal/logger"
	"execution-engine-api/internal/redis"
	"execution-engine-api/pkg/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var execUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSHandlerExec(w http.ResponseWriter, r *http.Request) {

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Token eksik", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Geçersiz token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["sub"] == nil {
		http.Error(w, "Token geçerli değil", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Token'dan kullanıcı bilgisi alınamadı", http.StatusUnauthorized)
		return
	}

	conn, err := execUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Warn("WebSocket upgrade failed (exec): " + err.Error())
		return
	}
	defer conn.Close()

	var req models.CommandRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error reading JSON: "+err.Error()))
		return
	}

	// Mongo'dan container bilgisi al
	record, err := db.FindContainerByUserID(userID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Database error: "+err.Error()))
		return
	}
	if record == nil || record.Status != "running" {
		conn.WriteMessage(websocket.TextMessage, []byte("No running container found for this user."))
		return
	}

	containerName := record.ContainerName

	// Komutu stdbuf ile çalıştır (anlık log yayını için)
	cmdString := fmt.Sprintf("stdbuf -oL -eL bash -c %q", req.Command)
	cmd := exec.Command("docker", "exec", containerName, "bash", "-c", cmdString)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error attaching stdout: "+err.Error()))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error attaching stderr: "+err.Error()))
		return
	}

	if err := cmd.Start(); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error starting command: "+err.Error()))
		return
	}

	_ = redis.SetContainerTTL(userID) // TTL'yi yenile

	// stdout/stderr stream
	go streamToWebSocket(conn, stdout, "")
	go streamToWebSocket(conn, stderr, "Error: ")

	cmd.Wait()

	conn.WriteMessage(websocket.TextMessage, []byte("ExecutionFinished"))
}
