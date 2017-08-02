// Copyright (c) starx Author. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package starx

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lonnng/starx/cluster"
	"github.com/lonnng/starx/log"
	"github.com/lonnng/starx/message"
	"github.com/lonnng/starx/packet"
	"github.com/lonnng/starx/session"
)

const sessionClosedRoute = "__Session.Closed"

var (
	ErrSessionOnNotify = errors.New("current session working on notify mode")
	ErrSessionNotFound = errors.New("session not found")
)

var (
	heartbeatPacket, _ = packet.Pack(&packet.Packet{Type: packet.Heartbeat})

	// transporter represents a manager, which manages low-level transport
	// layer object, that abstract as `agent` in frontend server or `acceptor`
	// in the backend server
	transporter = newTransporter()
)

type transportService struct {
	sync.RWMutex
	agents      map[int64]*agent    // agents map
	acceptorUid int64               // acceptor unique id
	acceptors   map[int64]*acceptor // acceptor map

	sessionCloseCbLock sync.RWMutex             // protect sessionCloseCb
	sessionCloseCb     []func(*session.Session) // callback on session closed
}

// Create new t service
func newTransporter() *transportService {
	return &transportService{
		agents:      make(map[int64]*agent),
		acceptorUid: 0,
		acceptors:   make(map[int64]*acceptor),
	}
}

// Create agent via transportService
func (t *transportService) createAgent(conn net.Conn) *agent {
	a := newAgent(conn)
	// add to maps
	t.Lock()
	defer t.Unlock()

	t.agents[a.id] = a
	return a
}

// get agent by session id
func (t *transportService) agent(id int64) (*agent, error) {
	t.RLock()
	defer t.RUnlock()

	a, ok := t.agents[id]
	if !ok {
		return nil, errors.New("agent id: " + string(id) + " not exists!")
	}

	return a, nil
}

// Create acceptor via transportService
func (t *transportService) createAcceptor(conn net.Conn) *acceptor {
	id := atomic.AddInt64(&t.acceptorUid, 1)
	a := newAcceptor(id, conn)

	// add to maps
	t.Lock()
	defer t.Unlock()

	t.acceptors[id] = a
	return a
}

func (t *transportService) acceptor(id int64) (*acceptor, error) {
	t.RLock()
	defer t.RUnlock()

	rs, ok := t.acceptors[id]
	if !ok || rs == nil {
		return nil, errors.New("acceptor id: " + string(id) + " not exists!")
	}

	return rs, nil
}

// Send packet data, call by package internal, the second argument was packaged packet
// if current server is frontend server, send to client by agent, else send to frontend
// server by acceptor
func (t *transportService) send(session *session.Session, data []byte) {
	session.Entity.Send(data)
}

// Push message to client
// call by all package, the last argument was packaged message
func (t *transportService) push(session *session.Session, route string, data []byte) error {
	m, err := message.Encode(&message.Message{
		Type:  message.MessageType(message.Push),
		Route: route,
		Data:  data,
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

	t.send(session, ep)
	return nil
}

// Response message to client
// call by all package, the last argument was packaged message
func (t *transportService) response(session *session.Session, data []byte) error {
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

	t.send(session, ep)
	return nil
}

// TODO: implement backend server broadcast
// broadcast message to all sessions
// Message level method
// call by all package, the last argument was packaged message
func (t *transportService) broadcast(route string, data []byte) {
	if !app.config.IsFrontend {
		return
	}

	for _, s := range t.agents {
		t.push(s.session, route, data)
	}
}

// Multicast message to special agent ids
func (t *transportService) multicast(aids []int64, route string, data []byte) {
	t.RLock()
	defer t.RUnlock()

	for _, aid := range aids {
		if agent, ok := t.agents[aid]; ok && agent != nil {
			t.push(agent.session, route, data)
		}
	}
}

func (t *transportService) Session(sid int64) (*session.Session, error) {
	t.RLock()
	defer t.RUnlock()

	a, ok := t.agents[sid]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return a.session, nil
}

// Close session
func (t *transportService) closeSession(session *session.Session) {
	t.sessionCloseCbLock.RLock()
	for _, cb := range t.sessionCloseCb {
		if cb != nil {
			cb(session)
		}
	}
	t.sessionCloseCbLock.RUnlock()

	t.Lock()
	defer t.Unlock()

	if app.config.IsFrontend {
		if agent, ok := t.agents[session.Entity.ID()]; ok && (agent != nil) {
			delete(t.agents, session.Entity.ID())
		}
		// notify all backend server, current session has been closed.
		cluster.SessionClosed(session)
	} else {
		if acceptor, ok := t.acceptors[session.Entity.ID()]; ok && (acceptor != nil) {
			delete(acceptor.sessionMap, session.ID)
			if fid, ok := acceptor.b2fMap[session.ID]; ok {
				delete(acceptor.b2fMap, session.ID)
				delete(acceptor.f2bMap, fid)
			}
		}
	}
}

func (t *transportService) removeAcceptor(a *acceptor) {
	t.Lock()
	defer t.Unlock()

	delete(t.acceptors, a.id)
}

// Send heartbeat packet
func (t *transportService) heartbeat() {
	if !app.config.IsFrontend || t.agents == nil {
		return
	}
	dt := time.Now().Add(-2 * env.heartbeatInternal)
	dtu := dt.Unix()

	for _, agent := range t.agents {
		if agent.status != statusWorking {
			continue
		}

		if agent.lastTime < dtu {
			log.Debugf("Session heartbeat timeout, LastTime=%d, Deadline=%d", agent.lastTime, dtu)
			agent.Close()
			continue
		}

		if err := agent.Send(heartbeatPacket); err != nil {
			log.Error(err)
			agent.Close()
			continue
		}
	}
}

// Dump all agents
func (t *transportService) dumpAgents() {
	t.RLock()
	defer t.RUnlock()

	log.Infof("current agent count: %d", len(t.agents))
	for _, ses := range t.agents {
		log.Infof("session: " + ses.String())
	}
}

// Dump all acceptor
func (t *transportService) dumpAcceptor() {
	t.RLock()
	defer t.RUnlock()

	log.Infof("current acceptor count: %d", len(t.acceptors))
	for _, ses := range t.acceptors {
		log.Infof("session: " + ses.String())
	}
}

func (t *transportService) sessionClosedCallback(cb func(*session.Session)) {
	t.sessionCloseCbLock.Lock()
	defer t.sessionCloseCbLock.Unlock()

	t.sessionCloseCb = append(t.sessionCloseCb, cb)
}

// Callback when session closed
// Waring: session has closed,
func OnSessionClosed(cb func(*session.Session)) {
	transporter.sessionClosedCallback(cb)
}
