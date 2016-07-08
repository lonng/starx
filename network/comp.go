package network

import (
	"github.com/chrislonng/starx/timer"
	"time"
)

var (
	comps             = make([]Component, 0)
	heartbeatInternal = 60 * time.Second
)

type Component interface {
	Init()
	AfterInit()
	BeforeShutdown()
	Shutdown()
}

func Register(c Component) {
	comps = append(comps, c)
}

func Startup() {
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
}

func Shutdown() {
	for _, c := range comps {
		c.BeforeShutdown()
	}
	for _, c := range comps {
		c.Shutdown()
	}
}
