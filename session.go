package starx

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type SessionStatus byte

const (
	_ SessionStatus = iota
	SS_START
	SS_HANDSHAKING
	SS_WORKING
	SS_CLOSED
)

// This session type as argument pass to Handler method, is a proxy session
// for frontend session in frontend server or backend session in backend
// server, correspond frontend session or backend session id as a field
// will be store in type instance
type Session struct {
	Id           int           // session global uniqe id
	Uid          int           // binding user id
	reqId        uint          // last request id
	status       SessionStatus // session current time
	lastTime     int64         // last heartbeat time
	rawSessionId uint64        // raw session id, frontendSession in frontend server, or backendSession in backend server
}

// Session for frontend server, used for store raw socket information
// only used in package internal, can not accessible by other package
type frontendSession struct {
	id          uint64
	socket      net.Conn
	status      SessionStatus
	userSession *Session
	lastTime    int64 // last heartbeat unix time stamp
}

// Session for backend server, used for store raw socket information
// only used in package internal, can not accessible by other package
type backendSession struct {
	id          uint64
	socket      net.Conn
	status      SessionStatus
	sessionMap  map[uint64]*Session
	userSession *Session
	lastTime    int64 // last heartbeat unix time stamp
}

// Create new session instance
func newSession() *Session {
	return &Session{
		Id:       connectionService.getNewSessionUUID(),
		status:   SS_START,
		lastTime: time.Now().Unix()}
}

// Create new frontend session instance
func newFrontendSession(id uint64, conn net.Conn) *frontendSession {
	fs := &frontendSession{
		id:       id,
		socket:   conn,
		status:   SS_START,
		lastTime: time.Now().Unix()}
	session := newSession()
	session.rawSessionId = fs.id
	fs.userSession = session
	return fs
}

// Create new backend session instance
func newBackendSession(id uint64, conn net.Conn) *backendSession {
	return &backendSession{
		id:       id,
		socket:   conn,
		status:   SS_START,
		lastTime: time.Now().Unix()}
}

// Session send packet data
func (session *Session) Send(data []byte) {
	Net.send(session, data)
}

// Push message to session
func (session *Session) Push(route string, data []byte) {
	Net.Push(session, route, data)
}

// Response message to session
func (session *Session) Response(data []byte) {
	Net.Response(session, data)
}

// Implement Stringer interface
func (fs *frontendSession) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		fs.id,
		fs.socket.RemoteAddr().String(),
		fs.lastTime)
}

func (fs *frontendSession) send(data []byte) {
	fs.socket.Write(data)
}

func (fs *frontendSession) heartbeat() {
	fs.lastTime = time.Now().Unix()
}

// Implement Stringer interface
func (bs *backendSession) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		bs.id,
		bs.socket.RemoteAddr().String(),
		bs.lastTime)
}

func (bs *backendSession) send(data []byte) {
	bs.socket.Write(data)
}

func (bs *backendSession) heartbeat() {
	bs.lastTime = time.Now().Unix()
}

func (session *Session) Bind(uid int) {
	if session.Uid > 0 {
		session.Uid = uid
	} else {
		Error("uid invalid")
	}
}

func (session *Session) String() string {
	return fmt.Sprintf("Id: %d, Uid: %d, RemoteAddr: %s",
		session.Id,
		session.Uid)
}

func (session *Session) heartbeat() {
	session.lastTime = time.Now().Unix()
}

func (session *Session) Request(route string, data []byte) {
	ri, err := decodeRouteInfo(string)
	if err != nil {
		Error(err.Error())
		return
	}
	if App.Config.Type == ri.server {
		msg := NewMessage()
		msg.Type = MT_REQUEST
		handler.localProcess(session, ri, msg.encoding())
	} else {
		remote.request(route, session, data)
	}
}

func (session *Session) Notify(route string, data []byte) {
	ri, err := decodeRouteInfo(route)
	if err != nil {
		Error(err.Error())
		return
	}
	if App.Config.Type == ri.server {
		msg := NewMessage()
		msg.Type = MT_NOTIFY
		handler.localProcess(session, ri, msg.encoding())
	} else {
		remote.notify(route, session, data)
	}
}

func (session *Session) Sync(string) {
	//TODO
	//synchronize session setting field to frontend server
}

type SessionService struct {
	sum             sync.RWMutex     // protect SessionUidMaps
	SessionUidMaps  map[int]*Session // uid map sesseion
	sam             sync.RWMutex     // protect sessionAddrMaps
	sessionAddrMaps map[string]*Session
}
