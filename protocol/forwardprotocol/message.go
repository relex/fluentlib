package forwardprotocol

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

// Message is the request msg to forward a chunk of logs
// The struct is not used directly for encoding but serves as a reference
type Message struct {
	_msgpack struct{}        `msgpack:",asArray"`
	Tag      string          `msgpack:"tag"`
	Entries  []EventEntry    `msgpack:"entries"` // Depending on MessageMode, the entries may be serialized as-is or in other formats
	Option   TransportOption `msgpack:"option"`
}

// EventEntry represents a single log record in forward messages
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
