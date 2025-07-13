package db

import (
	"context"
	"errors"
	"time"

	"execution-engine-api/pkg/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Kullanıcı ID'sine göre container kaydını getirir
func FindContainerByUserID(userID string) (*models.ContainerRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result models.ContainerRecord
	err := ContainerCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Kayıt yok
		}
		return nil, err
	}

	return &result, nil
}

// Yeni bir container kaydı ekler
func InsertContainer(record *models.ContainerRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := ContainerCollection.InsertOne(ctx, record)
	return err
}

// Status alanını günceller
func UpdateContainerStatus(userID, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"status": newStatus}}

	res, err := ContainerCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("container not found")
	}
	return nil
}

// Status'u "deleted" olarak işaretler
func MarkContainerDeleted(userID string) error {
	return UpdateContainerStatus(userID, "deleted")
}

// Status ve path alanlarını birlikte günceller
func UpdateContainer(userID, newStatus, newPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"status": newStatus,
			"path":   newPath,
		},
	}

	res, err := ContainerCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("container not found")
	}
	return nil
}
