package rest

import (
	"crypto/sha256"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	guuid "github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"log"
	"log/slog"
	"net/http"
	"os"
	wrtc "rooms/internal/webrtc"
	"text/template"
)

type Handler struct {
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func NewHandler(_ *slog.Logger) *Handler {
	return &Handler{}
}

func (h *Handler) InitRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	indexHTML, _ := os.ReadFile("index.html")
	indexTemplate := template.Must(template.New("").Parse(string(indexHTML)))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		newRoomID := guuid.New().String()
		log.Printf("new room: %s", newRoomID)
		http.Redirect(w, r, fmt.Sprintf("/room/%s", newRoomID), http.StatusMovedPermanently)
	})
	router.Get("/room/{roomID}", func(w http.ResponseWriter, r *http.Request) {
		roomId := chi.URLParam(r, "roomID")
		_, _ = createOrGetRoom(roomId)
		if err := indexTemplate.Execute(w, "wss://"+r.Host+"/websocket/room/"+roomId); err != nil {
			log.Fatal(err)
		}
	})

	router.HandleFunc("/websocket/room/{roomID}", websocketHandler)

	return router
}

func createOrGetRoom(uuid string) (string, *wrtc.Room) {
	wrtc.RoomsLock.Lock()
	defer wrtc.RoomsLock.Unlock()

	h := sha256.New()
	h.Write([]byte(uuid))

	if room := wrtc.Rooms[uuid]; room != nil {
		return uuid, room
	}

	p := &wrtc.Peers{}
	p.TrackLocals = make(map[string]*webrtc.TrackLocalStaticRTP)
	room := &wrtc.Room{
		Peers: p,
	}

	wrtc.Rooms[uuid] = room

	return uuid, room
}

// Handle incoming websockets
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to Websocket
	unsafeConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	roomId := chi.URLParam(r, "roomID")
	if roomId == "" {
		return
	}

	_, room := createOrGetRoom(roomId)
	wrtc.RoomConn(unsafeConn, room.Peers)
}
