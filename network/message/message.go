package message

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/chrislonng/starx/log"
	"strings"
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

var (
	routeDict = make(map[string]uint16)
	codeDict  = make(map[uint16]string)
)

var (
	ErrWrongMessageType  = errors.New("wrong message type")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrRouteInfoNotFound = errors.New("route info not found in dictionary")
)

type Message struct {
	Type       MessageType
	ID         uint
	Route      string
	Data       []byte
	compressed bool
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) String() string {
	return fmt.Sprintf("Type: %d, ID: %d, Route: %s, Compressed: %t, BodyLength: %d",
		m.Type,
		m.ID,
		m.Route,
		m.compressed,
		len(m.Data))
}

func (m *Message) Encode() ([]byte, error) {
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
func Encode(m *Message) ([]byte, error) {
	buf := make([]byte, 0)
	flag := byte(m.Type) << 1

	code, compressed := routeDict[m.Route]
	if compressed {
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
	case Notify, Push:
		if compressed {
			buf = append(buf, byte((code>>8)&0xFF))
			buf = append(buf, byte(code&0xFF))
		} else {
			buf = append(buf, byte(len(m.Route)))
			buf = append(buf, []byte(m.Route)...)
		}
	default:
		log.Error("wrong message type")
		return nil, ErrWrongMessageType
	}
	buf = append(buf, m.Data...)
	return buf, nil
}

func Decode(data []byte) (*Message, error) {
	if len(data) <= msgHeadLength {
		log.Info("invalid message")
		return nil, ErrInvalidMessage
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
	case Notify, Push:
		if flag&msgRouteCompressMask == 1 {
			m.compressed = true
			code := binary.BigEndian.Uint16(data[offset:(offset + 2)])
			route, ok := codeDict[code]
			if !ok {
				log.Error("message compressed, but can not find route infomation in dictionary")
				return nil, ErrRouteInfoNotFound
			}
			m.Route = route
			offset += 2
		} else {
			m.compressed = false
			rl := data[offset]
			offset += 1
			m.Route = string(data[offset:(offset + int(rl))])
			offset += int(rl)
		}
	default:
		log.Error("wrong message type")
		return nil, ErrWrongMessageType
	}
	m.Data = data[offset:]
	return m, nil
}

// TODO: ***NOTICE***
// Runtime set dictionary will be a dangerous operation!!!!!!
func SetDict(dict map[string]uint16) {
	for route, code := range dict {
		r := strings.TrimSpace(route)

		// duplication check
		if _, ok := routeDict[r]; ok {
			log.Warn("duplicated route(route: %s, code: %d)", r, code)
		}

		if _, ok := codeDict[code]; ok {
			log.Warn("duplicated route(route: %s, code: %d)", r, code)
		}

		// update map, using last value when key duplicated
		routeDict[r] = code
		codeDict[code] = r
	}
}
