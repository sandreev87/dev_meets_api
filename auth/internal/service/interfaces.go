package service

import (
	"auth/internal/domain/models"
	"time"
)

type UserStorageInt interface {
	CreateUser(user models.User) (int, error)
	UserByEmail(email string) (models.User, error)
	User(id int) (models.User, error)
}

type RedisStorageInt interface {
	GetUserId(tokenId string) (string, error)
	SetUserId(tokenId string, id int, duration time.Duration) error
	DeleteUser(tokenId string) error
}
