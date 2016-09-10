package starx

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/message"
	"github.com/chrislonng/starx/packet"
	"github.com/chrislonng/starx/session"
)

const sessionClosedRoute = "__Session.Closed"

var (
	ErrSessionOnNotify = errors.New("current session working on notify mode")
	ErrSessionNotFound = errors.New("session not found")
)

var (
	heartbeatPacket, _ = packet.Pack(&packet.Packet{Type: packet.Heartbeat})
	defaultNetService  = newNetService()
)

type netService struct {
	sync.RWMutex
	agents      map[int64]*agent    // agents map
	acceptorUid int64               // acceptor unique id
	acceptors   map[int64]*acceptor // acceptor map

	sessionCloseCbLock sync.RWMutex             // protect sessionCloseCb
	sessionCloseCb     []func(*session.Session) // callback on session closed
}

// Create new net service
func newNetService() *netService {
	return &netService{
		agents:      make(map[int64]*agent),
		acceptorUid: 0,
		acceptors:   make(map[int64]*acceptor),
	}
}

// Create agent via netService
func (net *netService) createAgent(conn net.Conn) *agent {
	a := newAgent(conn)
	// add to maps
	net.Lock()
	defer net.Unlock()

	net.agents[a.id] = a
	return a
}

// get agent by session id
func (net *netService) agent(id int64) (*agent, error) {
	net.RLock()
	defer net.RUnlock()

	a, ok := net.agents[id]
	if !ok {
		return nil, errors.New("agent id: " + string(id) + " not exists!")
	}

	return a, nil
}

// Create acceptor via netService
func (net *netService) createAcceptor(conn net.Conn) *acceptor {
	id := atomic.AddInt64(&net.acceptorUid, 1)
	a := newAcceptor(id, conn)

	// add to maps
	net.Lock()
	defer net.Unlock()

	net.acceptors[id] = a
	return a
}

func (net *netService) acceptor(id int64) (*acceptor, error) {
	net.RLock()
	defer net.RUnlock()

	rs, ok := net.acceptors[id]
	if !ok || rs == nil {
		return nil, errors.New("acceptor id: " + string(id) + " not exists!")
	}

	return rs, nil
}

// Send packet data, call by package internal, the second argument was packaged packet
// if current server is frontend server, send to client by agent, else send to frontend
// server by acceptor
func (net *netService) send(session *session.Session, data []byte) {
	session.Entity.Send(data)
}

// Push message to client
// call by all package, the last argument was packaged message
func (net *netService) Push(session *session.Session, route string, data []byte) error {
	m, err := message.Encode(&message.Message{Type: message.MessageType(message.Push), Route: route, Data: data})
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	p := packet.Packet{
		Type:   packet.Data,
		Length: len(m),
		Data:   m,
	}
	ep, err := p.Pack()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	net.send(session, ep)
	return nil
}

// Response message to client
// call by all package, the last argument was packaged message
func (net *netService) Response(session *session.Session, data []byte) error {
	// current message is notify message, can not response
	if session.LastID <= 0 {
		return ErrSessionOnNotify
	}
	m, err := message.Encode(&message.Message{
		Type: message.MessageType(message.Response),
		ID:   session.LastID,
		Data: data,
	})
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	p := packet.Packet{
		Type:   packet.Data,
		Length: len(m),
		Data:   m,
	}
	ep, err := p.Pack()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	net.send(session, ep)
	return nil
}

// Broadcast message to all sessions
// Message level method
// call by all package, the last argument was packaged message
func (net *netService) Broadcast(route string, data []byte) {
	if App.Config.IsFrontend {
		for _, s := range net.agents {
			net.Push(s.session, route, data)
		}
	}
}

// Multicast message to special agent ids
func (net *netService) Multicast(aids []int64, route string, data []byte) {
	net.RLock()
	defer net.RUnlock()

	for _, aid := range aids {
		if agent, ok := net.agents[aid]; ok && agent != nil {
			net.Push(agent.session, route, data)
		}
	}
}

func (net *netService) Session(sid int64) (*session.Session, error) {
	net.RLock()
	defer net.RUnlock()

	a, ok := net.agents[sid]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return a.session, nil
}

// Close session
func (net *netService) closeSession(session *session.Session) {
	net.sessionCloseCbLock.RLock()
	for _, cb := range net.sessionCloseCb {
		if cb != nil {
			cb(session)
		}
	}
	net.sessionCloseCbLock.RUnlock()

	net.Lock()
	defer net.Unlock()

	if App.Config.IsFrontend {
		if agent, ok := net.agents[session.Entity.ID()]; ok && (agent != nil) {
			delete(net.agents, session.Entity.ID())
		}
		// notify all backend server, current session has been closed.
		cluster.SessionClosed(session)
	} else {
		if acceptor, ok := net.acceptors[session.Entity.ID()]; ok && (acceptor != nil) {
			delete(acceptor.sessionMap, session.ID)
			if fid, ok := acceptor.b2fMap[session.ID]; ok {
				delete(acceptor.b2fMap, session.ID)
				delete(acceptor.f2bMap, fid)
			}
		}
	}
}

func (net *netService) removeAcceptor(a *acceptor) {
	net.Lock()
	defer net.Unlock()

	delete(net.acceptors, a.id)
}

// Send heartbeat packet
func (net *netService) heartbeat() {
	if !App.Config.IsFrontend || net.agents == nil {
		return
	}
	log.Debugf("heartbeat")
	for _, agent := range net.agents {
		if agent.status != statusWorking {
			continue
		}

		if err := agent.Send(heartbeatPacket); err != nil {
			agent.close()
			continue
		}
		agent.heartbeat()
	}
}

// Dump all agents
func (net *netService) dumpAgents() {
	net.RLock()
	defer net.RUnlock()

	log.Infof("current agent count: %d", len(net.agents))
	for _, ses := range net.agents {
		log.Infof("session: " + ses.String())
	}
}

// Dump all acceptor
func (net *netService) dumpAcceptor() {
	net.RLock()
	defer net.RUnlock()

	log.Infof("current acceptor count: %d", len(net.acceptors))
	for _, ses := range net.acceptors {
		log.Infof("session: " + ses.String())
	}
}

func (net *netService) sessionClosedCallback(cb func(*session.Session)) {
	net.sessionCloseCbLock.Lock()
	defer net.sessionCloseCbLock.Unlock()

	net.sessionCloseCb = append(net.sessionCloseCb, cb)
}

// Callback when session closed
// Waring: session has closed,
func OnSessionClosed(cb func(*session.Session)) {
	defaultNetService.sessionClosedCallback(cb)
}
