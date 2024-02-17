package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"rooms/internal/config"
	"rooms/internal/service"
	"rooms/internal/transport"
	"syscall"
	"time"
)

type RoomServiceInt interface {
	CloseAllConnections()
	DispatchKeyFrames(ctx context.Context)
	Sync(ctx context.Context)
}

type App struct {
	HTTPServer *http.Server
	roomServ   RoomServiceInt
	logger     *slog.Logger
}

func New(
	logger *slog.Logger,
	config *config.Config,
) *App {

	services := service.NewService(logger)
	handlers := transport.NewHandler(services, logger)
	router := handlers.InitRoutes()

	logger.Info("initializing server", slog.String("address", config.HTTPServer.Address))

	srv := &http.Server{
		Addr:    config.HTTPServer.Address,
		Handler: router,
	}

	return &App{
		HTTPServer: srv,
		roomServ:   services.RoomService,
		logger:     logger,
	}
}

func (a *App) Run() {
	a.logger.Info("starting server")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go a.roomServ.Sync(ctx)
	go a.roomServ.DispatchKeyFrames(ctx)
	go func() {
		if err := a.HTTPServer.ListenAndServe(); err != nil {
			a.logger.Error("failed to start server", err)
		}
	}()

	a.logger.Info("server started")

	<-done
	cancel()

	a.logger.Info("stopping server")
	a.Stop()
	a.logger.Info("server stopped")
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.roomServ.CloseAllConnections()
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.logger.Error("failed to stop server")

		return
	}
}
