package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"execution-engine-api/internal/handlers"
	"execution-engine-api/internal/logger"
)

func main() {
	// Log sistemini başlat
	logger.InitLogger()
	logger.Log.Info("Starting Execution Engine API Server...")

	r := mux.NewRouter()
	r.HandleFunc("/execute", handlers.CommandHandler).Methods("POST")
	r.HandleFunc("/ws", handlers.WSHandler)

	logger.Log.Info("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
