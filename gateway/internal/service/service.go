package service

import (
	"dev_meets/internal/storage"
	"log/slog"
)

type Service struct {
	*AuthService
	*UserService
}

func NewService(repos *storage.Repository, logger *slog.Logger) *Service {
	return &Service{
		AuthService: NewAuthService(repos.UserPostgres, repos.UserRedis, logger),
		UserService: NewUserService(repos.UserPostgres, logger),
	}
}
