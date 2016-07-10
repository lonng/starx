package network

import (
	"fmt"
	"net"
	"time"

	"encoding/json"
	"github.com/chrislonng/starx/log"
	routelib "github.com/chrislonng/starx/network/route"
	"github.com/chrislonng/starx/session"
	"github.com/chrislonng/starx/cluster/rpc"
)

// Acceptor corresponding a front server, used for store raw socket
// information.
// only used in package internal, can not accessible by other package
type acceptor struct {
	id         uint64
	socket     net.Conn
	status     networkStatus
	sessionMap map[uint64]*session.Session // backend sessions
	f2bMap     map[uint64]uint64           // frontend session id -> backend session id map
	b2fMap     map[uint64]uint64           // backend session id -> frontend session id map
	lastTime   int64                       // last heartbeat unix time stamp
}

// Create new backend session instance
func newAcceptor(id uint64, conn net.Conn) *acceptor {
	return &acceptor{
		id:         id,
		socket:     conn,
		status:     statusStart,
		sessionMap: make(map[uint64]*session.Session),
		f2bMap:     make(map[uint64]uint64),
		b2fMap:     make(map[uint64]uint64),
		lastTime:   time.Now().Unix(),
	}
}

// String implement Stringer interface
func (a *acceptor) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		a.id,
		a.socket.RemoteAddr().String(),
		a.lastTime)
}

func (a *acceptor) send(data []byte) {
	a.socket.Write(data)
}

func (a *acceptor) heartbeat() {
	a.lastTime = time.Now().Unix()
}

func (a *acceptor) Session(sid uint64) *session.Session {
	if bsid, ok := a.f2bMap[sid]; ok && bsid > 0 {
		return a.sessionMap[bsid]
	}
	s := session.NewSession(a)
	a.sessionMap[s.Id] = s
	a.f2bMap[sid] = s.Id
	a.b2fMap[s.Id] = sid
	return s
}

func (a *acceptor) close() {
	a.status = statusClosed
	//TODO:FIXED IT
	/*
		for _, session := range a.sessionMap {
			defaultNetService.closeSession(session)
		}
		defaultNetService.removeAcceptor(a)
		a.socket.Close()
	*/
}

func (a *acceptor) ID() uint64 {
	return a.id
}

func (a *acceptor) Send(data []byte) error {
	_, err := a.socket.Write(data)
	return err
}

func (a *acceptor) Push(session *session.Session, route string, v interface{}) error {
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
func (a *acceptor) Response(session *session.Session, v interface{}) error {
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

func (a *acceptor) AsyncCall(session *session.Session, route string, args ...interface{}) error {
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

func (a *acceptor) Call(session *session.Session, route string, args ...interface{}) ([]byte, error) {
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

// TODO: implement
func (a *agent) Sync(data map[string]interface{}) error {
	return nil
}
