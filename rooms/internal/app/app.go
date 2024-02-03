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
	DispatchKeyFrames()
}

type App struct {
	HTTPServer *http.Server
	RoomServ   RoomServiceInt
	logger     *slog.Logger
	config     *config.Config
}

func New(
	log *slog.Logger,
	conf *config.Config,
) *App {

	services := service.NewService(log)
	handlers := transport.NewHandler(services, log)
	router := handlers.InitRoutes()

	log.Info("initializing server", slog.String("address", conf.HTTPServer.Address))

	srv := &http.Server{
		Addr:    conf.HTTPServer.Address,
		Handler: router,
	}

	return &App{
		HTTPServer: srv,
		RoomServ:   services.RoomService,
		logger:     log,
		config:     conf,
	}
}

func (a *App) Run() {
	a.logger.Info("starting server")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go a.RoomServ.DispatchKeyFrames()
	go func() {
		if err := a.HTTPServer.ListenAndServe(); err != nil {
			a.logger.Error("failed to start server", err)
		}
	}()

	a.logger.Info("server started")

	<-done

	a.logger.Info("stopping server")
	a.Stop()
	a.logger.Info("server stopped")
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.RoomServ.CloseAllConnections()
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.logger.Error("failed to stop server")

		return
	}
}
