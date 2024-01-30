package service

import (
	"auth/internal/config"
	"auth/internal/storage"
	"log/slog"
)

type Service struct {
	*AuthService
	*UserService
}

func NewService(repos *storage.Repository, conf *config.Config, logger *slog.Logger) *Service {
	return &Service{
		AuthService: NewAuthService(repos.UserPostgres, repos.UserRedis, conf, logger),
		UserService: NewUserService(repos.UserPostgres, conf, logger),
	}
}
