package starx

import (
	"net"
	"time"
	"fmt"
)
// Session for frontend server, used for store raw socket information
// only used in package internal, can not accessible by other package
type handlerSession struct {
	id          uint64
	socket      net.Conn
	status      SessionStatus
	userSession *Session
	lastTime    int64 // last heartbeat unix time stamp
}

// Create new frontend session instance
func newHandlerSession(id uint64, conn net.Conn) *handlerSession {
	hs := &handlerSession{
		id:       id,
		socket:   conn,
		status:   SS_START,
		lastTime: time.Now().Unix()}
	session := newSession()
	session.rawSessionId = hs.id
	hs.userSession = session
	return hs
}

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