package forwardprotocol

import (
	"bufio"
	"errors"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

// AuthCallback is the definition of callback for (test) server to authenticate a client
//
// Returns (success?, reason)
type AuthCallback func(hostname, username, password string) (bool, string)

// DoServerHandshake performs server-side handshake on the given forward protocol connection.
//
// Returns (success?, network error)
func DoServerHandshake(conn net.Conn, sharedKey string, timeout time.Duration, auth AuthCallback) (bool, error) {
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return false, err
	}
	decoder := msgpack.NewDecoder(conn)
	bwriter := bufio.NewWriterSize(conn, 1024)
	encoder := msgpack.NewEncoder(bwriter)

	// send HELO
	nonce := strconv.Itoa(rand.Int())
	helo := Helo{
		Type: "HELO",
		Options: HeloOptions{
			Nonce:     nonce,
			Auth:      "",
			KeepAlive: true,
		},
	}
	if err := encoder.Encode(&helo); err != nil {
		return false, err
	}
	if err := bwriter.Flush(); err != nil {
		return false, err
	}

	// read PING
	ping := Ping{}
	if err := decoder.Decode(&ping); err != nil {
		return false, err
	}
	if ping.Type != "PING" {
		return false, errors.New("client sent garbage PING: " + ping.Type)
	}
	result, reason := auth(ping.ClientHostname, ping.Username, ping.Password)

	// send PONG
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	pong := Pong{
		Type:               "PONG",
		AuthResult:         result,
		Reason:             reason,
		ServerHostname:     hostname,
		SharedKeyHexdigest: sha512ToHexdigest(ping.SharedKeySalt + hostname + nonce + sharedKey),
	}
	if err := encoder.Encode(&pong); err != nil {
		return false, err
	}
	if err := bwriter.Flush(); err != nil {
		return false, err
	}

	return result, nil
}
