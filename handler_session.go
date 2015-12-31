package starx

import (
	"fmt"
	"net"
	"time"
)

// Session for frontend server, used for store raw socket information
// only used in package internal, can not accessible by other package
type handlerSession struct {
	id          uint64
	socket      net.Conn
	status      sessionStatus
	userSession *Session
	lastTime    int64 // last heartbeat unix time stamp
}

// Create new frontend session instance
func newHandlerSession(id uint64, conn net.Conn) *handlerSession {
	hs := &handlerSession{
		id:       id,
		socket:   conn,
		status:   _SS_START,
		lastTime: time.Now().Unix()}
	session := newSession()
	session.rawSessionId = hs.id
	hs.userSession = session
	return hs
}

// String
// Implement Stringer interface
func (hs *handlerSession) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		hs.id,
		hs.socket.RemoteAddr().String(),
		hs.lastTime)
}

func (hs *handlerSession) send(data []byte) {
	hs.socket.Write(data)
}

func (hs *handlerSession) heartbeat() {
	hs.lastTime = time.Now().Unix()
}

func (hs *handlerSession) close() {
	hs.status = _SS_CLOSED
	netService.closeSession(hs.userSession)
}
