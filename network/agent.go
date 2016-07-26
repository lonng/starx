package network

import (
	"errors"
	"fmt"
	"net"
	"time"

	"encoding/json"
	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/cluster/rpc"
	routelib "github.com/chrislonng/starx/network/route"
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
func newAgent(conn net.Conn) *agent {
	a := &agent{
		socket:   conn,
		status:   statusStart,
		lastTime: time.Now().Unix()}
	session := session.NewSession(a)
	a.session = session
	a.id = session.Id
	return a
}

// String, implementation for Stringer interface
func (a *agent) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		a.id,
		a.socket.RemoteAddr().String(),
		a.lastTime)
}

func (a *agent) heartbeat() {
	a.lastTime = time.Now().Unix()
}

func (a *agent) close() {
	a.status = statusClosed
	defaultNetService.closeSession(a.session)
	a.socket.Close()
}

func (a *agent) ID() uint64 {
	return a.id
}

func (a *agent) Send(data []byte) error {
	_, err := a.socket.Write(data)
	return err
}

func (a *agent) Push(session *session.Session, route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}
	return defaultNetService.Push(session, route, data)
}

// Response message to session
func (a *agent) Response(session *session.Session, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	return defaultNetService.Response(session, data)
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

	return cluster.Call(rpc.User, r, session, encodeArgs)
}

// TODO: implement
func (a *acceptor) Sync(data map[string]interface{}) error {
	return nil
}
