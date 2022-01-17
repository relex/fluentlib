package forwardprotocol

import (
	"fmt"
	"reflect"
)

// MessageMode determines the format in which Message.Entries are serialized
// The mode is to be detected by upstream, not itself specified during communication
type MessageMode string

const (
	// ModeForward serializes logs as a msgpack array, the original and fluent-bit compatible format
	ModeForward MessageMode = "Forward"

	// ModePackedForward packs serialized logs as a msgpack binary (double msgpack)
	ModePackedForward MessageMode = "PackedForward"

	// ModeCompressedPackedForward packs gzipped and serialized logs as a msgpack binary (double msgpack)
	// In production this should always be used because the saving of space and network bandwidth is 20-50x
	ModeCompressedPackedForward MessageMode = "CompressedPackedForward"
)

// Message represents a request to forward a batch of log events to Fluentd
//
// The struct is not used directly for encoding but serves as a reference
type Message struct {
	_msgpack struct{}        `msgpack:",asArray"`
	Tag      string          `msgpack:"tag"`
	Entries  []EventEntry    `msgpack:"entries"` // Depending on MessageMode, the entries may be serialized as-is or in other formats
	Option   TransportOption `msgpack:"option"`
}

// EventEntry represents a single log record in forward messages
//
// The struct is not used directly for encoding but serves as a reference
type EventEntry struct {
	_msgpack struct{}               `msgpack:",asArray"`
	Time     EventTime              `msgpack:"time"`
	Record   map[string]interface{} `msgpack:"record"`
}

// TransportOption is the option of each transport request (last value of array)
type TransportOption struct {
	_msgpack   struct{} `msgpack:",omitempty"`
	Size       int      `msgpack:"size" json:"size"`             // The numbers of log records in this msg
	Chunk      string   `msgpack:"chunk" json:"chunk"`           // Chunk ID, omitted if a response from server as ACK is not needed
	Compressed string   `msgpack:"compressed" json:"compressed"` // set to ForwardCompressionFormat for "CompressedPackedForward" mode
}

// Ack is the acknowledgement or response from server to client for receiving a chunk
type Ack struct {
	Ack string `msgpack:"ack"` // equals to ForwardTransportOption.Chunk
}

// CompressionFormat defines the compression format, only "gzip" is supported
const CompressionFormat = "gzip"

func init() {
	unusedStruct(Message{}._msgpack)
	unusedStruct(EventEntry{}._msgpack)
	unusedStruct(TransportOption{}._msgpack)
}

// ResolvePath attempts to fetch field by the given path, represented as one or more map keys
//
// Supports nested maps - the type should be map[string]interface{}
func (e *EventEntry) ResolvePath(path ...string) (interface{}, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path must contain at least one step")
	}

	parent := e.Record
	for step := range path {
		value, exists := parent[path[step]]
		if !exists {
			return nil, fmt.Errorf("failed to resolve %v at step %d: '%s' does not exist", path, step+1, path[step])
		}

		if step == len(path)-1 {
			return value, nil
		}

		if m, isMap := value.(map[string]interface{}); isMap {
			parent = m
			continue
		}
		return nil, fmt.Errorf("failed to resolve %v at step %d: '%s' is not a map[string]interface{}: type=%s value=%v",
			path, step+1, path[step], reflect.ValueOf(value).Type().Name(), value)
	}

	panic("unreachable code")
}
