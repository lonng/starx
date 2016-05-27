package starx

import (
	"fmt"
	"github.com/chrislonng/starx/log"
)

type messageType byte

const (
	_MT_REQUEST messageType = iota
	_MT_NOTIFY
	_MT_RESPONSE
	_MT_PUSH
)

const (
	_MSG_ROUTE_COMPRESS_MASK = 0x01
	_MSG_ROUTE_LIMIT_MASK    = 0xFF
	_MSG_TYPE_MASK           = 0x07
)

type message struct {
	kind       messageType
	id         uint
	route      string
	routeCode  uint
	isCompress bool
	body       []byte
}

func newMessage() *message {
	return &message{}
}

func (msg *message) String() string {
	return fmt.Sprintf("[MESSAGE]Type: %d, ID: %d, Route: %s, IsCompress: %t, RouteCode: %d, Body: %s",
		msg.kind,
		msg.id,
		msg.route,
		msg.isCompress,
		msg.routeCode,
		msg.body)
}

func (msg *message) encoding() []byte {
	return encodeMessage(msg)
}

// MESSAGE PROTOCOL
// refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
func encodeMessage(m *message) []byte {
	temp := make([]byte, 0)
	flag := byte(m.kind) << 1
	if m.isCompress {
		flag |= 0x01
	}
	temp = append(temp, flag)
	// response message
	if m.kind == _MT_RESPONSE {
		n := m.id
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
	} else if m.kind == _MT_PUSH {
		if m.isCompress {
			temp = append(temp, byte((m.routeCode>>8)&0xFF))
			temp = append(temp, byte(m.routeCode&0xFF))
		} else {
			temp = append(temp, byte(len(m.route)))
			temp = append(temp, []byte(m.route)...)
		}
	} else {
		log.Error("wrong message type")
	}
	temp = append(temp, m.body...)
	return temp
}

func decodeMessage(data []byte) *message {
	// filter invalid message
	if len(data) <= 3 {
		log.Info("invalid message")
		return nil
	}
	msg := newMessage()
	flag := data[0]
	// set offset to 1, because 1st byte will always be flag
	offset := 1
	msg.kind = messageType((flag >> 1) & _MSG_TYPE_MASK)
	if msg.kind == _MT_REQUEST {
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
		msg.id = id
	}
	if flag&_MSG_ROUTE_COMPRESS_MASK == 1 {
		msg.isCompress = true
		msg.routeCode = uint(bytesToInt(data[offset:(offset + 2)]))
		offset += 2
	} else {
		msg.isCompress = false
		rl := data[offset]
		offset += 1
		msg.route = string(data[offset:(offset + int(rl))])
		offset += int(rl)
	}
	msg.body = data[offset:]
	return msg
}
