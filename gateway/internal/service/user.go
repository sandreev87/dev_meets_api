package service

import (
	"dev_meets/internal/config"
	"dev_meets/internal/domain/models"
	"dev_meets/pkg/jwt"
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

func (s *UserService) CurrentUser(token string) (models.User, error) {
	const op = "service.UserService.CurrentUser"
	uid, _ := jwt.VerifyToken(token, s.conf.Secret)
	user, err := s.repo.User(uid)

	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
