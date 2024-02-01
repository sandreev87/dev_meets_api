package app

import (
	"auth/internal/config"
	"auth/internal/service"
	"auth/internal/storage"
	"auth/internal/transport/rest"
	"context"
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
	repos      *storage.Repository
}

func New(
	log *slog.Logger,
	conf *config.Config,
) *App {
	repos := storage.NewRepository(conf, log)
	services := service.NewService(repos, conf, log)
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
		repos:      repos,
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
	a.repos.CloseConnections()
}