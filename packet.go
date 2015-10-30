package starx

import (
	"fmt"
)

type PacketType byte

const (
	_ PacketType = iota
	Handshake
	HandshakeACK
	Heartbeat
	TransData
	ConnectionClose
)

const (
	headLength = 4
)

type Packet struct {
	Type   PacketType
	Length int
	Body   []byte
}

func NewPacket() *Packet {
	return &Packet{}
}

func Package(t PacketType, data []byte) []byte {
	var buf []byte
	return append(append(append(buf, byte(t)), intToBytes(len(data))...), data...)
}

func (p *Packet) String() string {
	return fmt.Sprintf("type: %d, length: %d, data: %s", p.Type, p.Length, string(p.Body))
}

func UnPackage(data []byte) []byte {
	t := PacketType(data[0])
	length := bytesToInt(data[1:headLength])
	// 包未传输完成
	if length > (len(data) - headLength) {
		return data
	}
	p := NewPacket()
	p.Type = t
	p.Length = length
	p.Body = data[headLength:(length + headLength)]
	// 将包放入处理队列
	App.PacketChan <- p
	// 返回截断的包
	return data[(length + headLength):]
}

// bigend byte
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// bigend, return 3 byte
func intToBytes(n int) []byte {
	var buf []byte
	buf = append(buf, byte((n >> 16)&0xFF))
	buf = append(buf, byte((n >> 8)&0xFF))
	buf = append(buf, byte(n&0xFF))
	return buf
}
