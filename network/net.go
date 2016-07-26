package network

import (
	"errors"
	"net"
	"sync"

	"fmt"
	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network/message"
	"github.com/chrislonng/starx/network/packet"
	"github.com/chrislonng/starx/session"
)

const sessionClosedRoute = "__Session.Closed"

var (
	ErrSessionOnNotify = errors.New("current session working on notify mode")
	ErrSessionNotFound = errors.New("session not found")
)

var (
	heartbeatPacket, _ = packet.Pack(&packet.Packet{Type: packet.Heartbeat})
	defaultNetService  = NewNetService()
)

type netService struct {
	agentMapLock sync.RWMutex      // protect agentMap
	agentMap     map[uint64]*agent // agents map

	acceptorUidLock sync.RWMutex         // protect acceptorUid
	acceptorUid     uint64               // acceptor unique id
	acceptorMapLock sync.RWMutex         // protect acceptorMap
	acceptorMap     map[uint64]*acceptor // acceptor map

	sessionCloseCbLock sync.RWMutex             // protect sessionCloseCb
	sessionCloseCb     []func(*session.Session) // callback on session closed
}

// Create new netservive
func NewNetService() *netService {
	return &netService{
		agentMap:    make(map[uint64]*agent),
		acceptorUid: 1,
		acceptorMap: make(map[uint64]*acceptor),
	}
}

// Create agent via netService
func (net *netService) createAgent(conn net.Conn) *agent {
	a := newAgent(conn)
	// add to maps
	net.agentMapLock.Lock()
	net.agentMap[a.id] = a
	net.agentMapLock.Unlock()
	return a
}

// get agent by session id
func (net *netService) agent(id uint64) (*agent, error) {
	net.agentMapLock.RLock()
	defer net.agentMapLock.RUnlock()

	a, ok := net.agentMap[id]
	if !ok {
		return nil, errors.New("agent id: " + string(id) + " not exists!")
	}

	return a, nil
}

// Create acceptor via netService
func (net *netService) createAcceptor(conn net.Conn) *acceptor {
	net.acceptorUidLock.Lock()
	id := net.acceptorUid
	net.acceptorUid++
	net.acceptorUidLock.Unlock()
	a := newAcceptor(id, conn)
	// add to maps
	net.acceptorMapLock.Lock()
	net.acceptorMap[id] = a
	net.acceptorMapLock.Unlock()
	return a
}

func (net *netService) acceptor(id uint64) (*acceptor, error) {
	net.acceptorMapLock.RLock()
	defer net.acceptorMapLock.RUnlock()

	rs, ok := net.acceptorMap[id]
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
		log.Error(err.Error())
		return err
	}
	p := packet.Packet{
		Type:   packet.Data,
		Length: len(m),
		Data:   m,
	}
	ep, err := p.Pack()
	if err != nil {
		log.Error(err.Error())
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
		log.Error(err.Error())
		return err
	}
	p := packet.Packet{
		Type:   packet.Data,
		Length: len(m),
		Data:   m,
	}
	ep, err := p.Pack()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	net.send(session, ep)
	return nil
}

// Broadcast message to all sessions
// Message level method
// call by all package, the last argument was packaged message
func (net *netService) Broadcast(route string, data []byte) {
	if appConfig.IsFrontend {
		for _, s := range net.agentMap {
			net.Push(s.session, route, data)
		}
	}
}

// Multicast message to special agent ids
func (net *netService) Multicast(aids []uint64, route string, data []byte) {
	net.agentMapLock.RLock()
	defer net.agentMapLock.RUnlock()

	for _, aid := range aids {
		if agent, ok := net.agentMap[aid]; ok && agent != nil {
			net.Push(agent.session, route, data)
		}
	}
}

func (net *netService) Session(sid uint64) (*session.Session, error) {
	net.agentMapLock.RLock()
	defer net.agentMapLock.RUnlock()

	a, ok := net.agentMap[sid]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return a.session, nil
}

// Close session
func (net *netService) closeSession(session *session.Session) {
	net.sessionCloseCbLock.RLock()
	if len(net.sessionCloseCb) > 0 {
		for _, cb := range net.sessionCloseCb {
			if cb != nil {
				cb(session)
			}
		}
	}
	net.sessionCloseCbLock.RUnlock()

	if appConfig.IsFrontend {
		net.agentMapLock.Lock()
		if agent, ok := net.agentMap[session.Entity.ID()]; ok && (agent != nil) {
			delete(net.agentMap, session.Entity.ID())
		}
		net.agentMapLock.Unlock()
		// notify all backend server, current session has been closed.
		cluster.SessionClosed(session)
	} else {
		net.acceptorMapLock.RLock()
		if acceptor, ok := net.acceptorMap[session.Entity.ID()]; ok && (acceptor != nil) {
			delete(acceptor.sessionMap, session.Id)
			if fid, ok := acceptor.b2fMap[session.Id]; ok {
				delete(acceptor.b2fMap, session.Id)
				delete(acceptor.f2bMap, fid)
			}
		}
		net.acceptorMapLock.RUnlock()
	}
}

func (net *netService) removeAcceptor(a *acceptor) {
	net.acceptorMapLock.Lock()
	delete(net.acceptorMap, a.id)
	net.acceptorMapLock.Unlock()
}

// Send heartbeat packet
func (net *netService) heartbeat() {
	if !appConfig.IsFrontend || net.agentMap == nil {
		return
	}
	for _, agent := range net.agentMap {
		if agent.status == statusWorking {
			if err := agent.Send(heartbeatPacket); err != nil {
				agent.close()
				continue
			}
			agent.heartbeat()
		}
	}
}

// Dump all agents
func (net *netService) dumpAgents() {
	net.agentMapLock.RLock()
	defer net.agentMapLock.RUnlock()
	log.Info("current agent count: %d", len(net.agentMap))
	for _, ses := range net.agentMap {
		log.Info("session: " + ses.String())
	}
}

// Dump all acceptor
func (net *netService) dumpAcceptor() {
	net.acceptorMapLock.RLock()
	defer net.acceptorMapLock.RUnlock()
	log.Info("current acceptor count: %d", len(net.acceptorMap))
	for _, ses := range net.acceptorMap {
		log.Info("session: " + ses.String())
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
