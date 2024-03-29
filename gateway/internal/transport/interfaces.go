package transport

import "dev_meets/internal/domain/models"

type AuthorizationServiceInt interface {
	RegisterNewUser(user models.User, pass string) (int, error)
	Login(username, password string) (string, error)
}

type UserServiceInt interface {
	CurrentUser(token string) (models.User, error)
}
