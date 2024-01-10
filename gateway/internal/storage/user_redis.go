package storage

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type UserRedis struct {
	client *redis.Client
	log    *slog.Logger
}

func NewUserRedis(client *redis.Client, logger *slog.Logger) *UserRedis {
	return &UserRedis{client: client, log: logger}
}

func (r *UserRedis) GetUserId(tokenId string) (string, error) {
	var ctx = context.Background()

	id, err := r.client.Get(ctx, tokenId).Result()
	r.log.Info(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *UserRedis) SetUserId(tokenId string, id int, duration time.Duration) error {
	var ctx = context.Background()

	err := r.client.Set(ctx, tokenId, id, duration).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *UserRedis) DeleteUser(tokenId string) error {
	var ctx = context.Background()

	res, err := r.client.Del(ctx, tokenId).Result()
	println("Number of keys deleted:", res)
	if err != nil {
		return err
	}

	return nil
}
