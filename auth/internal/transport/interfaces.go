package transport

import "auth/internal/domain/models"

type AuthorizationServiceInt interface {
	RegisterNewUser(user models.User, pass string) (int, error)
	Login(username, password string) (string, error)
	Logout(token string) error
	Verify(token string) error
}

type UserServiceInt interface {
	CurrentUser(token string) (models.User, error)
}
