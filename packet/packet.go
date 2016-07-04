package packet

import (
	"fmt"
)

type PacketType byte

const (
	_ PacketType = iota
	Handshake
	HandshakeAck
	Heartbeat
	Data
	Kick
)

const (
	HeadLength = 4
)

type Packet struct {
	Type   PacketType
	Length int
	Data   []byte
}

func NewPacket() *Packet {
	return &Packet{}
}

// PACKET PROTOCOL
// - protocol type(1 byte)
// - packet data length(3 byte big end)
// - data segment
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func Pack(t PacketType, data []byte) []byte {
	var buf []byte
	return append(append(append(buf, byte(t)), intToBytes(len(data))...), data...)
}

func (p *Packet) String() string {
	return fmt.Sprintf("[PACKET]Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}

// Decode binary data to packet
// If packet has not been received completely, return nil and incomplete data,
// concrete protocol ref pack function
func Unpack(data []byte) (*Packet, []byte) {
	t := PacketType(data[0])
	length := bytesToInt(data[1:HeadLength])
	// 包未传输完成
	if length > (len(data) - HeadLength) {
		return nil, data
	}
	p := NewPacket()
	p.Type = t
	p.Length = length
	p.Data = data[HeadLength:(length + HeadLength)]
	return p, data[(length + HeadLength):]
}

// Big end
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// Big end
func intToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}
