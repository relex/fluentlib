package dump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/relex/fluentlib/protocol/fluentbitchunk"
	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/relex/fluentlib/util"
	"github.com/relex/gotils/logger"
	"github.com/vmihailenco/msgpack/v4"
)

var newline = []byte{'\n'}

// PrintChunkFileInJSON dumps all logs in the given file in JSON format
func PrintChunkFileInJSON(path string, indented bool, writer io.Writer) {
	fileData, fileError := ioutil.ReadFile(path)
	if fileError != nil {
		logger.Infof("error opening %s: %s", path, fileError.Error())
		return
	}
	//
	if flbTag, flbPayload, flbError := fluentbitchunk.ParseChunk(fileData); flbError == nil {
		logger.Infof("parsed fluent-bit chunk file: %s, tag=%s", path, flbTag)
		iterError := fluentbitchunk.IterateRecords(flbPayload, func(event forwardprotocol.EventEntry) error {
			return PrintEventInJSON(event, flbTag, indented, writer)
		})
		if iterError != nil {
			logger.Warnf("corrupted fluent-bit chunk file: %s: %s", path, iterError)
		}
		return
	}
	//
	reader := bytes.NewReader(fileData)
	var message forwardprotocol.Message
	if msgError := msgpack.NewDecoder(reader).Decode(&message); msgError != nil {
		logger.Errorf("failed to decode forward message file %s: %s", path, msgError.Error())
	}
	if prtError := PrintMessageInJSON(message, indented, writer); prtError != nil {
		logger.Errorf("failed to print %s: %s", path, prtError.Error())
	}
}

// PrintMessageInJSON dumps all logs in the given message in JSON format
func PrintMessageInJSON(message forwardprotocol.Message, indented bool, writer io.Writer) error {
	for _, event := range message.Entries {
		if err := PrintEventInJSON(event, message.Tag, indented, writer); err != nil {
			return err
		}
	}
	return nil
}

// PrintEventInJSON prints a single record in EventEntry in JSON format. Each event is ended with a newline.
func PrintEventInJSON(event forwardprotocol.EventEntry, tag string, indented bool, writer io.Writer) error {
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
		return fmt.Errorf("failed to marshal as JSON: %s: %w", event, jsonErr)
	}
	if _, werr := writer.Write(jsonBin); werr != nil {
		return fmt.Errorf("failed to print JSON: %s: %w", event, werr)
	}
	if _, werr := writer.Write(newline); werr != nil {
		return fmt.Errorf("failed to print separator: %w", werr)
	}
	return nil
}
