package transport

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"log/slog"
	"net/http"
	"os"
	wrtc "rooms/internal/webrtc"
	"sync"
	"text/template"
)

type RoomServiceInt interface {
	CreateOrGetRoom(uuid string) (*wrtc.Room, error)
	InitPeerConnection(room *wrtc.Room) (*wrtc.PeerConnectionState, error)
}

type PeerInt interface {
	HandleEvent(event string, data string) error
	Close() error
}

type RoomHandler struct {
	service RoomServiceInt
	logger  *slog.Logger
}

func NewRoomHandler(service RoomServiceInt, logger *slog.Logger) *RoomHandler {
	return &RoomHandler{service: service, logger: logger}
}

func (h *RoomHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	indexHTML, _ := os.ReadFile("web/index.html")
	indexTemplate := template.Must(template.New("").Parse(string(indexHTML)))
	roomId := chi.URLParam(r, "roomID")
	_, _ = h.service.CreateOrGetRoom(roomId)
	if err := indexTemplate.Execute(w, "wss://"+r.Host+"/websocket/room/"+roomId); err != nil {
		log.Fatal(err)
	}
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

func (h *RoomHandler) SignalHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to Websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error(err.Error())
		return
	}
	ctx, close := context.WithCancel(r.Context())
	roomID := chi.URLParam(r, "roomID")
	room, err := h.service.CreateOrGetRoom(roomID)
	if err != nil {
		h.logger.Error(err.Error())
		close()
		return
	}
	peer, err := h.service.InitPeerConnection(room)
	if err != nil {
		h.logger.Error(err.Error())
		close()
		return
	}
	defer func() {
		if err := peer.Close(); err != nil {
			h.logger.Error(err.Error())
		}
		close()
	}()

	loc := sync.Mutex{}
	peer.ListenEvents(ctx, func(event string, data string) {
		msg := &websocketMessage{
			Event: event,
			Data:  data,
		}
		loc.Lock()
		if err = conn.WriteJSON(msg); err != nil {
			h.logger.Error(err.Error())
		}
		loc.Unlock()
	})

	for {
		message := &websocketMessage{}
		_, raw, err := conn.ReadMessage()
		if err != nil {
			h.logger.Error(err.Error())
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			h.logger.Error(err.Error())
			return
		}
		if err := peer.HandleEvent(message.Event, message.Data); err != nil {
			h.logger.Error(err.Error())
			return
		}
	}
}
