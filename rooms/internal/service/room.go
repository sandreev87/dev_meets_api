package service

import (
	"crypto/sha256"
	"fmt"
	"github.com/pion/webrtc/v4"
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

	room := wrtc.NewRoom(uuid)

	s.Rooms[uuid] = room

	return room, nil
}

func (s *RoomService) InitPeerConnection(room *wrtc.Room) (*wrtc.Peer, error) {
	const op = "service.RoomService.InitPeerConnection"

	newPeer, err := wrtc.NewPeer()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	room.AddPeerConnection(newPeer)

	// If PeerConnection is closed remove it from global list
	newPeer.OnConnectionStateChange(func(pp webrtc.PeerConnectionState) {
		switch pp {
		case webrtc.PeerConnectionStateFailed:
			if err := newPeer.Close(); err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
			}
		case webrtc.PeerConnectionStateClosed:
			room.SignalPeerConnections()
		default:
			return
		}
	})

	newPeer.OnICECandidate(func(i *webrtc.ICECandidate) {
		if err := newPeer.SendICECandidate(i); err != nil {
			s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
			return
		}
	})

	newPeer.OnTrack(func(t *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		output, err := room.AddTrack(t)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
		s.logger.Debug(fmt.Sprintf("Room: %s track with ID %s was added into store", room.ID, output.ID()))

		defer func() {
			room.RemoveTrack(output)
			s.logger.Debug(fmt.Sprintf("track: %s was removed from store", output.ID()))
			room.SignalPeerConnections()
		}()

		buffer := make([]byte, 1500)
		for {
			i, _, err := t.Read(buffer)
			if err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
				return
			}
			_, err = output.Write(buffer[:i])
			if err != nil {
				s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
				return
			}
		}
	})

	return newPeer, nil
}

func (s *RoomService) CloseAllConnections() {
	const op = "service.RoomService.CloseAllConnections"

	for _, room := range s.Rooms {
		if err := room.Close(); err != nil {
			s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
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
