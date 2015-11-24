package starx

import (
	"fmt"
)

type MessageType byte

const (
	MT_REQUEST MessageType = iota
	MT_NOTIFY
	MT_RESPONSE
	MT_PUSH
)

const (
	MSG_ROUTE_COMPRESS_MASK = 0x01
	MSG_ROUTE_LIMIT_MASK    = 0xFF
	MSG_TYPE_MASK           = 0x07
)

type Message struct {
	Type       MessageType
	ID         uint
	Route      string
	RouteCode  uint
	isCompress bool
	Body       []byte
}

func NewMessage() *Message {
	return &Message{}
}

func (this *Message) String() string {
	return fmt.Sprintf("[MESSAGE]Type: %d, ID: %d, Route: %s, IsCompress: %t, RouteCode: %d, Body: %s",
		this.Type,
		this.ID,
		this.Route,
		this.isCompress,
		this.RouteCode,
		this.Body)
}

// MESSAGE PROTOCOL
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func encodeMessage(m *Message) []byte {
	temp := make([]byte, 0)
	flag := byte(m.Type) << 1
	if m.isCompress {
		flag |= 0x01
	}
	temp = append(temp, flag)
	// response message
	if m.Type == MT_RESPONSE {
		n := m.ID
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
	} else if m.Type == MT_PUSH {
		if m.isCompress {
			temp = append(temp, byte((m.RouteCode>>8)&0xFF))
			temp = append(temp, byte(m.RouteCode&0xFF))
		} else {
			temp = append(temp, byte(len(m.Route)))
			temp = append(temp, []byte(m.Route)...)
		}
	} else {
		Error("wrong message type")
	}
	temp = append(temp, m.Body...)
	return temp
}

func decodeMessage(data []byte) *Message {
	// filter invalid message
	if len(data) <= 3 {
		Info("invalid message")
		return nil
	}
	msg := NewMessage()
	flag := data[0]
	// set offset to 1, because 1st byte will always be flag
	offset := 1
	msg.Type = MessageType((flag >> 1) & MSG_TYPE_MASK)
	if msg.Type == MT_REQUEST {
		id := uint(0)
		// little end byte order
		// WARNING: must can be stored in 64 bits integer
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
	if flag&MSG_ROUTE_COMPRESS_MASK == 1 {
		msg.isCompress = true
		msg.RouteCode = uint(bytesToInt(data[offset:(offset + 2)]))
		offset += 2
	} else {
		msg.isCompress = false
		rl := data[offset]
		offset += 1
		msg.Route = string(data[offset:(offset + int(rl))])
		offset += int(rl)
	}
	msg.Body = data[offset:]
	return msg
}
