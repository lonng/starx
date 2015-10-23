package mello

import (
	"fmt"
	"io/ioutil"
	"net"
)

type HandlerService struct{}

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

func NewHandler() *HandlerService {
	return &HandlerService{}
}

func (handler *HandlerService) Handle(conn net.Conn) {
	defer conn.Close()
	if buf, err := ioutil.ReadAll(conn); err != nil {
		Info(fmt.Sprintf("Data: (%s)", buf))
	}
}

func (handler *HandlerService) Register(rcvr HandlerComponent) {
	Info(fmt.Sprintf("Register Handler: %s", rcvr))
	rcvr.Setup()
}
