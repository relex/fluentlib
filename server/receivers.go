package server

import (
	"bufio"
	"errors"
	"io"
	"time"

	"github.com/relex/fluentlib/dump"
	"github.com/relex/fluentlib/protocol/forwardprotocol"
)

// MessageReceiver is the interface to receive decoded messages from ForwardServer
//
// Each of the messages may contain one or more log events
//
// The Tick is called periodically and may be used to flush I/O
type MessageReceiver interface {
	Accept(forwardprotocol.Message) error
	Tick() error
	End() error
}

type messageWriter struct {
	wrt *bufio.Writer
}

// NewMessageWriter creates a MessageReceiver which prints all logs in JSON format to the given writer
//
// Each of logs is terminated by a newline (no valid JSON separator)
func NewMessageWriter(wrt io.Writer) MessageReceiver {
	return &messageWriter{bufio.NewWriter(wrt)}
}

func (w *messageWriter) Accept(message forwardprotocol.Message) error {
	return dump.PrintMessageInJSON(message, false, w.wrt)
}

func (w *messageWriter) Tick() error {
	return w.wrt.Flush()
}

func (w *messageWriter) End() error {
	return w.wrt.Flush()
}

type messageCollector struct {
	ch      chan forwardprotocol.Message
	timeout time.Duration
}

// NewMessageCollector creates a MessageReceiver which sends every messages to the returned channel
func NewMessageCollector(timeout time.Duration) (MessageReceiver, chan forwardprotocol.Message) {
	ch := make(chan forwardprotocol.Message, 100)

	return &messageCollector{ch, timeout}, ch
}

func (w *messageCollector) Accept(message forwardprotocol.Message) error {
	select {
	case w.ch <- message:
		return nil
	case <-time.After(w.timeout):
		return errors.New("timeout writing to message channel")
	}
}

func (w *messageCollector) Tick() error {
	return nil
}

func (w *messageCollector) End() error {
	close(w.ch)
	return nil
}

type eventCollector struct {
	ch      chan forwardprotocol.EventEntry
	timeout time.Duration
}

// NewEventCollector creates a MessageReceiver which sends every log events to the returned channel
func NewEventCollector(timeout time.Duration) (MessageReceiver, chan forwardprotocol.EventEntry) {
	ch := make(chan forwardprotocol.EventEntry, 100)

	return &eventCollector{ch, timeout}, ch
}

func (w *eventCollector) Accept(message forwardprotocol.Message) error {
	t := time.After(w.timeout)

	for _, evt := range message.Entries {
		select {
		case w.ch <- evt:
			// ok
		case <-t:
			return errors.New("timeout writing to event channel")
		}
	}
	return nil
}

func (w *eventCollector) Tick() error {
	return nil
}

func (w *eventCollector) End() error {
	close(w.ch)
	return nil
}
