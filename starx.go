package starx

import (
	"strings"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/network"
	"github.com/chrislonng/starx/serialize"
	"github.com/chrislonng/starx/session"
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

// Register component
func Register(comp network.Component) {
	network.Register(comp)
}

// Set customized serializer
func Serializer(seri serialize.Serializer) {
	network.Serializer(seri)
}

func Router(svrType string, fn func(*session.Session) string) {
	cluster.Router(svrType, fn)
}
