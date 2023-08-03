package dump

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/relex/fluentlib/protocol/fluentbitchunk"
	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/relex/fluentlib/util"
	"github.com/relex/gotils/logger"
	"github.com/vmihailenco/msgpack/v4"
)

// PrintChunkFileInJSON dumps all logs in the given file in JSON format. Each log (event) is followed by a newline.
//
// The file must be either a Fluent-bit chunk or a Fluentd forward message
func PrintChunkFileInJSON(path string, indented bool, writer io.Writer) error {
	fileData, ioErr := ioutil.ReadFile(path)
	if ioErr != nil {
		return fmt.Errorf("error opening %s: %w", path, ioErr)
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".flb":
		flbTag, flbPayload, flbErr := fluentbitchunk.ParseChunk(fileData)
		if flbErr != nil {
			return fmt.Errorf("failed to parse fluent-bit chunk file %s: %w", path, flbErr)
		}
		logger.Infof("parsed fluent-bit chunk file: %s, tag=%s", path, flbTag)
		lastI := -1
		iterError := fluentbitchunk.IterateRecords(flbPayload, func(event forwardprotocol.EventEntry) error {
			lastI++
			return PrintEventInJSON(event, flbTag, indented, writer, lastI == 0)
		})
		if iterError != nil {
			return fmt.Errorf("corrupted fluent-bit chunk file %s on the %dth record: %w", path, lastI, iterError)
		}
		return nil
	case ".ff":
		var message forwardprotocol.Message
		if decErr := msgpack.NewDecoder(bytes.NewReader(fileData)).Decode(&message); decErr != nil {
			return fmt.Errorf("failed to decode forward message file %s: %w", path, decErr)
		}
		if prtError := PrintMessageInJSON(message, indented, writer); prtError != nil {
			return fmt.Errorf("failed to print %s: %w", path, prtError)
		}
	default:
		genericReader := bytes.NewReader(fileData)
		genericDecoder := msgpack.NewDecoder(genericReader)
		for {
			nextItem, nextErr := genericDecoder.DecodeInterface()
			decodedPos := genericReader.Size() - int64(genericReader.Len())
			if nextErr != nil {
				if errors.Is(nextErr, io.EOF) {
					break
				}
				return fmt.Errorf("failed to decode msgpack at %s:%d: %w", path, decodedPos, nextErr)
			}
			var jbin []byte
			var jerr error
			if indented {
				jbin, jerr = json.MarshalIndent(nextItem, "", "  ")
			} else {
				jbin, jerr = json.Marshal(nextItem)
			}
			if jerr != nil {
				return fmt.Errorf("failed to format JSON: %w", jerr)
			}
			if _, err := writer.Write(jbin); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			if _, err := writer.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
			logger.Debugf("completed decoding %s at position %d", path, decodedPos)
		}
	}
	return nil
}

// PrintMessageInJSON dumps all logs in the given message in JSON format. Each log (event) is followed by a newline.
func PrintMessageInJSON(message forwardprotocol.Message, indented bool, writer io.Writer) error {
	for i, event := range message.Entries {
		if err := PrintEventInJSON(event, message.Tag, indented, writer, i == 0); err != nil {
			_, _ = writer.Write([]byte("\n]\n")) // ignore error
			return err
		}
	}
	if len(message.Entries) > 0 {
		_, _ = writer.Write([]byte("\n]\n")) // ignore error
	}
	return nil
}

// PrintEventInJSON dump a single record in JSON format
func PrintEventInJSON(event forwardprotocol.EventEntry, tag string, indented bool, writer io.Writer, isFirst bool) error {
	if isFirst {
		if _, werr := writer.Write([]byte("[\n")); werr != nil {
			return fmt.Errorf("failed to print leading bracket: %w", werr)
		}
	} else {
		if _, werr := writer.Write([]byte(",\n")); werr != nil {
			return fmt.Errorf("failed to print leading comma: %w", werr)
		}
	}

	jsonBin, jsonErr := FormatEventInJSON(event, tag, indented)
	if jsonErr != nil {
		return jsonErr
	}
	if _, werr := writer.Write(jsonBin); werr != nil {
		return fmt.Errorf("failed to print JSON: %s: %w", event, werr)
	}
	return nil
}

// FormatEventInJSON formats a single record in EventEntry in JSON format.
func FormatEventInJSON(event forwardprotocol.EventEntry, tag string, indented bool) ([]byte, error) {
	var jsonBin []byte
	var jsonErr error
	if indented {
		jsonBin, jsonErr = json.MarshalIndent([]interface{}{
			tag,
			util.TimeToUnixFloat(event.Time.Time),
			event.Record,
		}, "", "  ")
	} else {
		jsonBin, jsonErr = json.Marshal([]interface{}{
			tag,
			util.TimeToUnixFloat(event.Time.Time),
			event.Record,
		})
	}
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal as JSON: %s: %w", event, jsonErr)
	}
	return jsonBin, nil
}
