package starx

type ProtocolState byte

const (
	_ ProtocolState = iota
	PROTOCOL_START
	PROTOCOL_HANDSHAKING
	PROTOCOL_WORKING
	PROTOCOL_CLOSED
)
