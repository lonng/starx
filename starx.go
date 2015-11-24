package starx

import (
	"fmt"
	"strings"
)

func Start() {
	ParseConfig()
	loadSettings()
	App.Start()
}

func Router(svrType string, fn func() string) {
	if t := strings.TrimSpace(svrType); t != "" {
		Route[svrType] = fn
	}
}

// Server setting function
// Usage:
//	mello.Set("gate", func() {
//		// setting just valid for gate
//	})
//
//	mello.Set("gate|connector" func() {
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
	Info(fmt.Sprintf("loading %s settings", App.CurSvrConfig.Type))
	if setting, ok := Settings[App.CurSvrConfig.Type]; ok && len(setting) > 0 {
		for _, fn := range setting {
			fn()
		}
	}
}

func Register(comp Component) {
	if App.CurSvrConfig.IsFrontend {
		handler.register(comp)
	} else {
		remote.register(comp)
	}
}
