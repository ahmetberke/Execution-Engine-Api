package redis

import (
	"context"
	"log"
	"strings"

	"execution-engine-api/internal/container"
	"execution-engine-api/internal/db"
)

func StartKeyExpirationListener() {
	pubsub := Client.PSubscribe(context.Background(), "__keyevent@0__:expired")
	defer pubsub.Close()

	log.Println("Redis key expiration listener started...")

	for msg := range pubsub.Channel() {
		expiredKey := msg.Payload
		if !strings.HasPrefix(expiredKey, "container:") {
			continue
		}

		userID := strings.TrimPrefix(expiredKey, "container:")
		log.Printf("Key expired for user: %s\n", userID)

		// Container'ı durdur
		err := container.StopAndRemoveContainer(userID)
		if err != nil {
			log.Printf("Failed to stop container for user %s: %v\n", userID, err)
			continue
		}

		// MongoDB'deki durumu güncelle
		err = db.UpdateContainerStatus(userID, "stopped")
		if err != nil {
			log.Printf("Failed to update MongoDB status for user %s: %v\n", userID, err)
		} else {
			log.Printf("MongoDB status updated to 'stopped' for user %s\n", userID)
		}
	}
}
