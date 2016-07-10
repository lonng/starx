package network

import (
	"github.com/chrislonng/starx/cluster"
)

var appConfig *cluster.ServerConfig

func SetAppConfig(c *cluster.ServerConfig) {
	appConfig = c
	// enable all app service
}