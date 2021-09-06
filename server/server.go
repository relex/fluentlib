// Package server provides a simple test server for Fluentd "Forward" protocol, which dumps every records received in JSON format
package server

import (
	"bufio"
	"crypto/tls"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/relex/gotils/logger"
	"github.com/vmihailenco/msgpack/v4"
)

// ForwardServer is a listener for Fluentd Forward protocol for testing
type ForwardServer struct {
	logger   logger.Logger
	config   Config
	receiver MessageReceiver
	listener net.Listener
	connMap  *sync.Map
}

// Config contains configuration for test server
type Config struct {
	Address        string  `help:"Address to listen requests"`
	Secret         string  `help:"The secret key for communication with clients if TLS is enabled"`
	TLS            bool    `help:"Enable TLS or not"`
	RandomAuthFail float64 `help:"Chance to fail authentication, from 0.0 to 1.0"`
	RandomConnKill float64 `help:"Chance to kill connection after receiving a request, from 0.0 to 1.0"`
	RandomNoAnswer float64 `help:"Chance to stop responding after receiving a request (but continue to receive logs)"`
}

// LaunchServer creates a new server and launches it in background
func LaunchServer(parentLogger logger.Logger, config Config, receiver MessageReceiver) (*ForwardServer, net.Addr) {

	slogger := parentLogger.WithField("component", "FluentdForwardTestServer")
	lsnr, err := net.Listen("tcp", config.Address)
	if err != nil {
		slogger.Panic("listen: ", err)
	}
	slogger.Infof("listening to %s", lsnr.Addr())
	server := &ForwardServer{
		logger:   slogger,
		config:   config,
		receiver: receiver,
		listener: lsnr,
		connMap:  new(sync.Map),
	}
	go server.run()
	return server, lsnr.Addr()
}

// Shutdown aborts the server
func (server *ForwardServer) Shutdown() {
	server.listener.Close()
	server.connMap.Range(func(rawAddr interface{}, rawConn interface{}) bool {
		addr := rawAddr.(string)
		conn := rawConn.(net.Conn)
		server.logger.Infof("force closing connection from %s", addr)
		conn.Close()
		return true
	})
}

func (server *ForwardServer) run() {
	outputChan := make(chan forwardprotocol.Message, 1000)
	go server.runWriter(outputChan)
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			server.logger.Info("listener stopped: ", err)
			return
		}
		server.logger.Info("accepted connection from ", conn.RemoteAddr())
		go server.runConn(conn, outputChan)
	}
}

func (server *ForwardServer) runWriter(outputChan <-chan forwardprotocol.Message) {
	numMessage := 0
	ticker := time.NewTicker(500 * time.Millisecond)

RECEIVE_LOOP:
	for {
		select {
		case message, ok := <-outputChan:
			if !ok {
				break RECEIVE_LOOP
			}
			numMessage++
			if err := server.receiver.Accept(message); err != nil {
				server.logger.Fatalf("failed to accept message: %v", err)
			}
		case <-ticker.C:
			if err := server.receiver.Tick(); err != nil {
				server.logger.Fatalf("failed to tick: %v", err)
			}
		}
	}

	if err := server.receiver.End(); err != nil {
		server.logger.Fatalf("failed to close receiver: %v", err)
	}
	server.logger.Infof("written %d log records", numMessage)
}

func (server *ForwardServer) runConn(conn net.Conn, outputChan chan<- forwardprotocol.Message) {
	addr := conn.RemoteAddr().String()
	clogger := server.logger.WithField("remote", conn.RemoteAddr())
	defer conn.Close()
	server.connMap.Store(addr, conn)
	defer server.connMap.Delete(addr)

	if server.config.TLS {
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = []tls.Certificate{
			makeTestServerCertificate(),
		}
		conn = tls.Server(conn, tlsConfig)
		clogger.Info("added TLS to connection ", conn.RemoteAddr())
		defer conn.Close()
	}

	authSuccess, err := forwardprotocol.DoServerHandshake(conn, server.config.Secret, defs.ForwarderHandshakeTimeout, server.onAuth)
	if err != nil {
		clogger.Warn("handshake error: ", err)
		return
	}
	if !authSuccess {
		clogger.Warn("client auth failed")
		return
	}
	clogger.Debug("handshaked")

	ackChannel := make(chan string, 1000)
	defer close(ackChannel)
	go server.runAcknowledger(ackChannel, conn, clogger)

	decoder := msgpack.NewDecoder(conn)
	stopAck := false
	for {
		if err := conn.SetReadDeadline(time.Now().Add(defs.ForwarderBatchSendTimeoutBase)); err != nil {
			clogger.Error("unable to set read timeout: ", err)
			return
		}
		var message forwardprotocol.Message
		if err := decoder.Decode(&message); err != nil {
			clogger.Error("unable to read: ", err)
			return
		}
		if r := rand.Float64(); r < server.config.RandomConnKill {
			clogger.Infof("kill connection (random %f)", r)
			return
		}
		clogger.Debugf("received msg: tag=%s, entries=%d, chunkID=%s", message.Tag, len(message.Entries), message.Option.Chunk)
		outputChan <- message
		if stopAck {
			continue
		}
		if len(message.Option.Chunk) > 0 {
			ackChannel <- message.Option.Chunk
		}
		if r := rand.Float64(); r < server.config.RandomNoAnswer {
			// simulate invalid server response to client
			clogger.Infof("stop server to client response (random %f)", r)
			stopAck = true
		}
	}
}

func (server *ForwardServer) runAcknowledger(ackChannel chan string, conn net.Conn, clogger logger.Logger) {
	alogger := clogger.WithField("part", "acknowledger")
	cwriter := bufio.NewWriter(conn)
	encoder := msgpack.NewEncoder(cwriter)
	for chunkID := range ackChannel {
		ack := forwardprotocol.Ack{
			Ack: chunkID,
		}
		if err := conn.SetWriteDeadline(time.Now().Add(defs.ForwarderBatchAckTimeout)); err != nil {
			alogger.Error("unable to set write timeout: ", err)
			return
		}
		if err := encoder.Encode(&ack); err != nil {
			alogger.Error("unable to ack: ", err)
			return
		}
		if err := cwriter.Flush(); err != nil {
			alogger.Error("unable to ack: ", err)
			return
		}
	}
	alogger.Infof("end")
}

func (server *ForwardServer) onAuth(hostname, username, password string) (bool, string) {
	if r := rand.Float64(); r < server.config.RandomAuthFail {
		logger.Infof("reject client auth  (random %f)", r)
		return false, "bad luck"
	}
	return true, ""
}

func makeTestServerCertificate() tls.Certificate {
	// certificate from https://golang.org/pkg/crypto/tls/#X509KeyPair example
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)
	keyPem := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		panic(err)
	}
	return cert
}
