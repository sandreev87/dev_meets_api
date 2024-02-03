package webrtc

import (
	"encoding/json"
	"github.com/pion/webrtc/v3"
)

type PeerConnectionState struct {
	PeerConnection    *webrtc.PeerConnection
	InternalEventChan chan InternalEvent
}

type InternalEvent struct {
	Event string
	Data  string
}

func NewPeerConnectionState() (*PeerConnectionState, error) {
	var config webrtc.Configuration
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	return &PeerConnectionState{
		PeerConnection:    peerConnection,
		InternalEventChan: make(chan InternalEvent, 5),
	}, nil
}

func (peer *PeerConnectionState) ShouldSendOffer() bool {
	return peer.IsNewConnection() || peer.IsChanged()
}

func (peer *PeerConnectionState) IsNewConnection() bool {
	return peer.PeerConnection.CurrentLocalDescription() == nil || peer.PeerConnection.PendingLocalDescription() != nil
}

func (peer *PeerConnectionState) IsChanged() bool {
	return peer.PeerConnection.CurrentLocalDescription() == nil || peer.PeerConnection.PendingLocalDescription() != nil
}

func (peer *PeerConnectionState) Close() error {
	if err := peer.PeerConnection.Close(); err != nil {
		return err
	}
	return nil
}

func (peer *PeerConnectionState) AcceptAnswer(data string) error {
	answer := webrtc.SessionDescription{}
	if err := json.Unmarshal([]byte(data), &answer); err != nil {
		return err
	}

	if err := peer.PeerConnection.SetRemoteDescription(answer); err != nil {
		return err
	}
	return nil
}

func (peer *PeerConnectionState) SendOffer() error {
	offer, err := peer.PeerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = peer.PeerConnection.SetLocalDescription(offer); err != nil {
		return err
	}

	offerString, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	peer.HandleEvent("send_offer", string(offerString))
	return nil
}

func (peer *PeerConnectionState) AcceptICECandidate(data string) error {
	candidate := webrtc.ICECandidateInit{}
	if err := json.Unmarshal([]byte(data), &candidate); err != nil {
		return err
	}

	if err := peer.PeerConnection.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (peer *PeerConnectionState) SendICECandidate(i *webrtc.ICECandidate) error {
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

func (peer *PeerConnectionState) HandleEvent(event string, data string) error {
	switch event {
	case "send_offer":
		peer.InternalEventChan <- InternalEvent{Event: "offer", Data: data}
	case "send_candidate":
		peer.InternalEventChan <- InternalEvent{Event: "candidate", Data: data}
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

func (peer *PeerConnectionState) ListenEvents(callback func(string, string)) {
	go func() {
		for {
			select {
			case event := <-peer.InternalEventChan:
				callback(event.Event, event.Data)
			}
		}
	}()
}
