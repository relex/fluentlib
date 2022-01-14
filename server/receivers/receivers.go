// Package receivers contains the interface and implementations to receive results from fake server
package receivers

import (
	"github.com/relex/fluentlib/protocol/forwardprotocol"
)

// Receiver is the interface to receive decoded Fluentd forward messages from ForwardServer
//
// Each message may contain one or more log events
//
// Receiver is used from a single goroutine only
type Receiver interface {

	// Accept accepts a new message from client
	Accept(ClientMessage) error

	// Tick is called periodically and may be used to flush I/O
	Tick() error

	// End is called when the server is shut down
	End() error
}

// ClientMessage represents a Fluentd forward message received from a client
type ClientMessage struct {
	ConnectionID int64
	forwardprotocol.Message
}
