package service

import (
	"auth/internal/domain/models"
	"context"
	"time"
)

type UserStorageInt interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	UserByEmail(ctx context.Context, email string) (models.User, error)
	User(ctx context.Context, id int) (models.User, error)
}

type RedisStorageInt interface {
	GetUserId(ctx context.Context, tokenId string) (string, error)
	SetUserId(ctx context.Context, tokenId string, id int, duration time.Duration) error
	DeleteUser(ctx context.Context, tokenId string) error
}
