package webrtc

import (
	"context"
	"errors"
	"fmt"
	guuid "github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"strings"
	"sync"
)

const (
	DefaultQuality     = "low"
	AudioPrefix        = "audio"
	ChangeQualityEvent = "change_quality"
)

type Peer struct {
	ID                     string
	peerConnection         *webrtc.PeerConnection
	NegotiationCoordinator *NegotiationCoordinator
	CurrentQuality         string
	syncMx                 sync.RWMutex
}

func NewPeer() (*Peer, error) {
	config := webrtc.Configuration{}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	return &Peer{
		ID:                     guuid.New().String(),
		peerConnection:         peerConnection,
		NegotiationCoordinator: NewNegotiationCoordinator(peerConnection),
		CurrentQuality:         DefaultQuality,
	}, nil
}

func (peer *Peer) OnTrack(f func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) {
	peer.peerConnection.OnTrack(f)
}

func (peer *Peer) OnICECandidate(f func(*webrtc.ICECandidate)) {
	peer.peerConnection.OnICECandidate(f)
}

func (peer *Peer) OnConnectionStateChange(f func(webrtc.PeerConnectionState)) {
	peer.peerConnection.OnConnectionStateChange(f)
}

func (peer *Peer) OutputTrackIDs() (ids []string) {
	senders := peer.peerConnection.GetSenders()
	for index := range senders {
		sender := senders[index]
		if sender.Track() == nil {
			continue
		}
		ids = append(ids, sender.Track().ID())
	}
	return
}

func (peer *Peer) InputTrackIDs() (ids []string) {
	receivers := peer.peerConnection.GetReceivers()
	for index, _ := range receivers {
		for _, track := range receivers[index].Tracks() {
			ids = append(ids, track.RID()+"_"+track.ID())
		}
	}
	return
}

func (peer *Peer) AddTrack(track *webrtc.TrackLocalStaticRTP) error {
	if !peer.CanAddTrack(track.ID()) {
		return errors.New(fmt.Sprintf("Peer is not support quality of this track. Current quality: %s", peer.CurrentQuality))
	}

	if _, err := peer.peerConnection.AddTrack(track); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) RemoveTrack(id string) error {
	for _, sender := range peer.peerConnection.GetSenders() {
		if sender.Track() == nil || sender.Track().ID() != id {
			continue
		}

		if err := peer.peerConnection.RemoveTrack(sender); err != nil {
			return err
		} else {
			return nil
		}
	}
	return nil
}

func (peer *Peer) CanAddTrack(id string) bool {
	return strings.HasPrefix(id, peer.CurrentQuality) || strings.HasPrefix(id, AudioPrefix)
}

func (peer *Peer) Close() error {
	if err := peer.peerConnection.Close(); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) DispatchKeyFrame() {
	for _, receiver := range peer.peerConnection.GetReceivers() {
		for _, track := range receiver.Tracks() {
			if track == nil {
				continue
			}

			_ = peer.peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(track.SSRC()),
				},
			})
		}
	}
}

func (peer *Peer) ChangeQuality(quality string) {
	peer.syncMx.Lock()
	peer.CurrentQuality = quality
	peer.syncMx.Unlock()
}

func (peer *Peer) HandleEvent(event string, data string) error {
	switch event {
	case ChangeQualityEvent:
		peer.ChangeQuality(data)
	default:
		if err := peer.NegotiationCoordinator.HandleEvent(event, data); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (peer *Peer) ListenSendEvents(ctx context.Context, callback func(string, string)) {
	peer.NegotiationCoordinator.ListenSendEvents(ctx, callback)
}

func (peer *Peer) Sync(outputTracks map[string]*webrtc.TrackLocalStaticRTP) bool {
	peer.syncMx.RLock()
	peerChanged := false
	existingSenders := map[string]struct{}{}

	// map of sender we already are sending, so we don't double send
	for _, id := range peer.OutputTrackIDs() {
		existingSenders[id] = struct{}{}

		if _, ok := outputTracks[id]; !ok {
			if err := peer.RemoveTrack(id); err != nil {
				return false
			}
			peerChanged = true
		}
	}

	// Don't receive videos we are sending, make sure we don't have loopback
	for _, id := range peer.InputTrackIDs() {
		existingSenders[id] = struct{}{}
	}

	for trackID := range outputTracks {
		if !peer.CanAddTrack(trackID) {
			continue
		}
		if _, ok := existingSenders[trackID]; !ok {
			if err := peer.AddTrack(outputTracks[trackID]); err != nil {
				return false
			}
			peerChanged = true
		}
	}

	if peerChanged {
		if err := peer.NegotiationCoordinator.SendOffer(); err != nil {
			return false
		}
	}

	peer.syncMx.RUnlock()
	return true
}
