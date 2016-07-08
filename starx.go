package starx

import (
	"fmt"
	"strings"

	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network"
	"github.com/chrislonng/starx/serialize"
)

// Start application
func Start() {
	welcomeMsg()
	parseConfig()
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

func loadSettings() {
	log.Info("loading %s settings", App.Config.Type)
	if setting, ok := settings[App.Config.Type]; ok && len(setting) > 0 {
		for _, fn := range setting {
			fn()
		}
	}
}

func welcomeMsg() {
	fmt.Println(asciiLogo)
}

// Handler register
func Register(comp network.Component) {
	network.Register(comp)
}

func Serializer(seri serialize.Serializer) {
	network.Serializer(seri)
}
