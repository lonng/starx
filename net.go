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
	fuuidLock    sync.RWMutex               // protect fsessionUUID
	fsessionUUID uint64                     // frontend session uuid
	fsmLock      sync.RWMutex               // protect fsessionMap
	fsessionMap  map[uint64]*handlerSession // frontend id to session map
	buuidLock    sync.RWMutex               // protect bsessionUUID
	bsessionUUID uint64                     // backend session uuid
	bsmLock      sync.RWMutex               // protect bsessionMap
	bsessionMap  map[uint64]*remoteSession  // backend id to session map
}

// Create new netservive
func newNetService() *_netService {
	return &_netService{
		fsessionUUID: 1,
		fsessionMap:  make(map[uint64]*handlerSession),
		bsessionUUID: 1,
		bsessionMap:  make(map[uint64]*remoteSession)}
}

// Create frontend session via netService
func (net *_netService) createHandlerSession(conn net.Conn) *handlerSession {
	net.fuuidLock.Lock()
	id := net.fsessionUUID
	net.fsessionUUID++
	net.fuuidLock.Unlock()
	fs := newHandlerSession(id, conn)
	// add to maps
	net.fsmLock.Lock()
	net.fsessionMap[id] = fs
	net.fsmLock.Unlock()
	return fs
}

func (net *_netService) getHandlerSessionBySid(sid uint64) (*handlerSession, error) {
	if hs, ok := net.fsessionMap[sid]; ok && hs != nil {
		return hs, nil
	} else {
		return nil, errors.New("handler session id " + string(sid) + " not exists!")
	}
}

// Create backend session via netService
func (net *_netService) createRemoteSession(conn net.Conn) *remoteSession {
	net.buuidLock.Lock()
	id := net.fsessionUUID
	net.fsessionUUID++
	net.buuidLock.Unlock()
	bs := newRemoteSession(id, conn)
	// add to maps
	net.bsmLock.Lock()
	net.bsessionMap[id] = bs
	net.bsmLock.Unlock()
	return bs
}

func (net *_netService) getRemoteSessionBySid(sid uint64) (*remoteSession, error) {
	if rs, ok := net.bsessionMap[sid]; ok && rs != nil {
		return rs, nil
	} else {
		return nil, errors.New("remote session id " + string(sid) + " not exists!")
	}
}

// Send packet data
// call by package internal, the second argument was packaged packet
func (net *_netService) send(session *Session, data []byte) {
	if App.Config.IsFrontend {
		if fs, ok := net.fsessionMap[session.rawSessionId]; ok && (fs != nil) {
			go fs.send(data)
		}
	} else {
		if bs, ok := net.bsessionMap[session.rawSessionId]; ok && (bs != nil) {
			go bs.send(data)
		}
	}
}

// Push
// Message level method
// call by all package, the last argument was packaged message
func (net *_netService) Push(session *Session, route string, data []byte) {
	m := encodeMessage(&Message{Type: MessageType(MT_PUSH), Route: route, Body: data})
	net.send(session, pack(PacketType(PACKET_DATA), m))
}

// Response
// Message level method
// call by all package, the last argument was packaged message
func (net *_netService) Response(session *Session, data []byte) {
	// current message is notify message, can not response
	if session.reqId <= 0 {
		return
	}
	m := encodeMessage(&Message{Type: MessageType(MT_RESPONSE), ID: session.reqId, Body: data})
	net.send(session, pack(PacketType(PACKET_DATA), m))
}

// Broadcast
// Push message to all sessions
// Message level method
// call by all package, the last argument was packaged message
func (net *_netService) Broadcast(route string, data []byte) {
	if App.Config.IsFrontend {
		for _, s := range net.fsessionMap {
			net.Push(s.userSession, route, data)
		}
	}
}

// TODO
func (net *_netService) Multcast(uids []int, route string, data []byte) {

}

// Close session
func (net *_netService) closeSession(session *Session) {
	if App.Config.IsFrontend {
		if fs, ok := net.fsessionMap[session.rawSessionId]; ok && (fs != nil) {
			fs.socket.Close()
			net.fsmLock.Lock()
			delete(net.fsessionMap, session.rawSessionId)
			net.fsmLock.Unlock()
		}
		netService.dumpHandlerSessions()
	} else {
		if bs, ok := net.bsessionMap[session.rawSessionId]; ok && (bs != nil) {
			bs.socket.Close()
			net.bsmLock.Lock()
			delete(net.bsessionMap, session.rawSessionId)
			net.bsmLock.Unlock()
		}
		netService.dumpRemoteSessions()
	}
}

// Send heartbeat packet
func (net *_netService) heartbeat() {
	if App.Config.IsFrontend {
		for _, session := range net.fsessionMap {
			if session.status == SS_WORKING {
				session.send(pack(PACKET_HEARTBEAT, nil))
				session.heartbeat()
			}
		}
	}
}

// Dump all frontend sessions
func (net *_netService) dumpHandlerSessions() {
	net.fsmLock.RLock()
	defer net.fsmLock.RUnlock()
	Info(fmt.Sprintf("current frontend session count: %d", len(net.fsessionMap)))
	for _, ses := range net.fsessionMap {
		Info("session: " + ses.String())
	}
}

// Dump all backend sessions
func (net *_netService) dumpRemoteSessions() {
	net.bsmLock.RLock()
	defer net.bsmLock.RUnlock()
	Info(fmt.Sprintf("current backen session count: %d", len(net.bsessionMap)))
	for _, ses := range net.bsessionMap {
		Info("session: " + ses.String())
	}
}
