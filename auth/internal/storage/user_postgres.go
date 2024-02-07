package storage

import (
	"auth/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log/slog"
)

type UserPostgres struct {
	db  *sql.DB
	log *slog.Logger
}

func NewUserPostgres(db *sql.DB, logger *slog.Logger) *UserPostgres {
	return &UserPostgres{db: db, log: logger}
}

func (r *UserPostgres) CreateUser(ctx context.Context, user models.User) (int, error) {
	const op = "repository.AuthPostgres.CreateUser"

	var id int
	err := r.db.QueryRowContext(ctx, "INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id", user.Email, user.PassHash).Scan(&id)

	if err != nil {
		var pgsErr *pq.Error
		if errors.As(err, &pgsErr) && pgsErr.Code.Name() == "unique_violation" {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *UserPostgres) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "repository.AuthPostgres.UserByEmail"
	var user models.User

	err := r.db.QueryRowContext(ctx, "SELECT id, email, pass_hash FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *UserPostgres) User(ctx context.Context, id int) (models.User, error) {
	const op = "repository.AuthPostgres.UserByEmail"
	var user models.User

	err := r.db.QueryRowContext(ctx, "SELECT id, email, pass_hash FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
