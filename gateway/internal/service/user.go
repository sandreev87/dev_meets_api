package service

import (
	"dev_meets/internal/domain/models"
	"dev_meets/pkg/jwt"
	"fmt"
	"log/slog"
)

type UserService struct {
	repo   UserStorageInt
	logger *slog.Logger
}

func NewUserService(repo UserStorageInt, logger *slog.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

func (s *UserService) CurrentUser(token string) (models.User, error) {
	const op = "service.UserService.CurrentUser"
	uid, _ := jwt.VerifyToken(token)
	user, err := s.repo.User(uid)

	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
