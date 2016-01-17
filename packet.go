package starx

import (
	"fmt"
)

type packetType byte

const (
	_ packetType = iota
	_PACKET_HANDSHAKE
	_PACKET_HANDSHAKE_ACK
	_PACKET_HEARTBEAT
	_PACKET_DATA
	_PACKET_KICK
)

const (
	headLength = 4
)

type packet struct {
	kind    packetType
	length  int
	body    []byte
	session *Session
}

type unhandledPacket struct {
	fs     *handlerSession
	packet *packet
}

func newPacket() *packet {
	return &packet{}
}

// PACKET PROTOCOL
// - protocol type(1 byte)
// - packet data length(3 byte big end)
// - data segment
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func pack(t packetType, data []byte) []byte {
	var buf []byte
	return append(append(append(buf, byte(t)), intToBytes(len(data))...), data...)
}

func (p *packet) String() string {
	return fmt.Sprintf("[PACKET]Type: %d, Length: %d, Data: %s", p.kind, p.length, string(p.body))
}

// Decode binary data to packet
// If packet has not been received completely, return nil and incomplete data,
// concrete protocol ref pack function
func unpack(data []byte) (*packet, []byte) {
	t := packetType(data[0])
	length := bytesToInt(data[1:headLength])
	// 包未传输完成
	if length > (len(data) - headLength) {
		return nil, data
	}
	p := newPacket()
	p.kind = t
	p.length = length
	p.body = data[headLength:(length + headLength)]
	return p, data[(length + headLength):]
}

// big end byte
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// big end, return 3 byte
func intToBytes(n int) []byte {
	var buf []byte
	buf = append(buf, byte((n>>16)&0xFF))
	buf = append(buf, byte((n>>8)&0xFF))
	buf = append(buf, byte(n&0xFF))
	return buf
}
