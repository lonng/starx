package starx

import (
	"fmt"
	"net"
)

type _app struct {
	Master       *ServerConfig      // master server config
	CurSvrConfig *ServerConfig      // current server info
	RemoveChan   chan string        // remove server channel
	RegisterChan chan *ServerConfig // add server channel
}

func newApp() *_app {
	return &_app{
		RemoveChan:   make(chan string, 10),
		RegisterChan: make(chan *ServerConfig, 10)}
}

func (app *_app) Start() {
	var endRunning = make(chan bool, 1)
	app.loadDefaultComps()

	// enable port listener
	go app.listenPort()
	go heartbeatService.start()
	// main goroutine
	app.listenChan()

	<-endRunning
	Info("server: " + app.CurSvrConfig.Id + " is stopping...")
	// close all channels
	close(app.RegisterChan)
	close(app.RemoveChan)
	close(endRunning)

	// close all of components
	remote.close()
}

// Enable current server backend listener
func (app *_app) listenPort() {
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
			go handler.handle(conn)
		} else {
			go remote.handle(conn)
		}
	}
}

func (app *_app) listenChan() {
	for {
		select {
		case svr := <-app.RegisterChan:
			registerServer(*svr)
		case svrId := <-app.RemoveChan:
			removeServer(svrId)
		}
	}
}

func (app *_app) loadDefaultComps() {
	remote.register(new(Manager))
}
