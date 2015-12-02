package starx

import (
	"fmt"
	"net"
)

type _app struct {
	Master     *ServerConfig      // master server config
	Config     *ServerConfig      // current server info
	removeChan chan string        // remove server channel
	registChan chan *ServerConfig // add server channel
}

func newApp() *_app {
	return &_app{
		removeChan: make(chan string, 10),
		registChan: make(chan *ServerConfig, 10)}
}

func (app *_app) start() {
	var endRunning = make(chan bool, 1)
	app.loadDefaultComps()

	// enable port listener
	go app.listenPort()
	go heartbeatService.start()
	// main goroutine
	app.listenChan()

	<-endRunning
	Info("server: " + app.Config.Id + " is stopping...")
	// close all channels
	close(app.registChan)
	close(app.removeChan)
	close(endRunning)

	// close all of components
	remote.close()
}

// Enable current server backend listener
func (app *_app) listenPort() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port))
	if err != nil {
		Error(err.Error())
	}
	Info(fmt.Sprintf("listen at %s:%d(%s)",
		app.Config.Host,
		app.Config.Port,
		app.Config.String()))

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		if app.Config.IsFrontend {
			go handler.handle(conn)
		} else {
			go remote.handle(conn)
		}
	}
}

func (app *_app) listenChan() {
	for {
		select {
		case svr := <-app.registChan:
			registerServer(*svr)
		case svrId := <-app.removeChan:
			removeServer(svrId)
		}
	}
}

func (app *_app) loadDefaultComps() {
	remote.register(new(Manager))
}
