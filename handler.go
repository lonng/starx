package mello

import (
	"fmt"
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
	Info(conn.RemoteAddr().String())
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("read hanlder message error " + err.Error())
			break
		}
		fmt.Println(string(buf[:n]))
	}
}

func (handler *HandlerService) Register(rcvr HandlerComponent) {
	Info(fmt.Sprintf("Register Handler: %s", rcvr))
	rcvr.Setup()
}
