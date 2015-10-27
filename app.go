package mello

import (
	"fmt"
	"net"
	"time"
)

type MessageType int

const (
	MessageToClient MessageType = iota
	MessageToGate
)

type MelloApp struct {
	Master        *ServerConfig     // master server config
	CurSvrConfig  *ServerConfig     // current server info
	RemoveChan chan string       // remove server channel
	RegisterChan    chan ServerConfig // add server channel
	MessageChan   chan Message      // message channel
}

func NewApp() *MelloApp {
	return &MelloApp{
		RemoveChan: make(chan string, 10),
		RegisterChan:    make(chan ServerConfig, 10),
		MessageChan:   make(chan Message, 10000)}
}

func (app *MelloApp) Start() {
	var endRunning = make(chan bool, 1)
	app.loadDefaultComps()
	
	// enable port listener
	if app.CurSvrConfig.IsFrontend {
		go app.handlerListen()
	} else {
		go app.rpcListen()
	}
	// main goroutine
	app.listenChan()
	<-endRunning
	Info(fmt.Sprintf("Server: %s is stopping..."))
	// close all channels
	close(app.MessageChan)
	close(app.RegisterChan)
	close(app.RemoveChan)
	close(endRunning)

	// close all of components
	Rpc.Close()
}

// Enable current server backend listener
func (app *MelloApp) rpcListen() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.CurSvrConfig.Host, app.CurSvrConfig.Port))
	if err != nil {
		Error(err.Error())
	}
	Info(fmt.Sprintf("listen at %s:%d(%s)",
		app.CurSvrConfig.Host,
		app.CurSvrConfig.Port,
		app.CurSvrConfig.String()))

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go Rpc.Handle(conn)
	}
}

func (app *MelloApp) handlerListen() {
	// create local listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.CurSvrConfig.Host, app.CurSvrConfig.Port))
	if err != nil {
		Error(err.Error())
	}
	defer listener.Close()
	
	Info(fmt.Sprintf("listen at %s:%d(%s)",
		app.CurSvrConfig.Host,
		app.CurSvrConfig.Port,
		app.CurSvrConfig.String()))
	time.AfterFunc(10 * time.Second, func(){Rpc.Request("chat.AuthRemote.Test")})
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go Handler.Handle(conn)
	}
}

func (app *MelloApp) listenChan() {
	for {
		select {
		case svr := <-app.RegisterChan:
			registerServer(svr)
		case svrId := <-app.RemoveChan:
			removeServer(svrId)
		case msg := <-app.MessageChan:
			app.handleMessage(msg)
		}
	}
}

func (app *MelloApp) handleMessage(msg Message) {
	Info(msg.String())
}

func (app *MelloApp) loadDefaultComps() {
	Rpc.Register(new(Manager))
}
