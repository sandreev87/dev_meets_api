package service

import (
	"auth/internal/config"
	"auth/internal/domain/models"
	"auth/pkg/jwt"
	"context"
	"fmt"
	"log/slog"
)

type UserService struct {
	repo   UserStorageInt
	conf   *config.Config
	logger *slog.Logger
}

func NewUserService(repo UserStorageInt, conf *config.Config, logger *slog.Logger) *UserService {
	return &UserService{repo: repo, conf: conf, logger: logger}
}

func (s *UserService) CurrentUser(ctx context.Context, token string) (*models.User, error) {
	const op = "service.UserService.CurrentUser"

	uid, err := jwt.VerifyToken(token, s.conf.Secret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.repo.User(ctx, uid)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
