package starx

type ProtocolType byte

const (
	_ ProtocolType = iota
	Handshake
	HandshakeACK
	Heartbeat
	TransData
	ConnectionClose
)
