package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/pion/webrtc/v4"
	"log/slog"
	wrtc "rooms/internal/webrtc"
	"sync"
	"time"
)

type RoomService struct {
	roomsLock sync.RWMutex
	rooms     map[string]*wrtc.Room
	logger    *slog.Logger
}

func NewRoomService(logger *slog.Logger) *RoomService {
	return &RoomService{
		rooms:  make(map[string]*wrtc.Room),
		logger: logger,
	}
}

func (s *RoomService) CreateOrGetRoom(uuid string) (*wrtc.Room, error) {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	hash := sha256.New()
	hash.Write([]byte(uuid))

	if room := s.rooms[uuid]; room != nil {
		return room, nil
	}

	room := wrtc.NewRoom(uuid)

	s.rooms[uuid] = room

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
			room.RemovePeerConnection(newPeer)
			if err := newPeer.Close(); err != nil {
				s.logger.Error(fmt.Errorf("OnConnectionStateChange callback %s: %w", op, err).Error())
			}
		case webrtc.PeerConnectionStateClosed:
			room.RemovePeerConnection(newPeer)
			s.logger.Debug(fmt.Sprintf("connection: %s was closed", newPeer.ID))
		default:
			return
		}
	})

	newPeer.OnICECandidate(func(i *webrtc.ICECandidate) {
		if err := newPeer.NegotiationCoordinator.SendICECandidate(i); err != nil {
			s.logger.Error(fmt.Errorf("OnICECandidate callback %s: %w", op, err).Error())
			return
		}
	})

	newPeer.OnTrack(func(t *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		output, err := room.AddTrack(t)
		if err != nil {
			s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
			return
		}
		s.logger.Debug(fmt.Sprintf("%s: room: %s track with ID %s was added into store", op, room.ID, output.ID()))

		defer func() {
			room.RemoveTrack(output)
			s.logger.Debug(fmt.Sprintf("%s: track: %s was removed from store", op, output.ID()))
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

	for _, room := range s.rooms {
		if err := room.Close(); err != nil {
			s.logger.Error(fmt.Errorf("%s: %w", op, err).Error())
		}
	}
}

func (s *RoomService) Sync(ctx context.Context) {
	const op = "service.RoomService.Sync"

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.NewTicker(time.Second).C:
			for _, room := range s.rooms {
				go func(r *wrtc.Room) {
					r.SignalAllPeers()
				}(room)
			}
		}
	}
}

func (s *RoomService) DispatchKeyFrames(ctx context.Context) {
	const op = "service.RoomService.DispatchKeyFrames"

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.NewTicker(time.Second * 3).C:
			for _, room := range s.rooms {
				room.DispatchKeyFrame()
			}
		}
	}
}
