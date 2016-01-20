package starx

import (
	"fmt"
	"net"
	"starx/rpc"
)

type starxApp struct {
	Master *ServerConfig // master server config
	Config *ServerConfig // current server information
}

func newApp() *starxApp {
	return &starxApp{}
}

func (app *starxApp) start() {
	app.loadDefaultComps()

	// enable all app service
	if app.Config.IsFrontend {
		go heartbeatService.start()
	}
	app.listenAndServe()

	// stop server
	<-endRunning
	Info("server: " + app.Config.Id + " is stopping...")
	close(endRunning)
}

// Enable current server accept connection
func (app *starxApp) listenAndServe() {
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
			Error(err.Error())
			continue
		}
		if app.Config.IsFrontend {
			go handler.handle(conn)
		} else {
			go remote.handle(conn)
		}
	}
}

func (app *starxApp) loadDefaultComps() {
	remote.register(rpc.SysRpc, new(Manager))
}
