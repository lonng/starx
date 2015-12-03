package starx

import (
	"fmt"
	"net"
)

type _app struct {
	Master *ServerConfig // master server config
	Config *ServerConfig // current server info
}

func newApp() *_app {
	return &_app{}
}

func (app *_app) start() {
	app.loadDefaultComps()
	// enable heartbeat service
	go heartbeatService.start()
	// enable port listener
	app.listenPort()
	// waiting for application shutdown
	<-endRunning
	Info("server: " + app.Config.Id + " is stopping...")
	// close all channels
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

func (app *_app) loadDefaultComps() {
	remote.register(new(Manager))
}
