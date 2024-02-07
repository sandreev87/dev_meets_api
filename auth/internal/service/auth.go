package service

import (
	"auth/internal/config"
	"auth/internal/domain/models"
	"auth/pkg/jwt"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

const (
	tokenTTL = 30 * time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	repo      UserStorageInt
	RedisRepo RedisStorageInt
	conf      *config.Config
	logger    *slog.Logger
}

func NewAuthService(repo UserStorageInt, redis RedisStorageInt, conf *config.Config, logger *slog.Logger) *AuthService {
	return &AuthService{repo: repo, RedisRepo: redis, conf: conf, logger: logger}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "service.AuthService.Login"

	s.logger.Info("attempting to login user")

	user, err := s.repo.UserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password)); err != nil {
		s.logger.Info("invalid credentials")

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	s.logger.Info("user logged in successfully")

	token, err := jwt.NewToken(user, s.conf.Secret, tokenTTL)
	if err != nil {
		s.logger.Error("failed to generate token")

		return "", fmt.Errorf("%s: %w", op, err)
	}
	if err := s.RedisRepo.SetUserId(ctx, token, user.ID, tokenTTL); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *AuthService) Verify(ctx context.Context, token string) error {
	const op = "service.AuthService.Verify"

	if _, err := jwt.VerifyToken(token, s.conf.Secret); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	uid, err := s.RedisRepo.GetUserId(ctx, token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if uid == "" {
		return fmt.Errorf("user is not found")
	}
	return nil
}

func (s *AuthService) RegisterNewUser(ctx context.Context, user models.User, pass string) (int, error) {
	const op = "service.AuthService.RegisterNewUser"

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to generate password hash")

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	user.PassHash = string(passHash)

	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	const op = "service.AuthService.Logout"

	err := s.RedisRepo.DeleteUser(ctx, token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
