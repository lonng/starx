package message

import (
	"encoding/binary"
	"fmt"
	"github.com/chrislonng/starx/log"
)

type MessageType byte

const (
	Request MessageType = iota
	Notify
	Response
	Push
)

const (
	msgRouteCompressMask = 0x01
	msgTypeMask          = 0x07
	msgRouteLengthMask   = 0xFF
)

type Message struct {
	Type       MessageType
	ID         uint
	Route      string
	RouteCode  uint32
	IsCompress bool
	Data       []byte
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) String() string {
	return fmt.Sprintf("Type: %d, ID: %d, Route: %s, IsCompress: %t, RouteCode: %d, Body: %s",
		m.Type,
		m.ID,
		m.Route,
		m.IsCompress,
		m.RouteCode,
		m.Data)
}

func (m *Message) encoding() []byte {
	return Encode(m)
}

// MESSAGE PROTOCOL
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func Encode(m *Message) []byte {
	buf := make([]byte, 0)
	flag := byte(m.Type) << 1
	if m.IsCompress {
		flag |= msgRouteCompressMask
	}
	buf = append(buf, flag)

	// response message
	if m.Type == Response {
		n := m.ID
		// variant length encode
		for {
			b := byte(n % 128)
			n >>= 7
			if n != 0 {
				buf = append(buf, b+128)
			} else {
				buf = append(buf, b)
				break
			}
		}
	} else if m.Type == Push {
		if m.IsCompress {
			buf = append(buf, byte((m.RouteCode>>8)&0xFF))
			buf = append(buf, byte(m.RouteCode&0xFF))
		} else {
			buf = append(buf, byte(len(m.Route)))
			buf = append(buf, []byte(m.Route)...)
		}
	} else {
		log.Error("wrong message type")
	}
	buf = append(buf, m.Data...)
	return buf
}

func Decode(data []byte) *Message {
	// filter invalid message
	if len(data) <= 3 {
		log.Info("invalid message")
		return nil
	}
	m := NewMessage()
	flag := data[0]
	// set offset to 1, because 1st byte will always be flag
	offset := 1
	m.Type = MessageType((flag >> 1) & msgTypeMask)
	if m.Type == Request {
		id := uint(0)
		// little end byte order
		// WARNING: must can be stored in 64 bits integer
		// variant length encode
		for i := offset; i < len(data); i++ {
			b := data[i]
			id += (uint(b&0x7F) << uint(7*(i-offset)))
			if b < 128 {
				offset = i + 1
				break
			}
		}
		m.ID = id
	}
	if flag&msgRouteCompressMask == 1 {
		m.IsCompress = true
		m.RouteCode = binary.BigEndian.Uint32(data[offset:(offset + 2)])
		offset += 2
	} else {
		m.IsCompress = false
		rl := data[offset]
		offset += 1
		m.Route = string(data[offset:(offset + int(rl))])
		offset += int(rl)
	}
	m.Data = data[offset:]
	return m
}
