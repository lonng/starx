package starx

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/packet"
	routelib "github.com/chrislonng/starx/route"
	"github.com/chrislonng/starx/session"
)

var (
	ErrRPCLocal          = errors.New("RPC object must location in different server type")
	ErrSidNotExists      = errors.New("sid not exists")
	ErrSendChannelClosed = errors.New("agent send channel closed")
)

// Agent corresponding a user, used for store raw socket information
// only used in package internal, can not accessible by other package
type agent struct {
	id         int64
	socket     net.Conn
	status     networkStatus
	session    *session.Session
	sendBuffer chan []byte
	recvBuffer chan *packet.Packet
	ending     chan bool
	lastTime   int64 // last heartbeat unix time stamp
}

// Create new agent instance
func newAgent(conn net.Conn) *agent {
	a := &agent{
		socket:     conn,
		status:     statusStart,
		lastTime:   time.Now().Unix(),
		sendBuffer: make(chan []byte, packetBufferSize),
		recvBuffer: make(chan *packet.Packet, packetBufferSize),
		ending:     make(chan bool, 1),
	}
	s := session.NewSession(a)
	a.session = s
	a.id = s.ID

	return a
}

// String, implementation for Stringer interface
func (a *agent) String() string {
	return fmt.Sprintf("Id=%d, Remote=%s, LastTime=%d",
		a.id,
		a.socket.RemoteAddr().String(),
		a.lastTime)
}

func (a *agent) heartbeat() {
	a.lastTime = time.Now().Unix()
}

func (a *agent) Close() {
	if a.status == statusClosed {
		return
	}

	a.status = statusClosed
	log.Debugf("Session closed, Id=%d, IP=%s", a.session.ID, a.socket.RemoteAddr())

	a.ending <- true

	// close all channel
	close(a.ending)
	close(a.recvBuffer)
	close(a.sendBuffer)

	defaultNetService.closeSession(a.session)
	a.socket.Close()
}

func (a *agent) ID() int64 {
	return a.id
}

func (a *agent) Send(data []byte) error {
	if a.status < statusClosed {
		a.sendBuffer <- data
		return nil
	}
	return ErrSendChannelClosed
}

func (a *agent) Push(session *session.Session, route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("Type=Push, UID=%d, Route=%s, Data=%+v", session.Uid, route, v)

	return defaultNetService.push(session, route, data)
}

// Response message to session
func (a *agent) Response(session *session.Session, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("Type=Response, UID=%d, Data=%+v", session.Uid, v)

	return defaultNetService.response(session, data)
}

func (a *agent) Call(session *session.Session, route string, reply interface{}, args ...interface{}) error {
	r, err := routelib.Decode(route)
	if err != nil {
		return err
	}

	if App.Config.Type == r.ServerType {
		return ErrRPCLocal
	}

	data, err := gobEncode(args...)
	if err != nil {
		return err
	}

	ret, err := cluster.Call(rpc.User, r, session, data)
	if err != nil {
		return err
	}

	return gobDecode(reply, ret)
}
