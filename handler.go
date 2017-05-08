package starx

import (
	"encoding/json"
	"errors"
	"net"
	"reflect"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/message"
	"github.com/chrislonng/starx/packet"
	"github.com/chrislonng/starx/route"
	"github.com/chrislonng/starx/session"
)

// Unhandled message buffer size
// Every connection has an individual message channel buffer
const (
	packetBufferSize = 256
)

var handler = newHandlerService()

type handlerService struct {
	serviceMap map[string]*component.Service
}

func newHandlerService() *handlerService {
	return &handlerService{
		serviceMap: make(map[string]*component.Service),
	}
}

func (hs *handlerService) register(rcvr component.Component) error {
	if hs.serviceMap == nil {
		hs.serviceMap = make(map[string]*component.Service)
	}

	s := &component.Service{
		Type: reflect.TypeOf(rcvr),
		Rcvr: reflect.ValueOf(rcvr),
	}
	s.Name = reflect.Indirect(s.Rcvr).Type().Name()

	if _, ok := hs.serviceMap[s.Name]; ok {
		return errors.New("handler: service already defined: " + s.Name)
	}

	if err := s.ScanHandler(); err != nil {
		return err
	}

	hs.serviceMap[s.Name] = s

	return nil
}

// Handle network connection
// Read data from Socket file descriptor and decode it, handle message in
// individual logic goroutine
func (hs *handlerService) handle(conn net.Conn) {
	defer conn.Close()

	// register new session when new connection connected in
	agent := defaultNetService.createAgent(conn)
	log.Debugf("New session established: %s", agent.String())

	// all user logic will be handled in single goroutine
	// synchronized in below routine
	go func() {
	LOOP:
		for {
			select {
			case p, ok := <-agent.recvBuffer:
				if ok && p != nil {
					hs.processPacket(agent, p)
				}
			case m, ok := <-agent.sendBuffer:
				if ok && m != nil {
					_, err := agent.socket.Write(m)
					if err != nil {
						log.Error(err)
						agent.Close()
					}
				}
			case <-agent.ending:
				break LOOP
			}
		}
	}()

	tmp := make([]byte, 0) // save truncated data
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Errorf("Read message error: %s, session will be closed immediately", err.Error())
			agent.Close()
			break // break read packet loop
		}
		tmp = append(tmp, buf[:n]...)

		// save decoded packet
		var p *packet.Packet
		for len(tmp) >= packet.HeadLength {
			p, tmp, err = packet.Unpack(tmp)
			if err != nil {
				agent.Close()
				break
			}

			if p == nil {
				break
			}
			agent.recvBuffer <- p
		}
	}
}

func (hs *handlerService) processPacket(a *agent, p *packet.Packet) {
	switch p.Type {
	case packet.Handshake:
		a.status = statusHandshake
		data, err := json.Marshal(map[string]interface{}{
			"code": 200,
			"sys":  map[string]float64{"heartbeat": heartbeatInternal.Seconds()},
		})
		if err != nil {
			log.Infof(err.Error())
		}

		rp := &packet.Packet{
			Type:   packet.Handshake,
			Length: len(data),
			Data:   data,
		}

		resp, err := rp.Pack()
		if err != nil {
			log.Errorf(err.Error())
			a.Close()
		}

		if err := a.Send(resp); err != nil {
			log.Errorf(err.Error())
			a.Close()
		}
		log.Debugf("Session handshake Id=%d, Remote=%s", a.id, a.socket.RemoteAddr())
	case packet.HandshakeAck:
		a.status = statusWorking
		log.Debugf("Receive handshake ACK Id=%d, Remote=%s", a.id, a.socket.RemoteAddr())
	case packet.Data:
		m, err := message.Decode(p.Data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		hs.processMessage(a.session, m)
		fallthrough
	case packet.Heartbeat:
		go a.heartbeat()
	default:
		log.Infof("invalid packet type")
		a.Close()
	}
}

func (hs *handlerService) processMessage(session *session.Session, msg *message.Message) {
	defer func() {
		if err := recover(); err != nil {
			log.Tracef("processMessage Error: %+v", err)
		}
	}()

	switch msg.Type {
	case message.Request:
		session.LastID = msg.ID
	case message.Notify:
		session.LastID = 0
	default:
		log.Errorf("invalid message type")
		return
	}

	r, err := route.Decode(msg.Route)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	// current server as default server type
	if r.ServerType == "" {
		r.ServerType = App.Config.Type
	}

	// message dispatch
	if r.ServerType == App.Config.Type {
		hs.localProcess(session, r, msg)
	} else {
		hs.remoteProcess(session, r, msg)
	}
}

// current message handle in local server
func (hs *handlerService) localProcess(session *session.Session, route *route.Route, msg *message.Message) {
	s, ok := hs.serviceMap[route.Service]
	if !ok || s == nil {
		log.Infof("handler: service: " + route.Service + " not found")
		return
	}

	m, ok := s.HandlerMethods[route.Method]
	if !ok || m == nil {
		log.Infof("handler: " + route.Service + " does not contain method: " + route.Method)
		return
	}

	var data interface{}
	if m.Raw {
		data = msg.Data
	} else {
		data = reflect.New(m.Type.Elem()).Interface()
		err := serializer.Deserialize(msg.Data, data)
		if err != nil {
			log.Errorf("deserialize error: %s", err.Error())
			return
		}
	}

	log.Debugf("Uid=%d, Message={%s}, Data=%+v", session.Uid, msg.String(), data)

	ret := m.Method.Func.Call([]reflect.Value{s.Rcvr, reflect.ValueOf(session), reflect.ValueOf(data)})
	if len(ret) > 0 {
		err := ret[0].Interface()
		if err != nil {
			log.Errorf(err.(error).Error())
		}
	}
}

// current message handle in remote server
func (hs *handlerService) remoteProcess(session *session.Session, route *route.Route, msg *message.Message) {
	if _, err := cluster.Call(rpc.Sys, route, session, msg.Data); err != nil {
		log.Errorf(err.Error())
	}
}

func (hs *handlerService) dumpServiceMap() {
	for sname, s := range hs.serviceMap {
		for mname := range s.HandlerMethods {
			log.Infof("registered service: %s.%s", sname, mname)
		}
	}
}
