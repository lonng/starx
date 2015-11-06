package starx

import "fmt"

type routeInfo struct {
	server string
	service string
	method string
}

func newRouteInfo(server, service, method string) *routeInfo {
	return &routeInfo{server, service, method}
}

func (r *routeInfo) String() string{
	return fmt.Sprintf("server: %s, service: %s, method: %s", r.server, r.service, r.method)
}