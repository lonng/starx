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
		Route[svrType] = fn
	}
}

// Set special server initial setting function
// Example:
//	starx.Set("gate", func() {
//		// setting just valid for gate
//	})
//
//	starx.Set("gate|connector" func() {
//		// setting valid for gate & connector
//	})
func Set(svrTypes string, fn func()) {
	var types = strings.Split(strings.TrimSpace(svrTypes), "|")
	for _, t := range types {
		t = strings.TrimSpace(t)
		Settings[t] = append(Settings[t], fn)
	}
}

func loadSettings() {
	Info(fmt.Sprintf("loading %s settings", App.Config.Type))
	if setting, ok := Settings[App.Config.Type]; ok && len(setting) > 0 {
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
