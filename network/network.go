package network

import (
	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"strings"
)

var appConfig *cluster.ServerConfig

func SetAppConfig(c *cluster.ServerConfig) {
	appConfig = c
	// enable all app service
}

// TODO: ***NOTICE***
// Runtime set dictionary will be a dangerous operation!!!!!!
func SetDict(dict map[string]int) {
	for route, code := range dict {
		r := strings.TrimSpace(route)

		// duplication check
		if _, ok := Handler.routeMap[r]; ok {
			log.Warn("duplicated route(route: %s, code: %d)", r, code)
		}

		if _, ok := Handler.codeMap[code]; ok {
			log.Warn("duplicated route(route: %s, code: %d)", r, code)
		}

		// update map, using last value when key duplicated
		Handler.routeMap[r] = code
		Handler.codeMap[code] = r
	}
}
