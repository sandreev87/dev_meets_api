package service

import "dev_meets/internal/domain/models"

type UserStorageInt interface {
	CreateUser(user models.User) (int, error)
	UserByEmail(email string) (models.User, error)
	User(id int) (models.User, error)
}
