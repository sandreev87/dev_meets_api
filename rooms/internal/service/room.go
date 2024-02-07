package service

import (
	"crypto/sha256"
	"fmt"
	"github.com/pion/webrtc/v3"
	"log/slog"
	wrtc "rooms/internal/webrtc"
	"sync"
	"time"
)

type RoomService struct {
	RoomsLock sync.RWMutex
	Rooms     map[string]*wrtc.Room
	logger    *slog.Logger
}

func NewRoomService(logger *slog.Logger) *RoomService {
	return &RoomService{Rooms: make(map[string]*wrtc.Room), logger: logger}
}

func (s *RoomService) CreateOrGetRoom(uuid string) (*wrtc.Room, error) {
	s.RoomsLock.Lock()
	defer s.RoomsLock.Unlock()

	hash := sha256.New()
	hash.Write([]byte(uuid))

	if room := s.Rooms[uuid]; room != nil {
		return room, nil
	}

	room := wrtc.NewRoom()

	s.Rooms[uuid] = room

	return room, nil
}

func (s *RoomService) InitPeerConnection(room *wrtc.Room) (*wrtc.PeerConnectionState, error) {
	const op = "service.RoomService.InitPeerConnection"

	newPeer, err := wrtc.NewPeerConnectionState()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	room.AddPeerConnection(newPeer)

	// Accept one audio and one video track incoming
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := newPeer.PeerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	// Trickle ICE. Emit server candidate to client
	newPeer.PeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if err := newPeer.SendICECandidate(i); err != nil {
			s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
			return
		}
	})

	// If PeerConnection is closed remove it from global list
	newPeer.PeerConnection.OnConnectionStateChange(func(pp webrtc.PeerConnectionState) {
		switch pp {
		case webrtc.PeerConnectionStateFailed:
			if err := newPeer.Close(); err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
			}
		case webrtc.PeerConnectionStateClosed:
			room.SignalPeerConnections()
		}
	})

	newPeer.PeerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		// Create a track to fan out our incoming video to all peers
		trackLocal := room.AddTrack(t)
		if trackLocal == nil {
			return
		}
		defer func() {
			room.RemoveTrack(trackLocal)
			room.SignalPeerConnections()
		}()

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			if err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
				return
			}

			_, err = trackLocal.Write(buf[:i])
			if err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
				return
			}
		}
	})

	room.SignalPeerConnections()

	return newPeer, nil
}

func (s *RoomService) CloseAllConnections() {
	const op = "service.RoomService.CloseAllConnections"

	for _, room := range s.Rooms {
		for _, peer := range room.Connections {
			if err := peer.Close(); err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
			}
		}
	}
}

func (s *RoomService) DispatchKeyFrames() {
	for range time.NewTicker(time.Second * 3).C {
		for _, room := range s.Rooms {
			room.DispatchKeyFrame()
		}
	}
}
