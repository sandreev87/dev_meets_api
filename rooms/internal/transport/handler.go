package transport

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	guuid "github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"rooms/internal/service"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type RoomHandlerInt interface {
	SignalHandler(w http.ResponseWriter, r *http.Request)
	IndexHandler(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	*RoomHandler
}

func NewHandler(services *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		RoomHandler: NewRoomHandler(services.RoomService, logger),
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/room/%s", guuid.New().String()), http.StatusMovedPermanently)
	})
	router.Get("/room/{roomID}", h.RoomHandler.IndexHandler)

	router.HandleFunc("/websocket/room/{roomID}", h.RoomHandler.SignalHandler)

	return router
}
