package starx

import (
	"fmt"
	"net"
)

type HandlerService struct{}

func NewHandler() *HandlerService {
	return &HandlerService{}
}

func (handler *HandlerService) Handle(conn net.Conn) {
	defer conn.Close()
	if sessionService.isSessionExists(conn.RemoteAddr().String()) {
		Info("sesesion addr already exists: " + conn.RemoteAddr().String())
	} else {
		session := NewSession(conn)
		sessionService.RegisterSession(session)
		sessionService.dumpSessions()
	}
	tmp := make([]byte, 0) //保存截断数据
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("remote address: " + conn.RemoteAddr().String())
			Info("connection error: " + err.Error())
			sessionService.RemoveSession(conn.RemoteAddr().String())
			sessionService.dumpSessions()
			break
		}
		p, tmp := UnPackage(append(tmp, buf[:n]...))
	}
}

func (handler *HandlerService) Register(rcvr HandlerComponent) {
	Info(fmt.Sprintf("Register Handler: %s", rcvr))
	rcvr.Setup()
}
