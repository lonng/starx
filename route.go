package starx

import (
	"fmt"
"strings"
	"errors"
)

type routeInfo struct {
	server  string
	service string
	method  string
}

func newRouteInfo(server, service, method string) *routeInfo {
	return &routeInfo{server, service, method}
}

func (r *routeInfo) String() string {
	return fmt.Sprintf("server: %s, service: %s, method: %s", r.server, r.service, r.method)
}

func decodeRouteInfo(route string) (*routeInfo, error) {
	parts := strings.Split(route, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid route")
	}
	return newRouteInfo(parts[0], parts[1], parts[2]), nil
}

