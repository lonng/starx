package starx

import (
	"fmt"
)

type PacketType byte

const (
	_ PacketType = iota
	PACKET_HANDSHAKE
	PACKET_HANDSHAKE_ACK
	PACKET_HEARTBEAT
	PACKET_DATA
	PACKET_KICK
)

const (
	headLength = 4
)

type Packet struct {
	Type    PacketType
	Length  int
	Body    []byte
	session *Session
}

type unhandledPacket struct {
	fs     *frontendSession
	packet *Packet
}

func NewPacket() *Packet {
	return &Packet{}
}

// PACKET PROTOCOL
// - protocol type(1 byte)
// - packet data length(3 byte big end)
// - data segment
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func pack(t PacketType, data []byte) []byte {
	var buf []byte
	return append(append(append(buf, byte(t)), intToBytes(len(data))...), data...)
}

func (p *Packet) String() string {
	return fmt.Sprintf("[PACKET]Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Body))
}

// Decode binary data to packet
// If packet has not been received completely, return nil and incomplete data,
// concrete protocol ref pack function
func unpack(data []byte) (*Packet, []byte) {
	t := PacketType(data[0])
	length := bytesToInt(data[1:headLength])
	// 包未传输完成
	if length > (len(data) - headLength) {
		return nil, data
	}
	p := NewPacket()
	p.Type = t
	p.Length = length
	p.Body = data[headLength:(length + headLength)]
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
