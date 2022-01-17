package receivers

import (
	"errors"
	"time"

	"github.com/relex/fluentlib/protocol/forwardprotocol"
)

type messageCollector struct {
	ch      chan forwardprotocol.Message
	timeout time.Duration
}

// NewMessageCollector creates a Receiver which sends every messages to the returned channel
func NewMessageCollector(timeout time.Duration) (Receiver, chan forwardprotocol.Message) {
	ch := make(chan forwardprotocol.Message, 100)

	return &messageCollector{ch, timeout}, ch
}

func (w *messageCollector) Accept(message ClientMessage) error {
	select {
	case w.ch <- message.Message:
		return nil
	case <-time.After(w.timeout):
		return errors.New("timeout writing to message channel")
	}
}

func (w *messageCollector) Tick() error {
	return nil
}

func (w *messageCollector) End() error {
	close(w.ch)
	return nil
}
