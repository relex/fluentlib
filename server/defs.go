package server

import "time"

var defs = struct {
	ForwarderHandshakeTimeout     time.Duration
	ForwarderBatchSendTimeoutBase time.Duration
	ForwarderBatchAckTimeout      time.Duration
	WriterEndingTimeout           time.Duration
}{
	ForwarderHandshakeTimeout:     10 * time.Second,
	ForwarderBatchSendTimeoutBase: 30 * time.Second,
	ForwarderBatchAckTimeout:      30 * time.Second,
	WriterEndingTimeout:           5 * time.Second,
}
