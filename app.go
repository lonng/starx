package starx

import (
	"fmt"
	"github.com/chrislonng/starx/rpc"
	"net"
	"time"
)

type starxApp struct {
	Master     *ServerConfig // master server config
	Config     *ServerConfig // current server information
	AppName    string
	Standalone bool // current server is running in standalone mode
	StartTime  time.Time
}

func newApp() *starxApp {
	return &starxApp{StartTime: time.Now()}
}

func (app *starxApp) start() {
	app.loadComps()

	// enable all app service
	if app.Config.IsFrontend {
		go heartbeat.start()
	}
	app.listenAndServe()

	// stop server
	<-endRunning
	Info("server: " + app.Config.Id + " is stopping...")
	app.shutdownComps()
	close(endRunning)
}

// Enable current server accept connection
func (app *starxApp) listenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port))
	if err != nil {
		Error(err.Error())
	}
	Info("listen at %s:%d(%s)",
		app.Config.Host,
		app.Config.Port,
		app.Config.String())

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

func (app *starxApp) loadComps() {
	// handlers
	for _, comp := range handlers {
		comp.Init()
	}
	for _, comp := range handlers {
		comp.AfterInit()
	}

	// remotes
	for _, comp := range remotes {
		comp.Init()
	}
	for _, comp := range remotes {
		comp.AfterInit()
	}

	// register
	for _, comp := range handlers {
		if App.Config.IsFrontend {
			handler.register(comp)
		} else {
			remote.register(rpc.SysRpc, comp)
		}
	}
	for _, comp := range remotes {
		remote.register(rpc.UserRpc, comp)
	}
	handler.dumpServiceMap()
}

func (app *starxApp) shutdownComps() {
	// handlers
	for _, comp := range handlers {
		comp.BeforeShutdown()
	}
	for _, comp := range handlers {
		comp.Shutdown()
	}

	// remotes
	for _, comp := range remotes {
		comp.BeforeShutdown()
	}
	for _, comp := range remotes {
		comp.Shutdown()
	}
}
