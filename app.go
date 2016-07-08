package starx

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network/rpc"
	"github.com/chrislonng/starx/timer"
)

type starxApp struct {
	Master     *cluster.ServerConfig // master server config
	Config     *cluster.ServerConfig // current server information
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
		timer.Register(heartbeatInternal, func() {
			defaultNetService.heartbeat()
		})
	}
	app.listenAndServe()

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// stop server
	select {
	case <-endRunning:
		log.Info("The app will shutdown in a few seconds")
	case s := <-sg:
		log.Info("Got signal: %v", s)
	}
	log.Info("server: " + app.Config.Id + " is stopping...")
	app.shutdownComps()
	close(endRunning)
}

// Enable current server accept connection
func (app *starxApp) listenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port))
	if err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
	log.Info("listen at %s:%d(%s)",
		app.Config.Host,
		app.Config.Port,
		app.Config.String())

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err.Error())
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
			remote.register(rpc.Sys, comp)
		}
	}
	for _, comp := range remotes {
		remote.register(rpc.User, comp)
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
