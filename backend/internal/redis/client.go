package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func InitRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     getRedisAddr(),
		Password: "", // no password by default
		DB:       0,  // use default DB
	})

	// Ping to verify connection
	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	fmt.Println("Connected to Redis")
}

func getRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}
