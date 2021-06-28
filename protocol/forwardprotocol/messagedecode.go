package forwardprotocol

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack/v4"
	"github.com/vmihailenco/msgpack/v4/codes"
)

var _ msgpack.CustomDecoder = (*Message)(nil)

// DecodeMsgpack is the custom msgpack decoding implementation for Message, in order to decode Entries properly
//
// See MessageMode for different types of Entries encoding
func (msg *Message) DecodeMsgpack(decoder *msgpack.Decoder) error {
	// first is array length; should be 3
	{
		len, err := decoder.DecodeArrayLen()
		if err != nil {
			return fmt.Errorf("message's field count: %w", err)
		}
		if len != 3 {
			return fmt.Errorf("message's field count: %d (should be 3)", len)
		}

	}
	// array[0] is tag
	{
		tag, err := decoder.DecodeString()
		if err != nil {
			return fmt.Errorf("message's tag: %w", err)
		}
		msg.Tag = tag
	}
	// array[1] is array of entries or binary
	var maybeEntriesBinary []byte
	{
		code, cerr := decoder.PeekCode()
		if cerr != nil {
			return fmt.Errorf("message's entries code: %w", cerr)
		}
		if !codes.IsBin(code) {
			if err := decoder.Decode(&msg.Entries); err != nil {
				return fmt.Errorf("message's entries as array of logs: %w", err)
			}
		} else if err := decoder.Decode(&maybeEntriesBinary); err != nil {
			return fmt.Errorf("message's entries as binary: %w", err)
		}
	}
	// array[2] is option
	if err := decoder.Decode(&msg.Option); err != nil {
		return fmt.Errorf("message's option map: %w", err)
	}
	// then decode bin if present
	if maybeEntriesBinary != nil {
		compressed := msg.Option.Compressed != ""
		entries, err := decodePackedEntriesStream(maybeEntriesBinary, compressed, msg.Option.Size)
		if err != nil {
			return fmt.Errorf("message's entries binary (compressed=%t): %w", compressed, err)
		}
		msg.Entries = entries
	}
	return nil
}

func decodePackedEntriesStream(v []byte, compressed bool, size int) ([]EventEntry, error) {
	var reader io.Reader = bytes.NewReader(v)
	if compressed {
		zreader, zerr := gzip.NewReader(reader)
		if zerr != nil {
			return nil, zerr
		}
		reader = zreader
	}
	decoder := msgpack.NewDecoder(reader)
	list := make([]EventEntry, 0, size)
	for {
		var record EventEntry
		if err := decoder.Decode(&record); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return list, err
		}
		list = append(list, record)
	}
	return list, nil
}
