package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"rooms/internal/config"
	rest "rooms/internal/transport"
	wrtc "rooms/internal/webrtc"
	"syscall"
	"time"
)

type App struct {
	HTTPServer *http.Server
	logger     *slog.Logger
	config     *config.Config
}

func New(
	log *slog.Logger,
	conf *config.Config,
) *App {
	handlers := rest.NewHandler(log)
	router := handlers.InitRoutes()

	log.Info("initializing server", slog.String("address", conf.HTTPServer.Address))

	srv := &http.Server{
		Addr:    conf.HTTPServer.Address,
		Handler: router,
	}

	return &App{
		HTTPServer: srv,
		logger:     log,
		config:     conf,
	}
}

func (a *App) Run() {
	a.logger.Info("starting server")
	wrtc.Rooms = make(map[string]*wrtc.Room)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go dispatchKeyFrames()
	go func() {
		if err := a.HTTPServer.ListenAndServe(); err != nil {
			a.logger.Error("failed to start server", err)
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

func dispatchKeyFrames() {
	for range time.NewTicker(time.Second * 3).C {
		for _, room := range wrtc.Rooms {
			room.Peers.DispatchKeyFrame()
		}
	}
}

func (a *App) Stop() {
}
