package network

import (
	"time"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/timer"
)

var (
	appConfig         *cluster.ServerConfig
	comps             = make([]Component, 0)
	heartbeatInternal = 60 * time.Second
)

func SetAppConfig(c *cluster.ServerConfig) {
	appConfig = c
	// enable all app service
}

func Register(c Component) {
	comps = append(comps, c)
}

func Startup() {
	cluster.SetSessionManager(defaultNetService)
	if appConfig.IsFrontend {
		timer.Register(heartbeatInternal, func() {
			defaultNetService.heartbeat()
		})
	}
	for _, c := range comps {
		c.Init()
	}
	for _, c := range comps {
		c.AfterInit()
	}

	for _, c := range comps {
		if appConfig.IsFrontend {
			Handler.Register(c)
		} else {
			Remote.Register(c)
		}
	}

	Handler.dumpServiceMap()
	Remote.dumpServiceMap()
}

func Shutdown() {
	for _, c := range comps {
		c.BeforeShutdown()
	}
	for _, c := range comps {
		c.Shutdown()
	}
}
