package service

import (
	"dev_meets/internal/domain/models"
	"dev_meets/pkg/jwt"
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
	repo   UserStorageInt
	logger *slog.Logger
}

func NewAuthService(repo UserStorageInt, logger *slog.Logger) *AuthService {
	return &AuthService{repo: repo, logger: logger}
}

func (s *AuthService) Login(email, password string) (string, error) {
	const op = "service.AuthService.Login"

	s.logger.Info("attempting to login user")

	user, err := s.repo.UserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password)); err != nil {
		s.logger.Info("invalid credentials")

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	s.logger.Info("user logged in successfully")

	token, err := jwt.NewToken(user, tokenTTL)
	if err != nil {
		s.logger.Error("failed to generate token")

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *AuthService) RegisterNewUser(user models.User, pass string) (int, error) {
	const op = "service.AuthService.RegisterNewUser"

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to generate password hash")

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	user.PassHash = string(passHash)

	return s.repo.CreateUser(user)
}
