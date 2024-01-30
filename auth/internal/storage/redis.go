package storage

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func initRedisConnection() *redis.Client {
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		Protocol: 3,  // specify 2 for RESP 2 or 3 for RESP 3
	})

	pingResult, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("PING:", pingResult)

	return redisClient
}
