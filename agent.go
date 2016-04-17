package starx

import (
	"fmt"
	"net"
	"time"
)

// Agent corresponding a user, used for store raw socket information
// only used in package internal, can not accessible by other package
type agent struct {
	id       uint64
	socket   net.Conn
	status   networkStatus
	session  *Session
	lastTime int64 // last heartbeat unix time stamp
}

// Create new agent instance
func newAgent(id uint64, conn net.Conn) *agent {
	a := &agent{
		id:       id,
		socket:   conn,
		status:   _STATUS_START,
		lastTime: time.Now().Unix()}
	session := newSession()
	session.entityID = a.id
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
	a.status = _STATUS_CLOSED
	defaultNetService.closeSession(a.session)
	a.socket.Close()
}
