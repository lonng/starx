/*
 Network handle
*/
package starx

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type _netService struct {
	agentUidLock    sync.RWMutex         // protect agentUid
	agentUid        uint64               // agent unique id
	agentMapLock    sync.RWMutex         // protect agentMap
	agentMap        map[uint64]*agent    // agents map
	acceptorUidLock sync.RWMutex         // protect acceptorUid
	acceptorUid     uint64               // acceptor unique id
	acceptorMapLock sync.RWMutex         // protect acceptorMap
	acceptorMap     map[uint64]*acceptor // acceptor map
}

// Create new netservive
func newNetService() *_netService {
	return &_netService{
		agentUid:    1,
		agentMap:    make(map[uint64]*agent),
		acceptorUid: 1,
		acceptorMap: make(map[uint64]*acceptor)}
}

// Create agent via netService
func (net *_netService) createAgent(conn net.Conn) *agent {
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
func (net *_netService) getAgent(sid uint64) (*agent, error) {
	if a, ok := net.agentMap[sid]; ok && a != nil {
		return a, nil
	} else {
		return nil, errors.New("agent id: " + string(sid) + " not exists!")
	}
}

// Create acceptor via netService
func (net *_netService) createAcceptor(conn net.Conn) *acceptor {
	net.acceptorUidLock.Lock()
	id := net.agentUid
	net.agentUid++
	net.acceptorUidLock.Unlock()
	a := newAcceptor(id, conn)
	// add to maps
	net.acceptorMapLock.Lock()
	net.acceptorMap[id] = a
	net.acceptorMapLock.Unlock()
	return a
}

func (net *_netService) getAcceptor(sid uint64) (*acceptor, error) {
	if rs, ok := net.acceptorMap[sid]; ok && rs != nil {
		return rs, nil
	} else {
		return nil, errors.New("acceptor id: " + string(sid) + " not exists!")
	}
}

// Send packet data
// call by package internal, the second argument was packaged packet
func (net *_netService) send(session *Session, data []byte) {
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

// Push
// Message level method
// call by all package, the last argument was packaged message
func (net *_netService) Push(session *Session, route string, data []byte) {
	m := encodeMessage(&message{kind: messageType(_MT_PUSH), route: route, body: data})
	net.send(session, pack(packetType(_PACKET_DATA), m))
}

// Response
// Message level method
// call by all package, the last argument was packaged message
func (net *_netService) Response(session *Session, data []byte) {
	// current message is notify message, can not response
	if session.reqId <= 0 {
		return
	}
	m := encodeMessage(&message{kind: messageType(_MT_RESPONSE), id: session.reqId, body: data})
	net.send(session, pack(packetType(_PACKET_DATA), m))
}

// Broadcast
// Push message to all sessions
// Message level method
// call by all package, the last argument was packaged message
func (net *_netService) Broadcast(route string, data []byte) {
	if App.Config.IsFrontend {
		for _, s := range net.agentMap {
			net.Push(s.session, route, data)
		}
	}
}

// TODO
func (net *_netService) Multcast(uids []int, route string, data []byte) {

}

// Close session
func (net *_netService) closeSession(session *Session) {
	if App.Config.IsFrontend {
		if fs, ok := net.agentMap[session.entityID]; ok && (fs != nil) {
			fs.socket.Close()
			net.agentMapLock.Lock()
			delete(net.agentMap, session.entityID)
			net.agentMapLock.Unlock()
		}
		netService.dumpAgents()
	} else {
		if bs, ok := net.acceptorMap[session.entityID]; ok && (bs != nil) {
			bs.socket.Close()
			net.acceptorMapLock.Lock()
			delete(net.acceptorMap, session.entityID)
			net.acceptorMapLock.Unlock()
		}
		netService.dumpAcceptor()
	}
}

// Send heartbeat packet
func (net *_netService) heartbeat() {
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
func (net *_netService) dumpAgents() {
	net.agentMapLock.RLock()
	defer net.agentMapLock.RUnlock()
	Info(fmt.Sprintf("current agent count: %d", len(net.agentMap)))
	for _, ses := range net.agentMap {
		Info("session: " + ses.String())
	}
}

// Dump all acceptor
func (net *_netService) dumpAcceptor() {
	net.acceptorMapLock.RLock()
	defer net.acceptorMapLock.RUnlock()
	Info(fmt.Sprintf("current acceptor count: %d", len(net.acceptorMap)))
	for _, ses := range net.acceptorMap {
		Info("session: " + ses.String())
	}
}
