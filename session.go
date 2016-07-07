package starx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network"
	"github.com/chrislonng/starx/network/rpc"
	"time"
)

type networkStatus byte

const (
	_ networkStatus = iota
	statusStart
	statusHandshake
	statusWorking
	statusClosed
)

var (
	ErrRPCLocal     = errors.New("RPC object must location in different server type")
	ErrSidNotExists = errors.New("sid not exists")
	ErrIllegalUID   = errors.New("illegal uid")
)

// This session type as argument pass to Handler method, is a proxy session
// for frontend session in frontend server or backend session in backend
// server, correspond frontend session or backend session id as a field
// will be store in type instance
//
// This is user sessions, not contain raw sockets information
type Session struct {
	Id       uint64 // session global uniqe id
	Uid      uint64 // binding user id
	reqId    uint   // last request id
	lastTime int64  // last heartbeat time
	entityID uint64 // raw session id, frontendSession in frontend server, or backendSession in backend server
}

// Create new session instance
func newSession() *Session {
	return &Session{
		Id:       connections.getNewSessionUUID(),
		lastTime: time.Now().Unix()}
}

// Session send packet data
func (session *Session) Send(data []byte) {
	defaultNetService.send(session, data)
}

// Push message to session
func (session *Session) Push(route string, data []byte) error {
	if App.Config.IsFrontend {
		return defaultNetService.Push(session, route, data)
	}

	rs, err := defaultNetService.getAcceptor(session.entityID)
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
func (session *Session) Response(data []byte) error {
	if App.Config.IsFrontend {
		return defaultNetService.Response(session, data)
	}

	rs, err := defaultNetService.getAcceptor(session.entityID)
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

func (session *Session) Bind(uid uint64) error {
	if uid < 1 {
		log.Error("uid invalid: %d", uid)
		return ErrIllegalUID
	}
	session.Uid = uid
	return nil
}

func (session *Session) String() string {
	return fmt.Sprintf("Id: %d, Uid: %d", session.Id, session.Uid)
}

func (session *Session) AsyncRPC(route string, args ...interface{}) error {
	r, err := network.DecodeRoute(route)
	if err != nil {
		return err
	}

	if App.Config.Type == r.ServerType {
		return ErrRPCLocal
	}

	encodeArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}
	_, err = remote.request(rpc.UserRpc, r, session, encodeArgs)
	return err
}

func (session *Session) RPC(route string, args ...interface{}) ([]byte, error) {
	r, err := network.DecodeRoute(route)
	if err != nil {
		return nil, err
	}

	if App.Config.Type == r.ServerType {
		return nil, ErrRPCLocal
	}

	encodeArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	return remote.request(rpc.UserRpc, r, session, encodeArgs)
}

// Sync session setting to frontend server
func (session *Session) Sync(string) {
	//TODO
	//synchronize session setting field to frontend server
}

// Sync all settings to frontend server
func (session *Session) SyncAll() {
}
