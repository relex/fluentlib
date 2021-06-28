package forwardprotocol

import (
	"bufio"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

// DoClientHandshake performs client-side handshake on the given forward protocol connection.
//
// Returns (success?, failure reason, network error)
func DoClientHandshake(conn net.Conn, sharedKey string, timeout time.Duration) (bool, string, error) {
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return false, "failed to set timeout: " + err.Error(), nil
	}
	decoder := msgpack.NewDecoder(conn)
	bwriter := bufio.NewWriterSize(conn, 1024)
	encoder := msgpack.NewEncoder(bwriter)

	// read HELO
	helo := Helo{}
	if err := decoder.Decode(&helo); err != nil {
		return false, "", err
	}
	if helo.Type != "HELO" {
		return false, "server sent garbage HELO: " + helo.Type, nil
	}

	// send PING
	salt := strconv.Itoa(rand.Int())
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	ping := Ping{
		Type:               "PING",
		ClientHostname:     hostname,
		SharedKeySalt:      salt,
		SharedKeyHexdigest: sha512ToHexdigest(salt + hostname + helo.Options.Nonce + sharedKey),
		Username:           "",
		Password:           "",
	}
	if err := encoder.Encode(&ping); err != nil {
		return false, "", err
	}
	if err := bwriter.Flush(); err != nil {
		return false, "", err
	}

	// read PONG
	pong := Pong{}
	if err := decoder.Decode(&pong); err != nil {
		return false, "", err
	}
	if pong.Type != "PONG" {
		return false, "server returned garbage PONG: " + pong.Type, nil
	}
	serverDigest := sha512ToHexdigest(salt + pong.ServerHostname + helo.Options.Nonce + sharedKey)
	if serverDigest != pong.SharedKeyHexdigest {
		return false, "server returned invalid digest, check shared key", nil
	}
	if pong.AuthResult {
		return true, "", nil
	}

	return false, pong.Reason, nil
}
