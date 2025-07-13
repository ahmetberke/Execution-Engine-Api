package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileMeta struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	Path      string             `bson:"path" json:"path"`
	Type      string             `bson:"type" json:"type"`
	MimeType  string             `bson:"mimetype" json:"mimetype"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
