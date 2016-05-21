package starx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chrislonng/starx/rpc"
	"time"
)

type networkStatus byte

const (
	_ networkStatus = iota
	_STATUS_START
	_STATUS_HANDSHAKING
	_STATUS_WORKING
	_STATUS_CLOSED
)

var (
	ErrRPCLocal = errors.New("RPC object must location in different server type")
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
func (session *Session) Push(route string, data []byte) {
	if App.Config.IsFrontend {
		defaultNetService.Push(session, route, data)
	} else {
		rs, err := defaultNetService.getAcceptor(session.entityID)
		if err != nil {
			Error(err.Error())
		} else {
			sid, ok := rs.b2fMap[session.Id]
			if !ok {
				Error("sid not exists")
				return
			}
			resp := rpc.Response{}
			resp.Route = route
			resp.Kind = rpc.HandlerPush
			resp.Data = data
			resp.Sid = sid
			writeResponse(rs, &resp)
		}
	}
}

// Response message to session
func (session *Session) Response(data []byte) {
	if App.Config.IsFrontend {
		defaultNetService.Response(session, data)
	} else {
		rs, err := defaultNetService.getAcceptor(session.entityID)
		if err != nil {
			Error(err.Error())
		} else {
			sid, ok := rs.b2fMap[session.Id]
			if !ok {
				Error("sid not exists")
				return
			}
			resp := rpc.Response{}
			resp.Kind = rpc.HandlerResponse
			resp.Data = data
			resp.Sid = sid
			writeResponse(rs, &resp)
		}
	}
}

func (session *Session) Bind(uid uint64) {
	if uid > 0 {
		session.Uid = uid
	} else {
		Error("uid invalid: %d", uid)
	}
}

func (session *Session) String() string {
	return fmt.Sprintf("Id: %d, Uid: %d", session.Id, session.Uid)
}

func (session *Session) AsyncRPC(route string, args ...interface{}) error {
	ri, err := decodeRouteInfo(route)
	if err != nil {
		return err
	}
	encodeArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}
	if App.Config.Type == ri.serverType {
		return ErrRPCLocal
	} else {
		remote.request(rpc.UserRpc, ri, session, encodeArgs)
		return nil
	}
}

func (session *Session) RPC(route string, args ...interface{}) ([]byte, error) {
	ri, err := decodeRouteInfo(route)
	if err != nil {
		return nil, err
	}
	encodeArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	if App.Config.Type == ri.serverType {
		return nil, ErrRPCLocal
	} else {
		return remote.request(rpc.UserRpc, ri, session, encodeArgs)
	}
}

// Sync session setting to frontend server
func (session *Session) Sync(string) {
	//TODO
	//synchronize session setting field to frontend server
}

// Sync all settings to frontend server
func (session *Session) SyncAll() {
}
