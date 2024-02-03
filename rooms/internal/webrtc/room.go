package webrtc

import (
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"log"
	"sync"
	"time"
)

type Room struct {
	ListLock    sync.RWMutex
	Connections []PeerConnectionState
	TrackLocals map[string]*webrtc.TrackLocalStaticRTP
}

func NewRoom() *Room {
	return &Room{
		TrackLocals: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

func (r *Room) AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	r.ListLock.Lock()
	defer func() {
		r.ListLock.Unlock()
		r.SignalPeerConnections()
	}()

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	r.TrackLocals[t.ID()] = trackLocal
	return trackLocal
}

func (r *Room) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	r.ListLock.Lock()
	defer func() {
		r.ListLock.Unlock()
		r.SignalPeerConnections()
	}()

	delete(r.TrackLocals, t.ID())
}

func (r *Room) SignalPeerConnections() {
	r.ListLock.Lock()
	defer func() {
		r.ListLock.Unlock()
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
	for i := range r.Connections {
		if r.Connections[i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			r.Connections = append(r.Connections[:i], r.Connections[i+1:]...)
			return false
		}

		existingSenders := map[string]struct{}{}
		for _, sender := range r.Connections[i].PeerConnection.GetSenders() {
			if sender.Track() == nil {
				continue
			}

			existingSenders[sender.Track().ID()] = struct{}{}

			if _, ok := r.TrackLocals[sender.Track().ID()]; !ok {
				if err := r.Connections[i].PeerConnection.RemoveTrack(sender); err != nil {
					return false
				}
			}
		}

		for _, receiver := range r.Connections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			existingSenders[receiver.Track().ID()] = struct{}{}
		}

		for trackID := range r.TrackLocals {
			if _, ok := existingSenders[trackID]; !ok {
				if _, err := r.Connections[i].PeerConnection.AddTrack(r.TrackLocals[trackID]); err != nil {
					return false
				}
			}
		}

		if r.Connections[i].ShouldSendOffer() {
			if err := r.Connections[i].SendOffer(); err != nil {
				return false
			}
		}
	}

	return true
}

func (r *Room) AddPeerConnection(newPeer *PeerConnectionState) {
	r.ListLock.Lock()
	r.Connections = append(r.Connections, *newPeer)
	r.ListLock.Unlock()
}

func (r *Room) DispatchKeyFrame() {
	r.ListLock.Lock()
	defer r.ListLock.Unlock()

	for i := range r.Connections {
		for _, receiver := range r.Connections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = r.Connections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
