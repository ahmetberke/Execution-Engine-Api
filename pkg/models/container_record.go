package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContainerRecord struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	ContainerName string             `bson:"container_name" json:"container_name"`
	Path          string             `bson:"path" json:"path"`     // Örn: "projects/ocr-demo"
	Status        string             `bson:"status" json:"status"` // created | running | stopped | deleted
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}
