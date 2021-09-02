package server

import (
	"bufio"
	"errors"
	"io"
	"time"

	"github.com/relex/fluentlib/dump"
	"github.com/relex/fluentlib/protocol/forwardprotocol"
)

type MessageReceiver interface {
	Accept(forwardprotocol.Message) error
	Tick() error
	End() error
}

type MessageWriter struct {
	wrt *bufio.Writer
}

func NewMessageWriter(wrt io.Writer) MessageReceiver {
	return &MessageWriter{bufio.NewWriter(wrt)}
}

func (w *MessageWriter) Accept(message forwardprotocol.Message) error {
	return dump.PrintMessageInJSON(message, false, w.wrt)
}

func (w *MessageWriter) Tick() error {
	return w.wrt.Flush()
}

func (w *MessageWriter) End() error {
	return w.wrt.Flush()
}

type MessageCollector struct {
	ch      chan forwardprotocol.Message
	timeout time.Duration
}

func NewMessageCollector(timeout time.Duration) (MessageReceiver, chan forwardprotocol.Message) {
	ch := make(chan forwardprotocol.Message, 100)

	return &MessageCollector{ch, timeout}, ch
}

func (w *MessageCollector) Accept(message forwardprotocol.Message) error {
	select {
	case w.ch <- message:
		return nil
	case <-time.After(w.timeout):
		return errors.New("timeout writing to message channel")
	}
}

func (w *MessageCollector) Tick() error {
	return nil
}

func (w *MessageCollector) End() error {
	close(w.ch)
	return nil
}

type EventCollector struct {
	ch      chan forwardprotocol.EventEntry
	timeout time.Duration
}

func NewEventCollector(timeout time.Duration) (MessageReceiver, chan forwardprotocol.EventEntry) {
	ch := make(chan forwardprotocol.EventEntry, 100)

	return &EventCollector{ch, timeout}, ch
}

func (w *EventCollector) Accept(message forwardprotocol.Message) error {
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

func (w *EventCollector) Tick() error {
	return nil
}

func (w *EventCollector) End() error {
	close(w.ch)
	return nil
}
