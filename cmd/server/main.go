package main

import (
	"log"
	"net/http"

	"execution-engine-api/internal/aws"
	"execution-engine-api/internal/db"
	"execution-engine-api/internal/handlers"
	"execution-engine-api/internal/logger"
	auth "execution-engine-api/internal/middlewares"
	"execution-engine-api/internal/redis"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	godotenv.Load()

	logger.InitLogger()
	db.InitMongo()
	aws.LoadAWSCredentials()

	redis.InitRedis()

	r := mux.NewRouter()
	r.Handle("/execute", auth.JWTMiddleware(http.HandlerFunc(handlers.CommandHandler))).Methods("POST")
	r.HandleFunc("/ws", handlers.WSHandler)
	r.HandleFunc("/ws/exec", handlers.WSHandlerExec)
	r.Handle("/container/init", auth.JWTMiddleware(http.HandlerFunc(handlers.InitContainerHandler))).Methods("POST")
	r.Handle("/container/status", auth.JWTMiddleware(http.HandlerFunc(handlers.GetContainerStatusHandler))).Methods("GET")
	r.Handle("/container", auth.JWTMiddleware(http.HandlerFunc(handlers.DeleteContainerHandler))).Methods("DELETE")

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // geliştirici için açık; prod'da domain bazlı kısıtlanmalı
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}
