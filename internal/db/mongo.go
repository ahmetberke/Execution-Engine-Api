package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var ContainerCollection *mongo.Collection

func InitMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set in environment")
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	MongoClient = client
	ContainerCollection = client.Database("execution").Collection("containers")

	log.Println("Connected to MongoDB")
}

func GetCollection(name string) *mongo.Collection {
	return MongoClient.Database("execution").Collection(name)
}

// FileMetaCollection returns the filemeta collection under storage DB
func FileMetaCollection() *mongo.Collection {
	return MongoClient.Database("storage").Collection("filemeta")
}
