package starx

import (
	"time"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/timer"
)

var (
	comps             = make([]component.Component, 0)
	heartbeatInternal = 60 * time.Second
)

func startupComps() {
	cluster.SetSessionManager(defaultNetService)
	if App.Config.IsFrontend {
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
		if App.Config.IsFrontend {
			handler.register(c)
		} else {
			remote.register(c)
		}
	}

	handler.dumpServiceMap()
	remote.dumpServiceMap()
}

func shutdownComps() {
	for _, c := range comps {
		c.BeforeShutdown()
	}
	for _, c := range comps {
		c.Shutdown()
	}
}
