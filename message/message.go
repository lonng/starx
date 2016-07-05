package message

import (
	"encoding/binary"
	"fmt"
	"github.com/chrislonng/starx/log"
)

type MessageType byte

const (
	Request  MessageType = 0x00
	Notify               = 0x01
	Response             = 0x02
	Push                 = 0x03
)

const (
	msgRouteCompressMask = 0x01
	msgTypeMask          = 0x07
	msgRouteLengthMask   = 0xFF
	msgHeadLength        = 0x03
)

type Message struct {
	Type       MessageType
	ID         uint
	Route      string
	RouteCode  uint16
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

func (m *Message) Encode() []byte {
	return Encode(m)
}

// Encode message. Different message types is corresponding to different message header,
// message types is identified by 2-4 bit of flag field. The relationship between message
// types and message header is presented as follows:
//
//   type      flag      other
//   ----      ----      -----
// request  |----000-|<message id>|<route>
// notify   |----001-|<route>
// response |----010-|<message id>|<route>
// push     |----011-|<route>
// The figure above indicates that the bit does not affect the type of message.
func Encode(m *Message) []byte {
	buf := make([]byte, 0)
	flag := byte(m.Type) << 1
	if m.IsCompress {
		flag |= msgRouteCompressMask
	}
	buf = append(buf, flag)

	switch m.Type {
	case Request, Response:
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
		fallthrough
	case Notify, Push:
		if m.IsCompress {
			buf = append(buf, byte((m.RouteCode>>8)&0xFF))
			buf = append(buf, byte(m.RouteCode&0xFF))
		} else {
			buf = append(buf, byte(len(m.Route)))
			buf = append(buf, []byte(m.Route)...)
		}
	default:
		log.Error("wrong message type")
	}
	buf = append(buf, m.Data...)
	return buf
}

func Decode(data []byte) *Message {
	if len(data) <= msgHeadLength {
		log.Info("invalid message")
		return nil
	}
	m := NewMessage()
	flag := data[0]
	offset := 1
	m.Type = MessageType((flag >> 1) & msgTypeMask)
	switch m.Type {
	case Request, Response:
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
		fallthrough
	case Notify, Push:
		if flag&msgRouteCompressMask == 1 {
			m.IsCompress = true
			m.RouteCode = binary.BigEndian.Uint16(data[offset:(offset + 2)])
			offset += 2
		} else {
			m.IsCompress = false
			rl := data[offset]
			offset += 1
			m.Route = string(data[offset:(offset + int(rl))])
			offset += int(rl)
		}
	default:
		log.Error("wrong message type")
	}
	m.Data = data[offset:]
	return m
}
