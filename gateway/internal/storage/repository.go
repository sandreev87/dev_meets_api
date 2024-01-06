package storage

import (
	"database/sql"
	"log/slog"
)

type Repository struct {
	*UserPostgres
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{
		UserPostgres: NewUserPostgres(db, logger),
	}
}
