package starx

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)

type remoteService struct {
	serviceMap   map[string]*service
	routeMap     map[string]uint
	routeCodeMap map[uint]string
}

func newRemote() *remoteService {
	return &remoteService{serviceMap: make(map[string]*service)}
}

func (remote *remoteService) handle(conn net.Conn) {
	defer conn.Close()
	// message buffer
	packetChan := make(chan *unhandledBackendPacket, packetBufferSize)
	// all user logic will be handled in single goroutine
	// synchronized in below routine
	go func() {
		for cpkg := range packetChan {
			remote.processPacket(cpkg.bs, cpkg.packet)
		}
	}()
	// register new session when new connection connected in
	bs := Net.createBackendSession(conn)
	Net.dumpBackendSessions()
	tmp := make([]byte, 512) // save truncated data
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("session closed(" + err.Error() + ")")
			bs.status = SS_CLOSED
			Net.closeSession(bs.userSession)
			Net.dumpBackendSessions()
			break
		}
		tmp = append(tmp, buf[:n]...)
		var pkg *Packet // save decoded packet
		// TODO
		// Refactor this loop
		for len(tmp) > headLength {
			if pkg, tmp = unpack(tmp); pkg != nil {
				packetChan <- &unhandledBackendPacket{bs, pkg}
			} else {
				break
			}
		}
	}
}

func (remote *remoteService) processPacket(bs *backendSession, pkg *Packet) {
	switch pkg.Type {
	case PACKET_HANDSHAKE:
		{
			bs.status = SS_HANDSHAKING
			data, err := json.Marshal(map[string]interface{}{"code": 200, "sys": map[string]float64{"heartbeat": heartbeatInternal.Seconds()}})
			if err != nil {
				Info(err.Error())
			}
			bs.send(pack(PACKET_HANDSHAKE, data))
		}
	case PACKET_HANDSHAKE_ACK:
		{
			bs.status = SS_WORKING
		}
	case PACKET_HEARTBEAT:
		{
			go bs.heartbeat()
		}
	case PACKET_DATA:
		{
			go bs.heartbeat()
			msg := decodeMessage(pkg.Body)
			if msg != nil {
				remote.processMessage(bs.userSession, msg)
			}
		}
	}
}

func (remote *remoteService) processMessage(session *Session, msg *Message) {
	ri, err := decodeRouteInfo(msg.Route)
	if err != nil {
		return
	}
	if msg.Type == MT_REQUEST {
		session.reqId = msg.ID
	} else if msg.Type == MT_NOTIFY {
		session.reqId = 0
	} else {
		Info("invalid message type")
		return
	}
	if s, present := handler.serviceMap[ri.service]; present {
		if m, ok := s.method[ri.method]; ok {
			m.method.Func.Call([]reflect.Value{s.rcvr, reflect.ValueOf(session), reflect.ValueOf(msg.Body)})
		} else {
			Info("method: " + ri.method + " not found")
		}
	} else {
		Info("service: " + ri.service + " not found")
	}
}

// Register publishes in the service the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the first argument is *starx.Session
//	- the second argument is []byte
func (remote *remoteService) register(rcvr HandlerComponent) {
	rcvr.Setup()
	remote._register(rcvr)
}

func (remote *remoteService) _register(rcvr HandlerComponent) {
	if remote.serviceMap == nil {
		remote.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		s := "remote.Register: no service name for type " + s.typ.String()
		Info(s)
		return
	}
	if !isExported(sname) {
		s := "remote.Register: type " + sname + " is not exported"
		Info(s)
		return
	}
	if _, present := remote.serviceMap[sname]; present {
		Info("remote: service already defined: " + sname)
		return
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type"
		}
		Info(str)
	}
	remote.serviceMap[s.name] = s
	remote.dumpServiceMap()
}

func (remote *remoteService) dumpServiceMap() {
	for sname, s := range remote.serviceMap {
		for mname, _ := range s.method {
			Info(fmt.Sprintf("registered service: %s.%s", sname, mname))
		}
	}
}
