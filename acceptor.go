package starx

import (
	"fmt"
	"net"
	"time"
)

// Acceptor corresponding a front server, used for store raw socket
// information.
// only used in package internal, can not accessible by other package
type acceptor struct {
	id         uint64
	socket     net.Conn
	status     networkStatus
	sessionMap map[uint64]*Session // backend sessions
	f2bMap     map[uint64]uint64   // frontend session id -> backend session id map
	b2fMap     map[uint64]uint64   // backend session id -> frontend session id map
	lastTime   int64               // last heartbeat unix time stamp
}

// Create new backend session instance
func newAcceptor(id uint64, conn net.Conn) *acceptor {
	return &acceptor{
		id:         id,
		socket:     conn,
		status:     _STATUS_START,
		sessionMap: make(map[uint64]*Session),
		f2bMap:     make(map[uint64]uint64),
		b2fMap:     make(map[uint64]uint64),
		lastTime:   time.Now().Unix()}
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

func (a *acceptor) GetUserSession(sid uint64) *Session {
	if bsid, ok := a.f2bMap[sid]; ok && bsid > 0 {
		return a.sessionMap[bsid]
	} else {
		s := newSession()
		s.entityID = a.id
		a.sessionMap[s.Id] = s
		a.f2bMap[sid] = s.Id
		a.b2fMap[s.Id] = sid
		return s
	}
}

func (a *acceptor) close() {
	a.status = _STATUS_CLOSED
	for _, session := range a.sessionMap {
		defaultNetService.closeSession(session)
	}
	defaultNetService.removeAcceptor(a)
	a.socket.Close()
}
