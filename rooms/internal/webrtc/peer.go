package webrtc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"strings"
	"sync"
)

const DefaultQuality = "high"

type Peer struct {
	peerConnection    *webrtc.PeerConnection
	internalEventChan chan InternalEvent
	CurrentQuality    string
	lock              sync.Mutex
}

type InternalEvent struct {
	Event string
	Data  string
}

func NewPeer() (*Peer, error) {
	config := webrtc.Configuration{}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	return &Peer{
		peerConnection:    peerConnection,
		CurrentQuality:    DefaultQuality,
		internalEventChan: make(chan InternalEvent, 20),
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

func (peer *Peer) OutputTrackIDs() []string {
	var ids []string
	for _, sender := range peer.peerConnection.GetSenders() {
		if sender.Track() == nil {
			continue
		}
		ids = append(ids, sender.Track().ID())
	}
	return ids
}

func (peer *Peer) InputTrackIDs() []string {
	var ids []string
	for _, receiver := range peer.peerConnection.GetReceivers() {
		for _, track := range receiver.Tracks() {
			trackID := track.RID() + "_" + track.ID()
			ids = append(ids, trackID)
		}
	}
	return ids
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
		if sender.Track() != nil && sender.Track().ID() == id {
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

func (peer *Peer) IsClosed() bool {
	return peer.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed
}

func (peer *Peer) CanAddTrack(id string) bool {
	return strings.HasPrefix(id, peer.CurrentQuality) || strings.HasPrefix(id, "audio")
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

func (peer *Peer) AcceptAnswer(data string) error {
	answer := webrtc.SessionDescription{}
	if err := json.Unmarshal([]byte(data), &answer); err != nil {
		return err
	}

	if err := peer.peerConnection.SetRemoteDescription(answer); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) SendAnswer() error {
	answer, err := peer.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := peer.peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	answerString, err := json.Marshal(answer)
	peer.HandleEvent("send_answer", string(answerString))
	return nil
}

func (peer *Peer) SendOffer() error {
	offer, err := peer.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = peer.peerConnection.SetLocalDescription(offer); err != nil {
		return err
	}

	offerString, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	peer.HandleEvent("send_offer", string(offerString))
	return nil
}

func (peer *Peer) AcceptOffer(data string) error {
	offer := webrtc.SessionDescription{}

	if err := json.Unmarshal([]byte(data), &offer); err != nil {
		return err
	}

	if err := peer.peerConnection.SetRemoteDescription(offer); err != nil {
		return err
	}

	if err := peer.SendAnswer(); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) AcceptICECandidate(data string) error {
	candidate := webrtc.ICECandidateInit{}
	if err := json.Unmarshal([]byte(data), &candidate); err != nil {
		return err
	}

	if err := peer.peerConnection.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (peer *Peer) SendICECandidate(i *webrtc.ICECandidate) error {
	if i == nil {
		return nil
	}

	candidateString, err := json.Marshal(i.ToJSON())
	if err != nil {
		return err
	}
	peer.HandleEvent("send_candidate", string(candidateString))
	return nil
}

func (peer *Peer) HandleEvent(event string, data string) error {
	switch event {
	case "send_offer":
		peer.internalEventChan <- InternalEvent{Event: "offer", Data: data}
	case "send_answer":
		peer.internalEventChan <- InternalEvent{Event: "answer", Data: data}
	case "send_candidate":
		peer.internalEventChan <- InternalEvent{Event: "candidate", Data: data}
	case "offer":
		if err := peer.AcceptOffer(data); err != nil {
			return err
		}
	case "candidate":
		if err := peer.AcceptICECandidate(data); err != nil {
			return err
		}
	case "answer":
		if err := peer.AcceptAnswer(data); err != nil {
			return err
		}
	}
	return nil
}

func (peer *Peer) ListenSendEvents(ctx context.Context, callback func(string, string)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			case event := <-peer.internalEventChan:
				callback(event.Event, event.Data)
			}
		}
	}()
}
