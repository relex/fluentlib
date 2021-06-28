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

// PrintFromFileToJSON dumps all logs in the given file in JSON format
func PrintFromFileToJSON(path string, separator []byte, indented bool, writer io.Writer) {
	fileData, fileError := ioutil.ReadFile(path)
	if fileError != nil {
		logger.Infof("error opening %s: %s", path, fileError.Error())
		return
	}
	//
	if flbTag, flbPayload, flbError := fluentbitchunk.ParseChunk(fileData); flbError == nil {
		logger.Infof("parsed fluent-bit chunk file: %s, tag=%s", path, flbTag)
		iter := 0
		iterError := fluentbitchunk.IterateRecords(flbPayload, func(event forwardprotocol.EventEntry) error {
			if iter > 0 {
				if _, err := writer.Write(separator); err != nil {
					return fmt.Errorf("failed to print separator: %w", err)
				}

			}
			iter++
			return PrintEventAsJSON(event, flbTag, indented, writer)
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
	if prtError := PrintFromMessageToJSON(message, separator, indented, writer); prtError != nil {
		logger.Errorf("failed to print %s: %s", path, prtError.Error())
	}
}

// PrintFromMessageToJSON dumps all logs in the given message in JSON format
func PrintFromMessageToJSON(message forwardprotocol.Message, separator []byte, indented bool, writer io.Writer) error {
	for iter, event := range message.Entries {
		if iter > 0 {
			if _, err := writer.Write(separator); err != nil {
				return fmt.Errorf("failed to print separator: %w", err)
			}
		}
		if err := PrintEventAsJSON(event, message.Tag, indented, writer); err != nil {
			return err
		}
	}
	return nil
}

// PrintEventAsJSON prints a single record in EventEntry in JSON format
func PrintEventAsJSON(event forwardprotocol.EventEntry, tag string, indented bool, writer io.Writer) error {
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
	return nil
}
