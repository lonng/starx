package mello

import (
	"fmt"
	"io/ioutil"
	"net"
)

type MelloHandler struct{}

type HandlerComponent interface {
	Setup()
}

type Message struct {
	Route string
	Body  []byte
}

func (this *Message) String() string {
	return fmt.Sprintf("Route: %s, Body: %s",
		this.Route,
		this.Body)
}

func NewHandler() *MelloHandler {
	return &MelloHandler{}
}

func (handler *MelloHandler) Handle(conn net.Conn) {
	defer conn.Close()
	if buf, err := ioutil.ReadAll(conn); err != nil {
		Info(fmt.Sprintf("Data: (%s)", buf))
	}
}

func (handler *MelloHandler) Register(rcvr HandlerComponent) {
	Info(fmt.Sprintf("Register Handler: %s", rcvr))
	rcvr.Setup()
}
