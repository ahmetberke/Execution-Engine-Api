package main

import (
	"log"
	"net/http"

	"execution-engine-api/internal/aws"
	"execution-engine-api/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {

	aws.LoadAWSCredentials()

	r := mux.NewRouter()
	r.HandleFunc("/execute", handlers.CommandHandler).Methods("POST")
	r.HandleFunc("/ws", handlers.WSHandler)
	r.HandleFunc("/container/init", handlers.InitContainerHandler).Methods("POST")
	r.HandleFunc("/container/{userID}", handlers.DeleteContainerHandler).Methods("DELETE")

	log.Println("Server started")

	log.Fatal(http.ListenAndServe(":8080", r))
}
