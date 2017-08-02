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
	"fmt"
	"net"
	"time"

	"github.com/lonnng/starx/cluster"
	"github.com/lonnng/starx/cluster/rpc"
	"github.com/lonnng/starx/log"
	routelib "github.com/lonnng/starx/route"
	"github.com/lonnng/starx/session"
)

// Acceptor corresponding a front server, used for store raw socket
// information.
// only used in package internal, can not accessible by other package
type acceptor struct {
	id         int64
	socket     net.Conn
	status     networkStatus
	sessionMap map[int64]*session.Session // backend sessions
	f2bMap     map[int64]int64            // frontend session id -> backend session id map
	b2fMap     map[int64]int64            // backend session id -> frontend session id map
	lastTime   int64                      // last heartbeat unix time stamp
}

// Create new backend session instance
func newAcceptor(id int64, conn net.Conn) *acceptor {
	return &acceptor{
		id:         id,
		socket:     conn,
		status:     statusStart,
		sessionMap: make(map[int64]*session.Session),
		f2bMap:     make(map[int64]int64),
		b2fMap:     make(map[int64]int64),
		lastTime:   time.Now().Unix(),
	}
}

// String implement Stringer interface
func (a *acceptor) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		a.id,
		a.socket.RemoteAddr().String(),
		a.lastTime)
}

func (a *acceptor) heartbeat() {
	a.lastTime = time.Now().Unix()
}

func (a *acceptor) Session(sid int64) *session.Session {
	if bsid, ok := a.f2bMap[sid]; ok && bsid > 0 {
		return a.sessionMap[bsid]
	}
	s := session.New(a)
	a.sessionMap[s.ID] = s
	a.f2bMap[sid] = s.ID
	a.b2fMap[s.ID] = sid
	return s
}

func (a *acceptor) Close() {
	a.status = statusClosed
	for _, s := range a.sessionMap {
		transporter.closeSession(s)
	}
	transporter.removeAcceptor(a)
	a.socket.Close()
}

func (a *acceptor) ID() int64 {
	return a.id
}

func (a *acceptor) Send(data []byte) error {
	_, err := a.socket.Write(data)
	return err
}

func (a *acceptor) Push(session *session.Session, route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("UID=%d, Type=Push, Route=%s, Data=%+v", session.Uid, route, v)

	rs, err := transporter.acceptor(session.Entity.ID())
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	sid, ok := rs.b2fMap[session.ID]
	if !ok {
		log.Errorf("sid not exists")
		return ErrSidNotExists
	}

	resp := &rpc.Response{
		Route: route,
		Kind:  rpc.HandlerPush,
		Data:  data,
		Sid:   sid,
	}
	return rpc.WriteResponse(a.socket, resp)
}

// Response message to session
func (a *acceptor) Response(session *session.Session, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("UID=%d, Type=Response, Data=%+v", session.Uid, v)

	rs, err := transporter.acceptor(session.Entity.ID())
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	sid, ok := rs.b2fMap[session.ID]
	if !ok {
		log.Errorf("sid not exists")
		return ErrSidNotExists
	}
	resp := &rpc.Response{
		Kind: rpc.HandlerResponse,
		Data: data,
		Sid:  sid,
	}
	return rpc.WriteResponse(a.socket, resp)
}

func (a *acceptor) Call(session *session.Session, route string, reply interface{}, args ...interface{}) error {
	r, err := routelib.Decode(route)
	if err != nil {
		return err
	}

	if app.config.Type == r.ServerType {
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
