package forwardprotocol

// Helo is the HELO message from server to client during forward protocol handshake step 1
type Helo struct {
	_msgpack struct{}    `msgpack:",asArray"`
	Type     string      `msgpack:"type"`
	Options  HeloOptions `msgpack:"options"`
}

// HeloOptions is a map of options returned from fluent server
type HeloOptions struct {
	Nonce     string `msgpack:"nonce"`
	Auth      string `msgpack:"auth"`
	KeepAlive bool   `msgpack:"keepalive"`
}

// Ping is the PING message from client to server during forward protocol handshake step 2
type Ping struct {
	_msgpack           struct{} `msgpack:",asArray"`
	Type               string   `msgpack:"type"`
	ClientHostname     string   `msgpack:"client_hostname"`
	SharedKeySalt      string   `msgpack:"shared_key_salt"`
	SharedKeyHexdigest string   `msgpack:"shared_key_hexdigest"`
	Username           string   `msgpack:"username"`
	Password           string   `msgpack:"password"`
}

// Pong is the PONG message from server to client during forward protocol handshake step 3
type Pong struct {
	_msgpack           struct{} `msgpack:",asArray"`
	Type               string   `msgpack:"type"`
	AuthResult         bool     `msgpack:"auth_result"`
	Reason             string   `msgpack:"reason"`
	ServerHostname     string   `msgpack:"server_hostname"`
	SharedKeyHexdigest string   `msgpack:"shared_key_hexdigest"`
}

func init() {
	unusedStruct(Helo{}._msgpack)
	unusedStruct(Ping{}._msgpack)
	unusedStruct(Pong{}._msgpack)
}
