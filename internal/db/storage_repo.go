package db

import (
	"context"
	"execution-engine-api/pkg/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpsertFileMeta(meta models.FileMeta) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": meta.UserID,
		"path":    meta.Path,
		"name":    meta.Name,
	}

	update := bson.M{
		"$set": bson.M{
			"type":       meta.Type,
			"mimetype":   meta.MimeType,
			"created_at": meta.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := FileMetaCollection().UpdateOne(ctx, filter, update, opts)
	return err
}
