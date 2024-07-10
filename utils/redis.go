package utils

import (
	"os"

	"github.com/go-redis/redis"
)

func InitializeRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASS"), // no password set
		DB:       0,
	})
	return rdb
}
