package starx

import (
	"net"
	"time"
	"fmt"
)

// Session for backend server, used for store raw socket information
// only used in package internal, can not accessible by other package
type remoteSession struct {
	id            uint64
	socket        net.Conn
	status        SessionStatus
	sessionMap    map[uint64]*Session
	fsessionIdMap map[uint64]uint64 // session id map(frontend session id -> backend session id)
	bsessionIdMap map[uint64]uint64 // session id map(backend session id -> frontend session id)
	lastTime      int64 // last heartbeat unix time stamp
}

// Create new backend session instance
func newRemoteSession(id uint64, conn net.Conn) *remoteSession {
	return &remoteSession{
		id:            id,
		socket:        conn,
		status:        SS_START,
		sessionMap:    make(map[uint64]*Session),
		fsessionIdMap: make(map[uint64]uint64),
		bsessionIdMap: make(map[uint64]uint64),
		lastTime:      time.Now().Unix()}
}


// Implement Stringer interface
func (rs *remoteSession) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		rs.id,
		rs.socket.RemoteAddr().String(),
		rs.lastTime)
}

func (rs *remoteSession) send(data []byte) {
	rs.socket.Write(data)
}

func (rs *remoteSession) heartbeat() {
	rs.lastTime = time.Now().Unix()
}

func (rs *remoteSession) GetUserSession(sid uint64) *Session {
	if bsid, ok := rs.fsessionIdMap[sid]; ok && bsid > 0 {
		return rs.sessionMap[bsid]
	} else {
		session := newSession()
		session.rawSessionId = rs.id
		rs.fsessionIdMap[sid] = session.Id
		rs.sessionMap[session.Id] = session
		rs.bsessionIdMap[session.Id] = sid
		return session
	}
}
