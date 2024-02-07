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

func (r *UserRedis) GetUserId(ctx context.Context, tokenId string) (string, error) {
	id, err := r.client.Get(ctx, tokenId).Result()
	r.log.Info(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *UserRedis) SetUserId(ctx context.Context, tokenId string, id int, duration time.Duration) error {
	err := r.client.Set(ctx, tokenId, id, duration).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *UserRedis) DeleteUser(ctx context.Context, tokenId string) error {
	res, err := r.client.Del(ctx, tokenId).Result()
	println("Number of keys deleted:", res)
	if err != nil {
		return err
	}

	return nil
}
