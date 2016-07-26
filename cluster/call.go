package cluster

import (
	"errors"

	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network/route"
	"github.com/chrislonng/starx/session"
)

var (
	sessionClosedRoute = &route.Route{Service: "__Session", Method: "Closed"}
)

func AsyncCall(route *route.Route, session *session.Session, args ...interface{}) {

}

// Client send request
// First argument is namespace, can be set `user` or `sys`
func Call(rpcKind rpc.RpcKind, route *route.Route, session *session.Session, args []byte) ([]byte, error) {
	client, err := ClientByType(route.ServerType, session)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	reply := new([]byte)
	err = client.Call(rpcKind, route.Service, route.Method, session.Entity.ID(), reply, args)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return *reply, nil
}

func SessionClosed(session *session.Session) {
	for _, t := range svrTypes {
		client, err := ClientByType(t, session)
		if err != nil {
			continue
		}
		err = client.Call(rpc.Sys, sessionClosedRoute.Service, sessionClosedRoute.Method, session.Entity.ID(), nil, nil)
	}
}
