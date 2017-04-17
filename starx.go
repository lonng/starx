package starx

import (
	"strings"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/session"
	"time"
)

// Run server
func Run() {
	//welcomeMsg()
	parseConfig()
	App.init()
	loadSettings()
	App.start()
}

// Set special server initial function, starx.Set("oneServerType | anotherServerType", func(){})
func Set(svrTypes string, fn func()) {
	var types = strings.Split(strings.TrimSpace(svrTypes), "|")
	for _, t := range types {
		t = strings.TrimSpace(t)
		settings[t] = append(settings[t], fn)
	}
}

func SetRouter(svrType string, fn func(*session.Session) string) {
	cluster.Router(svrType, fn)
}

func Register(c component.Component) {
	comps = append(comps, c)
}

func SetServerID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		panic("empty server id")
	}
	serverID = id
}

// Set the path of app.json
func SetAppConfig(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		panic("empty app path")
	}
	appConfigPath = path
}

// Set the path of master.json
func SetMasterConfig(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		panic("empty app path")
	}
	masterConfigPath = path
}

// Set the path of servers.json
func SetServersConfig(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		panic("empty app path")
	}
	serversConfigPath = path
}

// Set heartbeat time internal
func SetHeartbeatInternal(d time.Duration) {
	heartbeatInternal = d
}
