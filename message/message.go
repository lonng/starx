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

func newMessage() *Message {
	return &Message{}
}

func (msg *Message) String() string {
	return fmt.Sprintf("[MESSAGE]Type: %d, ID: %d, Route: %s, IsCompress: %t, RouteCode: %d, Body: %s",
		msg.Type,
		msg.ID,
		msg.Route,
		msg.IsCompress,
		msg.RouteCode,
		msg.Data)
}

func (msg *Message) encoding() []byte {
	return Encode(msg)
}

// MESSAGE PROTOCOL
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func Encode(m *Message) []byte {
	temp := make([]byte, 0)
	flag := byte(m.Type) << 1
	if m.IsCompress {
		flag |= 0x01
	}
	temp = append(temp, flag)

	// response message
	if m.Type == Response {
		n := m.ID
		// variant length encode
		for {
			b := byte(n % 128)
			n >>= 7
			if n != 0 {
				temp = append(temp, b+128)
			} else {
				temp = append(temp, b)
				break
			}
		}
	} else if m.Type == Push {
		if m.IsCompress {
			temp = append(temp, byte((m.RouteCode>>8)&0xFF))
			temp = append(temp, byte(m.RouteCode&0xFF))
		} else {
			temp = append(temp, byte(len(m.Route)))
			temp = append(temp, []byte(m.Route)...)
		}
	} else {
		log.Error("wrong message type")
	}
	temp = append(temp, m.Data...)
	return temp
}

func Decode(data []byte) *Message {
	// filter invalid message
	if len(data) <= 3 {
		log.Info("invalid message")
		return nil
	}
	msg := newMessage()
	flag := data[0]
	// set offset to 1, because 1st byte will always be flag
	offset := 1
	msg.Type = MessageType((flag >> 1) & msgTypeMask)
	if msg.Type == Request {
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
		msg.ID = id
	}
	if flag&msgRouteCompressMask == 1 {
		msg.IsCompress = true
		msg.RouteCode = binary.BigEndian.Uint32(data[offset:(offset + 2)])
		offset += 2
	} else {
		msg.IsCompress = false
		rl := data[offset]
		offset += 1
		msg.Route = string(data[offset:(offset + int(rl))])
		offset += int(rl)
	}
	msg.Data = data[offset:]
	return msg
}
