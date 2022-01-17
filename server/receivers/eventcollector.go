package receivers

import (
	"errors"
	"time"

	"github.com/relex/fluentlib/protocol/forwardprotocol"
)

type eventCollector struct {
	ch      chan forwardprotocol.EventEntry
	timeout time.Duration
}

// NewEventCollector creates a Receiver which sends every log events to the returned channel
func NewEventCollector(timeout time.Duration) (Receiver, chan forwardprotocol.EventEntry) {
	ch := make(chan forwardprotocol.EventEntry, 100)

	return &eventCollector{ch, timeout}, ch
}

func (w *eventCollector) Accept(message ClientMessage) error {
	t := time.After(w.timeout)

	for _, evt := range message.Entries {
		select {
		case w.ch <- evt:
			// ok
		case <-t:
			return errors.New("timeout writing to event channel")
		}
	}
	return nil
}

func (w *eventCollector) Tick() error {
	return nil
}

func (w *eventCollector) End() error {
	close(w.ch)
	return nil
}
