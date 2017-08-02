package cluster

import (
	"strings"

	"github.com/lonnng/starx/session"
)

var router map[string]func(*session.Session) string

func Router(svrType string, fn func(*session.Session) string) {
	if t := strings.TrimSpace(svrType); t != "" {
		router[svrType] = fn
	}
}
