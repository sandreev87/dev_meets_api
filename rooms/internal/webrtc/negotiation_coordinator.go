package webrtc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/pion/webrtc/v4"
)

const (
	OfferEvent         = "offer"
	CandidateEvent     = "candidate"
	AnswerEvent        = "answer"
	SendOfferEvent     = "send_offer"
	SendCandidateEvent = "send_candidate"
	SendAnswerEvent    = "send_answer"
)

var UndefinedEventError = errors.New("undefined event")

type NegotiationCoordinator struct {
	peerConnection    *webrtc.PeerConnection
	internalEventChan chan InternalEvent
}

type InternalEvent struct {
	Event string
	Data  string
}

func NewNegotiationCoordinator(peerConnection *webrtc.PeerConnection) *NegotiationCoordinator {
	return &NegotiationCoordinator{
		peerConnection:    peerConnection,
		internalEventChan: make(chan InternalEvent, 20),
	}
}

func (c *NegotiationCoordinator) AcceptAnswer(data string) error {
	answer := webrtc.SessionDescription{}
	if err := json.Unmarshal([]byte(data), &answer); err != nil {
		return err
	}

	if err := c.peerConnection.SetRemoteDescription(answer); err != nil {
		return err
	}
	return nil
}

func (c *NegotiationCoordinator) SendAnswer() error {
	answer, err := c.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := c.peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	answerString, err := json.Marshal(answer)
	if err := c.HandleEvent(SendAnswerEvent, string(answerString)); err != nil {
		return err
	}
	return nil
}

func (c *NegotiationCoordinator) SendOffer() error {
	offer, err := c.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = c.peerConnection.SetLocalDescription(offer); err != nil {
		return err
	}

	offerString, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	if err := c.HandleEvent(SendOfferEvent, string(offerString)); err != nil {
		return err
	}

	return nil
}

func (c *NegotiationCoordinator) AcceptOffer(data string) error {
	offer := webrtc.SessionDescription{}

	if err := json.Unmarshal([]byte(data), &offer); err != nil {
		return err
	}

	if err := c.peerConnection.SetRemoteDescription(offer); err != nil {
		return err
	}

	if err := c.SendAnswer(); err != nil {
		return err
	}
	return nil
}

func (c *NegotiationCoordinator) AcceptICECandidate(data string) error {
	candidate := webrtc.ICECandidateInit{}
	if err := json.Unmarshal([]byte(data), &candidate); err != nil {
		return err
	}

	if err := c.peerConnection.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (c *NegotiationCoordinator) SendICECandidate(i *webrtc.ICECandidate) error {
	if i == nil {
		return nil
	}

	candidateString, err := json.Marshal(i.ToJSON())
	if err != nil {
		return err
	}
	if err := c.HandleEvent(SendCandidateEvent, string(candidateString)); err != nil {
		return err
	}

	return nil
}

func (c *NegotiationCoordinator) HandleEvent(event string, data string) error {
	switch event {
	case SendOfferEvent:
		c.internalEventChan <- InternalEvent{Event: "offer", Data: data}
	case SendAnswerEvent:
		c.internalEventChan <- InternalEvent{Event: "answer", Data: data}
	case SendCandidateEvent:
		c.internalEventChan <- InternalEvent{Event: "candidate", Data: data}
	case OfferEvent:
		if err := c.AcceptOffer(data); err != nil {
			return err
		}
	case CandidateEvent:
		if err := c.AcceptICECandidate(data); err != nil {
			return err
		}
	case AnswerEvent:
		if err := c.AcceptAnswer(data); err != nil {
			return err
		}
	default:
		return UndefinedEventError
	}
	return nil
}

func (c *NegotiationCoordinator) ListenSendEvents(ctx context.Context, callback func(string, string)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(c.internalEventChan)
				return
			case event := <-c.internalEventChan:
				callback(event.Event, event.Data)
			}
		}
	}()
}
