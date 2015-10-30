package starx

import (
	"fmt"
)

type Packet struct {
	Type   ProtocolType
	Length int
	Body   []byte
}

func NewPacket() *Packet {
	return &Packet{}
}

func Package(t ProtocolType, data []byte) []byte {
	var buf []byte
	return append(append(append(buf, byte(t)), intToBytes(len(data))...), data...)
}

func (p *Packet) String() string {
	return fmt.Sprintf("type: %d, length: %d, data: %s", p.Type, p.Length, string(p.Body))
}

func UnPackage(data []byte) []byte {
	t := ProtocolType(data[0])
	length := bytesToInt(data[1:3])
	// 包未传输完成
	if length > (len(data) - 3) {
		return data
	}
	p := NewPacket()
	p.Type = t
	p.Length = length
	p.Body = data[3:(length + 3)]
	// 将包放入处理队列
	App.PacketChan <- p
	// 返回截断的包
	return data[(length + 3):]
}

// bigend byte
func bytesToInt(b []byte) int {
	var result int
	for i, v := range b {
		result = result<<(uint(i)*8) + int(v)
	}
	return result
}

// bigend
func intToBytes(n int) []byte {
	var buf []byte
	return append(append(buf, byte((n>>8)&0xFF)), byte(n&0xFF))
}
