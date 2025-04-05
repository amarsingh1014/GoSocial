package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(addr, pw string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil
	}

	return client
}