package route

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chrislonng/starx/log"
)

var (
	ErrRouteFieldCantEmpty = errors.New("route field can not empty")
	ErrInvalidRoute        = errors.New("invalid route")
)

type Route struct {
	ServerType string
	Service    string
	Method     string
}

func NewRoute(server, service, method string) *Route {
	return &Route{server, service, method}
}

func (r *Route) String() string {
	return fmt.Sprintf("%s.%s.%s", r.ServerType, r.Service, r.Method)
}

func Decode(route string) (*Route, error) {
	r := strings.Split(route, ".")
	for _, s := range r {
		if strings.TrimSpace(s) == "" {
			return nil, ErrRouteFieldCantEmpty
		}
	}
	switch len(r) {
	case 3:
		return NewRoute(r[0], r[1], r[2]), nil
	case 2:
		return NewRoute("", r[0], r[1]), nil
	default:
		log.Error("invalid route: " + route)
		return nil, ErrInvalidRoute
	}
}
