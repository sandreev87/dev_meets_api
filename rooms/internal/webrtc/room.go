package webrtc

import (
	"errors"
	"github.com/pion/webrtc/v4"
	"sync"
	"time"
)

type Room struct {
	ID           string
	listLock     sync.RWMutex
	peers        []*Peer
	outputTracks map[string]*webrtc.TrackLocalStaticRTP
}

func NewRoom(uuid string) *Room {
	return &Room{
		ID:           uuid,
		listLock:     sync.RWMutex{},
		peers:        []*Peer{},
		outputTracks: map[string]*webrtc.TrackLocalStaticRTP{},
	}
}

func (r *Room) AddTrack(t *webrtc.TrackRemote) (*webrtc.TrackLocalStaticRTP, error) {
	r.listLock.Lock()
	defer func() {
		r.listLock.Unlock()
		r.SignalPeerConnections()
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

func (r *Room) makeTrackID(t *webrtc.TrackRemote) (string, error) {
	if t.Kind().String() == "audio" {
		return "audio" + "_" + t.ID(), nil
	} else {
		if t.RID() == "" {
			return "", errors.New("track's RID is blank")
		}
		return t.RID() + "_" + t.ID(), nil
	}
}

func (r *Room) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	r.listLock.Lock()
	defer func() {
		r.listLock.Unlock()
		r.SignalPeerConnections()
	}()

	delete(r.outputTracks, t.ID())
}

func (r *Room) SignalPeerConnections() {
	r.listLock.Lock()
	defer func() {
		r.listLock.Unlock()
		r.DispatchKeyFrame()
	}()

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			go func() {
				time.Sleep(time.Second * 3)
				r.SignalPeerConnections()
			}()
			return
		}

		if r.attemptSync() {
			break
		}
	}
}

func (r *Room) attemptSync() bool {
	for i := range r.peers {
		if r.peers[i].IsClosed() {
			r.peers = append(r.peers[:i], r.peers[i+1:]...)
			return false
		}

		peerChanged := false
		existingSenders := map[string]struct{}{}

		// map of sender we already are sending, so we don't double send
		for _, id := range r.peers[i].OutputTrackIDs() {
			existingSenders[id] = struct{}{}

			if _, ok := r.outputTracks[id]; !ok {
				if err := r.peers[i].RemoveTrack(id); err != nil {
					return false
				}
				peerChanged = true
			}
		}
		// Don't receive videos we are sending, make sure we don't have loopback
		for _, id := range r.peers[i].InputTrackIDs() {
			existingSenders[id] = struct{}{}
		}

		for trackID := range r.outputTracks {
			if !r.peers[i].CanAddTrack(trackID) {
				continue
			}
			if _, ok := existingSenders[trackID]; !ok {
				if err := r.peers[i].AddTrack(r.outputTracks[trackID]); err != nil {
					return false
				}
				peerChanged = true
			}
		}

		if peerChanged {
			r.peers[i].SendOffer()
		}
	}

	return true
}

func (r *Room) AddPeerConnection(newPeer *Peer) {
	r.listLock.Lock()
	r.peers = append(r.peers, newPeer)
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
	r.listLock.RLock()
	defer r.listLock.RUnlock()

	for _, peer := range r.peers {
		peer.DispatchKeyFrame()
	}
}
