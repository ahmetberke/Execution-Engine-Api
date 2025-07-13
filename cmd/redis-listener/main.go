package main

import (
	"execution-engine-api/internal/db"
	"execution-engine-api/internal/logger"
	"execution-engine-api/internal/redis"

	"github.com/joho/godotenv"
)

func main() {
	// Logger, MongoDB ve Redis bağlantılarını başlat
	godotenv.Load()
	logger.InitLogger()
	db.InitMongo()
	redis.InitRedis()

	// Redis key expiration listener'ı başlat
	redis.StartKeyExpirationListener()
}
