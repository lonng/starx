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
	"fmt"
	"net"
	"time"

	"github.com/lonnng/starx/cluster"
	"github.com/lonnng/starx/cluster/rpc"
	"github.com/lonnng/starx/log"
	"github.com/lonnng/starx/packet"
	routelib "github.com/lonnng/starx/route"
	"github.com/lonnng/starx/session"
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
	die        chan bool
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
		die:        make(chan bool, 1),
	}
	s := session.New(a)
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

	a.die <- true

	// close all channel
	close(a.die)
	close(a.recvBuffer)
	close(a.sendBuffer)

	transporter.closeSession(a.session)
	a.socket.Close()
}

func (a *agent) ID() int64 {
	return a.id
}

func (a *agent) Send(data []byte) (err error) {
	defer func() {
		if e := recover(); err != nil {
			if er, ok := e.(error); ok {
				err = er
			}
		}
	}()

	if a.status < statusClosed {
		a.sendBuffer <- data
		return nil
	}

	err = ErrSendChannelClosed
	return
}

func (a *agent) Push(session *session.Session, route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("Type=Push, UID=%d, Route=%s, Data=%+v", session.Uid, route, v)

	return transporter.push(session, route, data)
}

// Response message to session
func (a *agent) Response(session *session.Session, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("Type=Response, UID=%d, Data=%+v", session.Uid, v)

	return transporter.response(session, data)
}

func (a *agent) Call(session *session.Session, route string, reply interface{}, args ...interface{}) error {
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
