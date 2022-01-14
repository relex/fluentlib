package receivers

import (
	"bufio"
	"io"

	"github.com/relex/fluentlib/dump"
)

type messageWriter struct {
	wrt *bufio.Writer
}

// NewMessageWriter creates a Receiver which prints all logs in JSON format to the given writer
//
// Each of logs is terminated by a newline (no valid JSON separator)
func NewMessageWriter(wrt io.Writer) Receiver {
	return &messageWriter{bufio.NewWriter(wrt)}
}

func (w *messageWriter) Accept(message ClientMessage) error {
	return dump.PrintMessageInJSON(message.Message, false, w.wrt)
}

func (w *messageWriter) Tick() error {
	return w.wrt.Flush()
}

func (w *messageWriter) End() error {
	return w.wrt.Flush()
}
