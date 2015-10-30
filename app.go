package starx

import (
	"fmt"
	"net"
)

type MessageType int

const (
	MessageToClient MessageType = iota
	MessageToGate
)

type StarxApp struct {
	Master       *ServerConfig      // master server config
	CurSvrConfig *ServerConfig      // current server info
	RemoveChan   chan string        // remove server channel
	RegisterChan chan *ServerConfig // add server channel
	MessageChan  chan *Message      // message channel
	PacketChan   chan *Packet       // package channel
}

func NewApp() *StarxApp {
	return &StarxApp{
		RemoveChan:   make(chan string, 10),
		RegisterChan: make(chan *ServerConfig, 10),
		MessageChan:  make(chan *Message, 10000),
		PacketChan:   make(chan *Packet, 1000)}
}

func (app *StarxApp) Start() {
	var endRunning = make(chan bool, 1)
	app.loadDefaultComps()

	// enable port listener
	go app.listenPort()
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
func (app *StarxApp) listenPort() {
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
		if app.CurSvrConfig.IsFrontend {
			go Handler.Handle(conn)
		} else {
			go Rpc.Handle(conn)
		}
	}
}

func (app *StarxApp) listenChan() {
	for {
		select {
		case svr := <-app.RegisterChan:
			registerServer(*svr)
		case svrId := <-app.RemoveChan:
			removeServer(svrId)
		case msg := <-app.MessageChan:
			app.handleMessage(msg)
		case pkg := <-app.PacketChan:
			app.handlePacket(pkg)
		}
	}
}

func (app *StarxApp) handleMessage(msg *Message) {
	Info(msg.String())
}

func (app *StarxApp) handlePacket(pkg *Packet) {
	fmt.Println(pkg.String())
	Net.Broadcast(Package(TransData, []byte("message broadcast from "+app.CurSvrConfig.Id)))
}

func (app *StarxApp) loadDefaultComps() {
	Rpc.Register(new(Manager))
}
