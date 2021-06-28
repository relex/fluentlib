// Package fluentbitchunk provides parsing and creation of fluent-bit chunk files
package fluentbitchunk

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/vmihailenco/msgpack/v4"
)

/*
 *    +--------------+----------------+
 *    |     0xC1     |     0x00       + <<<<<<<<<<<<<<<<<<< Ident1, Ident2
 *    +--------------+----------------+
 *    |       4 BYTES CRC32           | <<<<<<<<<<<<<<<<<<< CRC of everything after padding
 *    |      16 BYTES Padding         |
 *    +-------------------------------| <<<<<<<<<<<<<<<<<<< TagLengthStart
 *    |          Tag length           |
 *    |         (big endian)          |
 *    +-------------------------------+ <<<<<<<<<<<<<<<<<<< TagValueStart
 *    |                               |
 *    |          Tag value            | length = tag length
 *    |                               |
 *    +-------------------------------+
 *    |                               |
 *    |   sequence of msgpack arrays  |
 *    |                               |
 *    +-------------------------------+
 */

const (
	ident1         = 0xC1
	ident2         = 0x00
	tagLengthStart = 22
	tagValueStart  = 24
)

// MakeHeader creates a file header without CRC for testing
func MakeHeader(tag string) []byte {
	btag := []byte(tag)
	if len(btag) > 65535 {
		panic(fmt.Sprintf("input tag is too long: len=%d", len(btag)))
	}
	header := make([]byte, tagValueStart+len(btag))
	header[0] = ident1
	header[1] = ident2
	header[tagLengthStart] = byte(len(btag) >> 8)
	header[tagLengthStart+1] = byte(len(btag) & 0xFF)
	copy(header[tagValueStart:], btag)
	return header
}

// ParseChunk parses the .flb file content or returns error
// Returns (tag, payload, error)
func ParseChunk(data []byte) (string, []byte, error) {
	if len(data) < tagValueStart {
		return "", nil, errors.New(".flb chunk is too small")
	}
	if data[0] != ident1 || data[1] != ident2 {
		return "", nil, errors.New(".flb chunk header is incorrect")
	}
	tagLen := (uint16(data[tagLengthStart]) << 8) | uint16(data[tagLengthStart+1])
	tag := string(data[tagValueStart : tagValueStart+tagLen])
	payload := data[tagValueStart+tagLen:]
	return tag, payload, nil
}

// IterateRecords iterates through all records in fluent-bit chunk payload
func IterateRecords(payload []byte, callback func(event forwardprotocol.EventEntry) error) error {
	buffer := bytes.NewBuffer(payload) // the real underlying reader of data, as bytes.Buffer
	decoder := msgpack.NewDecoder(buffer)
	eventStartPos := 0
	numEvents := 0
	for {
		entry := forwardprotocol.EventEntry{}
		decodeErr := decoder.Decode(&entry)
		// calculate how many bytes have been decoded for this record
		eventEndPos := len(payload) - buffer.Len()
		if decodeErr != nil {
			if errors.Is(decodeErr, io.EOF) {
				return nil
			}
			return fmt.Errorf("record %d at %d-%d/%d: %w", numEvents, eventStartPos, eventEndPos, len(payload), decodeErr)
		}
		// garbage and all NULs might happen when fluent-bit crashes
		if eventEndPos-eventStartPos < 12 {
			return fmt.Errorf("record %d at %d-%d/%d: too small", numEvents, eventStartPos, eventEndPos, len(payload))
		}
		if err := callback(entry); err != nil {
			return fmt.Errorf("record %d callback: %w", numEvents, err)
		}
		eventStartPos = eventEndPos
		numEvents++
	}
}
