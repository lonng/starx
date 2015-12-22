package starx

import (
	"fmt"
	"starx/rpc"
	"strings"
)

// Start application
func Start() {
	parseConfig()
	loadSettings()
	App.start()
}

func Router(svrType string, fn func(*Session) string) {
	if t := strings.TrimSpace(svrType); t != "" {
		route[svrType] = fn
	}
}

// Set special server initial function, starx.Set("oneServerType | anotherServerType", func(){})
func Set(svrTypes string, fn func()) {
	var types = strings.Split(strings.TrimSpace(svrTypes), "|")
	for _, t := range types {
		t = strings.TrimSpace(t)
		settings[t] = append(settings[t], fn)
	}
}

func loadSettings() {
	Info(fmt.Sprintf("loading %s settings", App.Config.Type))
	if setting, ok := settings[App.Config.Type]; ok && len(setting) > 0 {
		for _, fn := range setting {
			fn()
		}
	}
}

// Handler register
func Handler(comp Component) {
	if App.Config.IsFrontend {
		handler.register(comp)
	} else {
		remote.register(rpc.SysRpc, comp)
	}
}

// Remote register
func Remote(comp Component) {
	if App.Config.IsFrontend {
		Error("current server is frontend server")
	} else {
		remote.register(rpc.UserRpc, comp)
	}
}
