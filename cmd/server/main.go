package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"execution-engine-api/internal/handlers"
	"execution-engine-api/internal/aws"
)

func main() {

	aws.LoadAWSCredentials()

	r := mux.NewRouter()
	r.HandleFunc("/execute", handlers.CommandHandler).Methods("POST")
	r.HandleFunc("/ws", handlers.WSHandler)

	log.Println("Server started")

	log.Fatal(http.ListenAndServe(":8080", r))
}
