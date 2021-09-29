package server

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/relex/fluentlib/dump"
	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/relex/fluentlib/testdata"
	"github.com/relex/fluentlib/util"
	"github.com/relex/gotils/logger"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v4"
)

func TestServerBasic(t *testing.T) {
	recv, ch := NewEventCollector(5 * time.Second)
	srv, srvAddr := LaunchServer(logger.WithField("test", t.Name()), Config{
		Address: "localhost:0",
		Secret:  "hi",
		TLS:     true,
	}, recv)

	conn, connErr := openConn(srvAddr.String(), "hi")
	assert.Nil(t, connErr)

	request := forwardprotocol.Message{
		Tag: "hello",
		Entries: []forwardprotocol.EventEntry{
			{
				Time: forwardprotocol.EventTime{Time: time.Date(2020, 10, 31, 1, 2, 3, 4, time.UTC)},
				Record: map[string]interface{}{
					"field1": "foo",
					"field2": "bar",
				},
			},
			{
				Time: forwardprotocol.EventTime{Time: time.Date(2020, 11, 31, 1, 2, 3, 4, time.UTC)},
				Record: map[string]interface{}{
					"field1": "FOO",
					"field2": "BAR",
				},
			},
		},
		Option: forwardprotocol.TransportOption{Chunk: "first"},
	}
	encoder := msgpack.NewEncoder(conn)
	assert.Nil(t, encoder.Encode(request))

	decoder := msgpack.NewDecoder(conn)
	var response forwardprotocol.Ack
	assert.Nil(t, decoder.Decode(&response))
	assert.Equal(t, request.Option.Chunk, response.Ack)

	msg1 := <-ch
	assert.Equal(t, request.Entries[0].Time.Format(time.RFC3339Nano), msg1.Time.UTC().Format(time.RFC3339Nano))
	assert.Equal(t, request.Entries[0].Record, msg1.Record)

	msg2 := <-ch
	assert.Equal(t, request.Entries[1].Time.Format(time.RFC3339Nano), msg2.Time.UTC().Format(time.RFC3339Nano))
	assert.Equal(t, request.Entries[1].Record, msg2.Record)

	srv.Shutdown()
}

func TestServerFailureEmulation(t *testing.T) {
	if util.IsTestGenerationMode() {
		return
	}
	recv, ch := NewMessageCollector(5 * time.Second)
	srv, srvAddr := LaunchServer(logger.WithField("test", t.Name()), Config{
		Address:          "localhost:0",
		Secret:           "hi",
		TLS:              true,
		RandomFailAuth:   0.6,
		RandomKillConn:   0.2,
		RandomNoResponse: 0.0, // timeout would block tests for too long
	}, recv)

	var conn net.Conn

	for _, fn := range testdata.ListInputFiles(t) {
		sampleInput, sampleErr := ioutil.ReadFile(fn)
		assert.Nil(t, sampleErr, fn)

		assert.Nil(t, send(&conn, srvAddr.String(), "hi", sampleInput), fn)

		expectedFn := testdata.GetOutputFilename(t, fn)
		expected, readErr := ioutil.ReadFile(expectedFn)
		assert.Nil(t, readErr, expectedFn)

		msg := <-ch
		wrt := &bytes.Buffer{}
		assert.Nil(t, dump.PrintMessageInJSON(msg, true, wrt))
		assert.Equal(t, string(expected), wrt.String())
	}

	if conn != nil {
		conn.Close()
	}

	srv.Shutdown()
}

func send(connHolder *net.Conn, addr string, secret string, data []byte) error {
	const retryLimit = 10
	retry := 0

	for {
		if *connHolder == nil {
			for {
				conn, connErr := openConn(addr, secret)
				if connErr == nil {
					*connHolder = conn
					break
				}
				if retry >= retryLimit {
					return connErr
				}

				logger.Warn("failed to connect: ", connErr)
				retry++
			}
		}

		_ = (*connHolder).SetWriteDeadline(time.Now().Add(5 * time.Second)) // ignore error
		_, wrtErr := (*connHolder).Write(data)
		if wrtErr != nil {
			if retry >= retryLimit {
				return wrtErr
			}
			logger.Warn("failed to send: ", wrtErr)
			(*connHolder).Close()
			(*connHolder) = nil
			retry++
			continue
		}

		var response forwardprotocol.Ack
		decoder := msgpack.NewDecoder(*connHolder)
		rspErr := decoder.Decode(&response)
		if rspErr != nil {
			if retry >= retryLimit {
				return rspErr
			}
			logger.Warn("failed to receive: ", rspErr)
			(*connHolder).Close()
			(*connHolder) = nil
			retry++
			continue
		}
		logger.Info("received ACK: ", response.Ack)

		return nil
	}
}

func openConn(addr string, secret string) (net.Conn, error) {
	conn, connErr := net.Dial("tcp", addr)
	if connErr != nil {
		return nil, connErr
	}

	conn = tls.Client(conn, &tls.Config{InsecureSkipVerify: true})

	success, reason, netErr := forwardprotocol.DoClientHandshake(conn, secret, 5*time.Second)
	if !success {
		conn.Close()
		if netErr != nil {
			return nil, netErr
		}
		return nil, errors.New(reason)
	}

	return conn, nil
}
