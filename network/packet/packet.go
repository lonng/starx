package packet

import (
	"errors"
	"fmt"
	"github.com/chrislonng/starx/log"
)

type PacketType byte

const (
	_            PacketType = iota
	Handshake               = 0x01 // packet for handshake request(client) <====> handshake response(server)
	HandshakeAck            = 0x02 // packet for handshake ack from client to server
	Heartbeat               = 0x03 // heartbeat packet
	Data                    = 0x04 // data packet
	Kick                    = 0x05 // disconnect message from server
)

const HeadLength = 4

var ErrWrongPacketType = errors.New("wrong packet type")

type Packet struct {
	Type   PacketType
	Length int
	Data   []byte
}

func NewPacket() *Packet {
	return &Packet{}
}

func (p *Packet) Pack() ([]byte, error) {
	return Pack(p)
}

// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func Pack(p *Packet) ([]byte, error) {
	if p.Type < Handshake || p.Type > Kick {
		log.Errorf("wrong packet type")
		return nil, ErrWrongPacketType
	}

	p.Length = len(p.Data)

	buf := make([]byte, p.Length+HeadLength)
	buf[0] = byte(p.Type)

	copy(buf[1:HeadLength], intToBytes(p.Length))
	copy(buf[HeadLength:], p.Data)
	return buf, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}

// Unpack binary data to packet, if packet has not been received completely,
// return nil and incomplete data, concrete protocol ref pack function
func Unpack(data []byte) (*Packet, []byte, error) {
	t := PacketType(data[0])
	if t < Handshake || t > Kick {
		log.Errorf("wrong packet type")
		return nil, nil, ErrWrongPacketType
	}

	length := bytesToInt(data[1:HeadLength])
	if length > (len(data) - HeadLength) {
		return nil, data, nil
	}
	p := &Packet{
		Type:   t,
		Length: length,
		Data:   data[HeadLength:(length + HeadLength)],
	}
	return p, data[(length + HeadLength):], nil
}

// Decode packet data length byte to int(Big end)
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// Encode packet data length to bytes(Big end)
func intToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}
