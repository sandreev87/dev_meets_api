package app

import (
	"context"
	"database/sql"
	"dev_meets/internal/config"
	"dev_meets/internal/service"
	"dev_meets/internal/storage"
	"dev_meets/internal/transport/rest"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	HTTPServer *http.Server
	logger     *slog.Logger
	config     *config.Config
	db         *sql.DB
}

func New(
	log *slog.Logger,
	conf *config.Config,
) *App {
	db := initDbConnection(conf)
	repos := storage.NewRepository(db, log)
	services := service.NewService(repos, log)
	handlers := rest.NewHandler(services, log)
	router := handlers.InitRoutes()

	log.Info("initializing server", slog.String("address", conf.HTTPServer.Address))

	srv := &http.Server{
		Addr:         conf.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  conf.HTTPServer.Timeout,
		WriteTimeout: conf.HTTPServer.Timeout,
		IdleTimeout:  conf.HTTPServer.IdleTimeout,
	}

	return &App{
		HTTPServer: srv,
		logger:     log,
		config:     conf,
		db:         db,
	}
}

func (a *App) Run() {
	a.logger.Info("starting server", slog.String("address", a.config.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := a.HTTPServer.ListenAndServe(); err != nil {
			a.logger.Error("failed to start server")
		}
	}()

	a.logger.Info("server started")

	<-done
	a.logger.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.logger.Error("failed to stop server")

		return
	}

	a.Stop()
	a.logger.Info("server stopped")
}

func (a *App) Stop() {
	a.db.Close()
}

func initDbConnection(cnf *config.Config) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cnf.Postgresql.Host, cnf.Postgresql.Port, cnf.Postgresql.User, cnf.Postgresql.Password, cnf.Postgresql.DB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}
