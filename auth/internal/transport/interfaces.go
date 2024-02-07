package transport

import (
	"auth/internal/domain/models"
	"context"
)

type AuthorizationServiceInt interface {
	RegisterNewUser(ctx context.Context, user models.User, pass string) (int, error)
	Login(ctx context.Context, username, password string) (string, error)
	Logout(ctx context.Context, token string) error
	Verify(ctx context.Context, token string) error
}

type UserServiceInt interface {
	CurrentUser(ctx context.Context, token string) (models.User, error)
}
