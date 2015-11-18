package starx

import (
	"fmt"
	"net"
)

type _app struct {
	Master       *ServerConfig      // master server config
	CurSvrConfig *ServerConfig      // current server info
}

func newApp() *_app {
	return &_app{}
}

func (app *_app) Start() {
	var endRunning = make(chan bool, 1)
	app.loadDefaultComps()

	go heartbeatService.start()
	// enable port listener
	app.listenPort()

	<-endRunning
	Info("server: " + app.CurSvrConfig.Id + " is stopping...")
	// close all of components
	Rpc.Close()
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
			go Rpc.Handle(conn)
		}
	}
}

func (app *_app) loadDefaultComps() {
	Rpc.Register(new(Manager))
}
