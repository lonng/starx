/*
 Network handle
*/
package starx

import (
	"errors"
	"net"
	"sync"
)

type netService struct {
	agentUidLock       sync.RWMutex         // protect agentUid
	agentUid           uint64               // agent unique id
	agentMapLock       sync.RWMutex         // protect agentMap
	agentMap           map[uint64]*agent    // agents map
	acceptorUidLock    sync.RWMutex         // protect acceptorUid
	acceptorUid        uint64               // acceptor unique id
	acceptorMapLock    sync.RWMutex         // protect acceptorMap
	acceptorMap        map[uint64]*acceptor // acceptor map
	sessionCloseCbLock sync.RWMutex         // protect sessionCloseCb
	sessionCloseCb     []func(*Session)     // callback on session closed
}

// Create new netservive
func newNetService() *netService {
	return &netService{
		agentUid:    1,
		agentMap:    make(map[uint64]*agent),
		acceptorUid: 1,
		acceptorMap: make(map[uint64]*acceptor)}
}

// Create agent via netService
func (net *netService) createAgent(conn net.Conn) *agent {
	net.agentUidLock.Lock()
	id := net.agentUid
	net.agentUid++
	net.agentUidLock.Unlock()
	a := newAgent(id, conn)
	// add to maps
	net.agentMapLock.Lock()
	net.agentMap[id] = a
	net.agentMapLock.Unlock()
	return a
}

// get agent by session id
func (net *netService) getAgent(sid uint64) (*agent, error) {
	if a, ok := net.agentMap[sid]; ok && a != nil {
		return a, nil
	} else {
		return nil, errors.New("agent id: " + string(sid) + " not exists!")
	}
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

func (net *netService) getAcceptor(sid uint64) (*acceptor, error) {
	if rs, ok := net.acceptorMap[sid]; ok && rs != nil {
		return rs, nil
	} else {
		return nil, errors.New("acceptor id: " + string(sid) + " not exists!")
	}
}

// Send packet data, call by package internal, the second argument was packaged packet
// if current server is frontend server, send to client by agent, else send to frontend
// server by acceptor
func (net *netService) send(session *Session, data []byte) {
	if App.Config.IsFrontend {
		if fs, ok := net.agentMap[session.entityID]; ok && (fs != nil) {
			go fs.send(data)
		}
	} else {
		if bs, ok := net.acceptorMap[session.entityID]; ok && (bs != nil) {
			go bs.send(data)
		}
	}
}

// Push message to client
// call by all package, the last argument was packaged message
func (net *netService) Push(session *Session, route string, data []byte) {
	m := encodeMessage(&message{kind: messageType(_MT_PUSH), route: route, body: data})
	net.send(session, pack(packetType(_PACKET_DATA), m))
}

// Response message to client
// call by all package, the last argument was packaged message
func (net *netService) Response(session *Session, data []byte) {
	// current message is notify message, can not response
	if session.reqId <= 0 {
		return
	}
	m := encodeMessage(&message{kind: messageType(_MT_RESPONSE), id: session.reqId, body: data})
	net.send(session, pack(packetType(_PACKET_DATA), m))
}

// Broadcast message to all sessions
// Message level method
// call by all package, the last argument was packaged message
func (net *netService) Broadcast(route string, data []byte) {
	if App.Config.IsFrontend {
		for _, s := range net.agentMap {
			net.Push(s.session, route, data)
		}
	}
}

// Multicast message to special agent ids
func (net *netService) Multicast(aids []uint64, route string, data []byte) {
	for _, aid := range aids {
		if agent, ok := net.agentMap[aid]; ok && agent != nil {
			net.Push(agent.session, route, data)
		}
	}
}

// Close session
func (net *netService) closeSession(session *Session) {
	// TODO: notify all backend server, current session has closed.
	// session close callback
	net.sessionCloseCbLock.RLock()
	if len(net.sessionCloseCb) > 0 {
		for _, cb := range net.sessionCloseCb {
			if cb != nil {
				cb(session)
			}
		}
	}
	net.sessionCloseCbLock.RUnlock()
	if App.Config.IsFrontend {
		net.agentMapLock.Lock()
		if agent, ok := net.agentMap[session.entityID]; ok && (agent != nil) {
			delete(net.agentMap, session.entityID)
		}
		net.agentMapLock.Unlock()
		defaultNetService.dumpAgents()
	} /* else {
		net.acceptorMapLock.RLock()
		if acceptor, ok := net.acceptorMap[session.entityID]; ok && (acceptor != nil) {
			// TODO: FIXED IT
			// backend session close should not cause acceptor remove from acceptor map
		}
		net.acceptorMapLock.RUnlock()
		defaultNetService.dumpAcceptor()
	}*/
}

func (net *netService) removeAcceptor(a *acceptor) {
	net.acceptorMapLock.Lock()
	delete(net.acceptorMap, a.id)
	net.acceptorMapLock.Unlock()
}

// Send heartbeat packet
func (net *netService) heartbeat() {
	if !App.Config.IsFrontend || net.agentMap == nil {
		return
	}
	for _, session := range net.agentMap {
		if session.status == _STATUS_WORKING {
			session.send(pack(_PACKET_HEARTBEAT, nil))
			session.heartbeat()
		}
	}
}

// Dump all agents
func (net *netService) dumpAgents() {
	net.agentMapLock.RLock()
	defer net.agentMapLock.RUnlock()
	Info("current agent count: %d", len(net.agentMap))
	for _, ses := range net.agentMap {
		Info("session: " + ses.String())
	}
}

// Dump all acceptor
func (net *netService) dumpAcceptor() {
	net.acceptorMapLock.RLock()
	defer net.acceptorMapLock.RUnlock()
	Info("current acceptor count: %d", len(net.acceptorMap))
	for _, ses := range net.acceptorMap {
		Info("session: " + ses.String())
	}
}

func (net *netService) sessionClosedCallback(cb func(*Session)) {
	net.sessionCloseCbLock.Lock()
	defer net.sessionCloseCbLock.Unlock()
	net.sessionCloseCb = append(net.sessionCloseCb, cb)
}

// Callback when session closed
// Waring: session has closed,
func OnSessionClosed(cb func(*Session)) {
	defaultNetService.sessionClosedCallback(cb)
}
