package webrtc

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"sync"
)

const MaxNumberAttempts = 25

type Room struct {
	ID           string
	listLock     sync.RWMutex
	peers        map[string]*Peer
	outputTracks map[string]*webrtc.TrackLocalStaticRTP
}

func NewRoom(uuid string) *Room {
	return &Room{
		ID:           uuid,
		listLock:     sync.RWMutex{},
		peers:        map[string]*Peer{},
		outputTracks: map[string]*webrtc.TrackLocalStaticRTP{},
	}
}

func (r *Room) AddTrack(t *webrtc.TrackRemote) (*webrtc.TrackLocalStaticRTP, error) {
	r.listLock.Lock()
	defer func() {
		r.listLock.Unlock()
	}()

	trackID, err := r.makeTrackID(t)
	if err != nil {
		return nil, err
	}

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, trackID, t.StreamID(), webrtc.WithRTPStreamID(t.RID()))
	if err != nil {
		return nil, err
	}

	r.outputTracks[trackID] = trackLocal
	return trackLocal, nil
}

func (r *Room) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	r.listLock.Lock()
	defer func() {
		r.listLock.Unlock()
	}()

	delete(r.outputTracks, t.ID())
}

func (r *Room) makeTrackID(t *webrtc.TrackRemote) (string, error) {
	switch t.Kind() {
	case webrtc.RTPCodecTypeAudio:
		return fmt.Sprintf("audio_%s", t.ID()), nil
	case webrtc.RTPCodecTypeVideo:
		if t.RID() == "" {
			return "", errors.New("track's RID is blank")
		}
		return fmt.Sprintf("video_%s_%s", t.RID(), t.ID()), nil
	default:
		return "", errors.New("forbidden codec type")
	}
}

func (r *Room) SignalAllPeers() {
	r.listLock.Lock()
	defer func() {
		r.listLock.Unlock()
	}()

	for syncAttempt := 0; syncAttempt <= MaxNumberAttempts; syncAttempt++ {
		if r.attemptSync() {
			break
		}
	}
}

func (r *Room) attemptSync() bool {
	for _, peer := range r.peers {
		if !peer.Sync(r.outputTracks) {
			return false
		}
	}
	return true
}

func (r *Room) AddPeerConnection(newPeer *Peer) {
	r.listLock.Lock()
	r.peers[newPeer.ID] = newPeer
	r.listLock.Unlock()
}

func (r *Room) RemovePeerConnection(newPeer *Peer) {
	r.listLock.Lock()
	delete(r.peers, newPeer.ID)
	r.listLock.Unlock()
}

func (r *Room) Close() error {
	for _, peer := range r.peers {
		if err := peer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Room) DispatchKeyFrame() {
	r.listLock.Lock()
	defer r.listLock.Unlock()

	for _, peer := range r.peers {
		peer.DispatchKeyFrame()
	}
}
