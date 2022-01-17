package server

import (
	"time"

	"github.com/relex/fluentlib/server/receivers"
	"github.com/relex/gotils/channels"
	"github.com/relex/gotils/logger"
)

func launchWriter(wlogger logger.Logger, receiver receivers.Receiver) (chan<- receivers.ClientMessage, channels.Awaitable) {
	outputChan := make(chan receivers.ClientMessage, 1000)
	endsignal := channels.NewSignalAwaitable()

	go func() {
		defer endsignal.Signal()

		numMessage := 0
		ticker := time.NewTicker(500 * time.Millisecond)

	RECEIVE_LOOP:
		for {
			select {
			case message, ok := <-outputChan:
				if !ok {
					break RECEIVE_LOOP
				}
				numMessage++
				if err := receiver.Accept(message); err != nil {
					wlogger.Fatalf("failed to accept message: %v", err)
				}
			case <-ticker.C:
				if err := receiver.Tick(); err != nil {
					wlogger.Fatalf("failed to tick: %v", err)
				}
			}
		}

		if err := receiver.End(); err != nil {
			wlogger.Fatalf("failed to close receiver: %v", err)
		}
		wlogger.Infof("written %d log records", numMessage)
	}()

	return outputChan, endsignal
}
