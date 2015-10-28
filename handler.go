package mello

import (
	"fmt"
	"net"
)

type HandlerService struct{}

type HandlerComponent interface {
	Setup()
}

type Package struct {
	Type   ProtocolType
	Length int
	Body   []byte
}

func NewPackage() *Package {
	return &Package{}
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
	tmp := make([]byte, 0) //保存截断数据
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("connection error: " + err.Error())
			break
		}
		tmp = decodePackage(append(tmp, buf[:n]...))
	}
}

func decodePackage(data []byte) []byte {
	t := ProtocolType(data[0])
	length := bytesToInt(data[1:3])
	// 包未传输完成
	if length > (len(data) - 3) {
		return data
	}
	p := NewPackage()
	p.Type = t
	p.Length = length
	p.Body = data[3:(length + 3)]
	// 将包放入处理队列
	App.PackageChan <- p
	// 返回截断的包
	return data[(length + 3):]
}

// bigend byte
func bytesToInt(b []byte) int {
	var result int
	for i, v := range b {
		result = result<<(uint(i)*8) + int(v)
	}
	return result
}

func (handler *HandlerService) Register(rcvr HandlerComponent) {
	Info(fmt.Sprintf("Register Handler: %s", rcvr))
	rcvr.Setup()
}
