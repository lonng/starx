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
	ID         int
	Route      string
	RouteCode  int
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
