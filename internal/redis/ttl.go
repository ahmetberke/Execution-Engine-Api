package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		panic("REDIS_ADDR is not set in environment")
	}

	Client = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func SetContainerTTL(userID string) error {
	key := fmt.Sprintf("container:%s", userID)
	return Client.SetEx(context.Background(), key, "active", 10*time.Minute).Err()
}
