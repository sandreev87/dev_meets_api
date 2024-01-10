package storage

import (
	"dev_meets/internal/config"
	"log/slog"
)

type Repository struct {
	*UserPostgres
	*UserRedis
}

func (r *Repository) CloseConnections() {
	r.UserPostgres.db.Close()
}

func NewRepository(config *config.Config, logger *slog.Logger) *Repository {
	postgresDB := initDbConnection(config)
	redis := initRedisConnection()
	return &Repository{
		UserPostgres: NewUserPostgres(postgresDB, logger),
		UserRedis:    NewUserRedis(redis, logger),
	}
}
