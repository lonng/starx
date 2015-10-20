package mello

import (
	"fmt"
	"net"
)

type MessageType int

const (
	MessageToClient MessageType = iota
	MessageToGate
)

type MelloApp struct {
	Master        *ServerConfig     // master server config
	CurSvrConfig  *ServerConfig     // current server info
	SvrRemoveChan chan string       // remove server channel
	SvrAddChan    chan ServerConfig // add server channel
	MessageChan   chan Message      // message channel
}

func NewApp() *MelloApp {
	return &MelloApp{
		SvrRemoveChan: make(chan string, 10),
		SvrAddChan:    make(chan ServerConfig, 10),
		MessageChan:   make(chan Message, 10000)}
}

func (app *MelloApp) Start() {
	var endRunning = make(chan bool, 1)
	app.loadDefaultComps()
	go app.enablePortListen()
	if app.CurSvrConfig.IsFrontend {
		go app.enableClientPortListen()
	}
	app.listenChan()
	<-endRunning
	Info(fmt.Sprintf("Server: %s is stopping..."))
	// close all channels
	close(app.MessageChan)
	close(app.SvrAddChan)
	close(app.SvrRemoveChan)
	close(endRunning)

	// close all of components
	Rpc.Close()
}

// Enable current server backend listener
func (app *MelloApp) enablePortListen() {
	Info(fmt.Sprintf("enable port listener(%s)", app.CurSvrConfig.String()))
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", app.CurSvrConfig.Host, app.CurSvrConfig.Port))
	if err != nil {
		Error(err.Error())
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		Error(err.Error())
	}
	Info(fmt.Sprintf("listen at %s:%d successfully(%s)",
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

func (app *MelloApp) enableClientPortListen() {
	// create local listener
	Info(fmt.Sprintf("enable clientPort listener(%s)", app.CurSvrConfig.String()))
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", app.CurSvrConfig.Host, app.CurSvrConfig.ClientPort))
	if err != nil {
		Error(err.Error())
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		Error(err.Error())
	}
	defer listener.Close()
	Info(fmt.Sprintf("listen at %s:%d successfully(%s)",
		app.CurSvrConfig.Host,
		app.CurSvrConfig.ClientPort,
		app.CurSvrConfig.String()))

	Rpc.Request("master.Manager.Test")
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
		case svr := <-app.SvrAddChan:
			addServer(svr)
		case svrId := <-app.SvrRemoveChan:
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
