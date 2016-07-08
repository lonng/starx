package network

import (
	"errors"
	"fmt"
	"net"
	"time"

	"encoding/json"
	"github.com/chrislonng/starx/log"
	routelib "github.com/chrislonng/starx/network/route"
	"github.com/chrislonng/starx/network/rpc"
	"github.com/chrislonng/starx/session"
)

var (
	ErrRPCLocal     = errors.New("RPC object must location in different server type")
	ErrSidNotExists = errors.New("sid not exists")
)

// Agent corresponding a user, used for store raw socket information
// only used in package internal, can not accessible by other package
type agent struct {
	id       uint64
	socket   net.Conn
	status   networkStatus
	session  *session.Session
	lastTime int64 // last heartbeat unix time stamp
}

// Create new agent instance
func newAgent(id uint64, conn net.Conn) *agent {
	a := &agent{
		id:       id,
		socket:   conn,
		status:   statusStart,
		lastTime: time.Now().Unix()}
	session := session.NewSession(a)
	a.session = session
	return a
}

// String, implementation for Stringer interface
func (a *agent) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		a.id,
		a.socket.RemoteAddr().String(),
		a.lastTime)
}

// send data to user
func (a *agent) send(data []byte) {
	a.socket.Write(data)
}

func (a *agent) heartbeat() {
	a.lastTime = time.Now().Unix()
}

func (a *agent) close() {
	a.status = statusClosed
	//TODO:FIXED IT
	/*
	   defaultNetService.closeSession(a.session)
	   a.socket.Close()
	*/
}

func (a *agent) ID() uint64 {
	return a.id
}

func (a *agent) Send(data []byte) error {
	_, err := a.socket.Write(data)
	return err
}

func (a *agent) Push(session *session.Session, route string, v interface{}) error {
	data, err := serializer.Serialize(v)
	if err != nil {
		return err
	}

	if appConfig.IsFrontend {
		return defaultNetService.Push(session, route, data)
	}

	rs, err := defaultNetService.getAcceptor(session.Entity.ID())
	if err != nil {
		log.Error(err.Error())
		return err
	}

	sid, ok := rs.b2fMap[session.Id]
	if !ok {
		log.Error("sid not exists")
		return ErrSidNotExists
	}

	resp := rpc.Response{
		Route: route,
		Kind:  rpc.HandlerPush,
		Data:  data,
		Sid:   sid,
	}
	return writeResponse(rs, &resp)
}

// Response message to session
func (a *agent) Response(session *session.Session, v interface{}) error {
	data, err := serializer.Serialize(v)
	if err != nil {
		return err
	}

	if appConfig.IsFrontend {
		return defaultNetService.Response(session, data)
	}

	rs, err := defaultNetService.getAcceptor(session.Entity.ID())
	if err != nil {
		log.Error(err.Error())
		return err
	}

	sid, ok := rs.b2fMap[session.Id]
	if !ok {
		log.Error("sid not exists")
		return ErrSidNotExists
	}
	resp := rpc.Response{
		Kind: rpc.HandlerResponse,
		Data: data,
		Sid:  sid,
	}
	return writeResponse(rs, &resp)
}

func (a *agent) AsyncCall(session *session.Session, route string, args ...interface{}) error {
	r, err := routelib.Decode(route)
	if err != nil {
		return err
	}

	if appConfig.Type == r.ServerType {
		return ErrRPCLocal
	}

	encodeArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}
	_, err = Remote.request(rpc.User, r, session, encodeArgs)
	return err
}

func (a *agent) Call(session *session.Session, route string, args ...interface{}) ([]byte, error) {
	r, err := routelib.Decode(route)
	if err != nil {
		return nil, err
	}

	if appConfig.Type == r.ServerType {
		return nil, ErrRPCLocal
	}

	encodeArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	return Remote.request(rpc.User, r, session, encodeArgs)
}
