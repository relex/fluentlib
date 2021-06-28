package forwardprotocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/relex/fluentlib/util"
	"github.com/vmihailenco/msgpack/v4"
)

// EventTime represents the custom timestamp type used by Fluentd
type EventTime struct {
	time.Time
}

func init() {
	msgpack.RegisterExt(0, (*EventTime)(nil))
}

// MarshalJSON defines custom JSON marshaling for log record to match its msgpack format (the simplest Forward mode)
func (tm EventTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(util.TimeToUnixFloat(tm.Time))
}

// MarshalMsgpack encodes EventTime in msgpack format
func (tm EventTime) MarshalMsgpack() ([]byte, error) {
	// from https://godoc.org/github.com/vmihailenco/msgpack#example-RegisterExt
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uint32(tm.Unix()))
	binary.BigEndian.PutUint32(b[4:], uint32(tm.Nanosecond()))
	return b, nil
}

// UnmarshalMsgpack decodes EventTime from msgpack bytes
func (tm *EventTime) UnmarshalMsgpack(b []byte) error {
	// from https://godoc.org/github.com/vmihailenco/msgpack#example-RegisterExt
	if len(b) != 8 {
		return fmt.Errorf("invalid data length: got %d, wanted 8", len(b))
	}
	sec := binary.BigEndian.Uint32(b)
	nsec := binary.BigEndian.Uint32(b[4:])
	tm.Time = time.Unix(int64(sec), int64(nsec))
	return nil
}
