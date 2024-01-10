package migrator

import (
	"dev_meets/internal/config"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(cfg *config.Config) {
	m, err := migrate.New(
		"file://"+cfg.Postgresql.MigrationsPath,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.Postgresql.User, cfg.Postgresql.Password, cfg.Postgresql.Host, cfg.Postgresql.Port, cfg.Postgresql.DB))

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied")
}
