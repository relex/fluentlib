package receivers

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/relex/fluentlib/dump"
	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/relex/gotils/logger"
)

type splittingFileWriter struct {
	keys          []string
	pathFormat    string
	strict        bool
	connIDToTitle map[int64]string       // connection ID to the title of the latest log event
	titleToOutput map[string]splitOutput // title to file; title is "tag-key1,key2,key3,..."
}

type splitOutput struct {
	file   *os.File
	writer *bufio.Writer
}

func VerifySplittingFilePath(pathFormat string) error {
	testPath := fmt.Sprintf(pathFormat, "hello")
	if strings.Contains(testPath, "%!(EXTRA ") || strings.Contains(testPath, "%!s(MISSING)") {
		return errors.New("must contain exactly one %" + "s: " + pathFormat)
	}
	return nil
}

// NewSplittingFileWriter creates a Receiver which splits output by specified key fields to different files
//
// Each output file is a valid JSON itself, as an array of log events
//
// pathFormat must contain a "%s", which would be replaced by "tag-key1,key2,key3,..."
//
// strict mode means each connection may only send log events of the same key field set, or an error would be logged
func NewSplittingFileWriter(keys []string, pathFormat string, strict bool) Receiver {
	if err := VerifySplittingFilePath(pathFormat); err != nil {
		logger.Panic("pathFormat: ", err.Error())
	}

	return &splittingFileWriter{
		keys:          keys,
		pathFormat:    pathFormat,
		strict:        strict,
		connIDToTitle: make(map[int64]string),
		titleToOutput: make(map[string]splitOutput),
	}
}

func (w *splittingFileWriter) Accept(message ClientMessage) error {
	if len(message.Entries) == 0 {
		return nil
	}

	for _, event := range message.Entries {
		if err := w.acceptEvent(event, message.Tag, message.ConnectionID); err != nil {
			return err
		}
	}

	return nil
}

func (w *splittingFileWriter) Tick() error {
	for _, out := range w.titleToOutput {
		if err := out.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush %s: %w", out.file.Name(), err)
		}
	}
	return nil
}

func (w *splittingFileWriter) End() error {
	for _, out := range w.titleToOutput {
		if _, err := out.writer.Write([]byte("\n]\n")); err != nil {
			return fmt.Errorf("failed to write end mark to %s: %w", out.file.Name(), err)
		}
		if err := out.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush %s: %w", out.file.Name(), err)
		}
		if err := out.file.Close(); err != nil {
			return fmt.Errorf("failed to close %s: %w", out.file.Name(), err)
		}
		logger.Debug("closed ", out.file.Name())
	}
	return nil
}

func (w *splittingFileWriter) acceptEvent(event forwardprotocol.EventEntry, tag string, connID int64) error {
	title := w.makeEventTitle(event, tag)

	isFirst := false
	var output splitOutput

	if out, exists := w.titleToOutput[title]; !exists {
		isFirst = true
		path := fmt.Sprintf(w.pathFormat, title)
		f, ferr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if ferr != nil {
			return fmt.Errorf("failed to create file %s: %w", path, ferr)
		}
		logger.Info("created ", path)
		output = splitOutput{
			file:   f,
			writer: bufio.NewWriter(f),
		}
		w.titleToOutput[title] = output
	} else {
		output = out
	}

	if w.strict {
		if lastConnTitle, exists := w.connIDToTitle[connID]; exists {
			if lastConnTitle != title {
				logger.WithFields(logger.Fields{
					"oldTitle": lastConnTitle,
					"newTitle": title,
					"entry":    event,
				}).Errorf("incoming connection changed tag or key fields")
				w.connIDToTitle[connID] = title
			}
		} else {
			w.connIDToTitle[connID] = title
		}
	}

	return dump.PrintEventInJSON(event, tag, true, output.writer, isFirst)
}

func (w *splittingFileWriter) makeEventTitle(event forwardprotocol.EventEntry, tag string) string {
	if len(w.keys) == 0 {
		return strings.ReplaceAll(tag, "/", "_")
	}

	values := make([]string, len(w.keys))
	for i, k := range w.keys {

		if v, err := event.ResolvePath(strings.Split(k, "/")...); err == nil {
			values[i] = fmt.Sprint(v)
		} else if w.strict {
			logger.WithFields(logger.Fields{
				"tag":   tag,
				"entry": event,
			}).Errorf("missing key field '%s': %v", k, err)
		}
	}
	return strings.ReplaceAll(tag, "/", "_") + "-" + strings.Join(values, ",")
}
