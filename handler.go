package starx

import (
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"runtime"
	"sync"

	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/message"
	"github.com/chrislonng/starx/network"
	"github.com/chrislonng/starx/network/rpc"
	"github.com/chrislonng/starx/packet"
	"github.com/chrislonng/starx/utils"
)

// Unhandled message buffer size
// Every connection has an individual message channel buffer
const (
	packetBufferSize = 256
)

type unhandledPacket struct {
	fs     *agent
	packet *packet.Packet
}

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	dataType   reflect.Type
	numCalls   uint
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type handlerService struct {
	serviceMap   map[string]*service
	routeMap     map[string]uint
	routeCodeMap map[uint]string
}

func newHandler() *handlerService {
	return &handlerService{
		serviceMap: make(map[string]*service)}
}

// Handle network connection
// Read data from Socket file descriptor and decode it, handle message in
// individual logic routine
func (handler *handlerService) handle(conn net.Conn) {
	defer conn.Close()
	// message buffer
	packetChan := make(chan *unhandledPacket, packetBufferSize)
	endChan := make(chan bool, 1)
	// all user logic will be handled in single goroutine
	// synchronized in below routine
	go func() {
		for {
			select {
			case cpkg := <-packetChan:
				handler.processPacket(cpkg.fs, cpkg.packet)
			case <-endChan:
				close(packetChan)
				return
			}
		}

	}()
	// register new session when new connection connected in
	agent := defaultNetService.createAgent(conn)
	defaultNetService.dumpAgents()
	tmp := make([]byte, 0) // save truncated data
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Info("session closed(" + err.Error() + ")")
			agent.close()
			endChan <- true
			break
		}
		tmp = append(tmp, buf[:n]...)
		var pkg *packet.Packet // save decoded packet
		for len(tmp) >= packet.HeadLength {
			pkg, tmp, err = packet.Unpack(tmp)
			if err != nil {
				agent.close()
				break
			}
			packetChan <- &unhandledPacket{agent, pkg}
		}
	}
}

func (handler *handlerService) processPacket(a *agent, p *packet.Packet) {
	switch p.Type {
	case packet.Handshake:
		a.status = statusHandshake
		data, err := json.Marshal(map[string]interface{}{"code": 200, "sys": map[string]float64{"heartbeat": heartbeatInternal.Seconds()}})
		if err != nil {
			log.Info(err.Error())
		}
		rp := &packet.Packet{
			Type:   packet.Handshake,
			Length: len(data),
			Data:   data,
		}
		resp, err := rp.Pack()
		if err != nil {
			log.Error(err.Error())
			a.close()
		}
		a.send(resp)
	case packet.HandshakeAck:
		a.status = statusWorking
	case packet.Data:
		m, err := message.Decode(p.Data)
		if err != nil {
			log.Error(err.Error())
			return
		}
		handler.processMessage(a.session, m)
		fallthrough
	case packet.Heartbeat:
		go a.heartbeat()
	default:
		log.Info("invalid packet type")
		a.close()
	}
}

func (handler *handlerService) processMessage(session *Session, m *message.Message) {
	defer func() {
		if err := recover(); err != nil {
			runtime.Caller(2)
			log.Fatal("processMessage Error: %+v", err)
		}
	}()
	log.Info("Route: %s, Length: %d", m.Route, len(m.Data))
	r, err := network.DecodeRoute(m.Route)
	if err != nil {
		log.Error(err.Error())
		return
	}
	// if serverType equal nil, message handle in local server
	if r.ServerType == "" || r.ServerType == App.Config.Type {
		handler.localProcess(session, r, m)
	} else {
		handler.remoteProcess(session, r, m)
	}
}

// current message handle in local server
func (handler *handlerService) localProcess(session *Session, route *network.Route, msg *message.Message) {
	switch msg.Type {
	case message.Request:
		session.reqId = msg.ID
	case message.Notify:
		session.reqId = 0
	default:
		log.Error("invalid message type")
		return
	}

	s, ok := handler.serviceMap[route.Service]
	if !ok || s == nil {
		log.Info("handler: service: " + route.Service + " not found")
	}

	m, ok := s.method[route.Method]
	if !ok || m == nil {
		log.Info("handler: " + route.Service + " does not contain method: " + route.Method)
	}

	data := reflect.New(m.dataType.Elem()).Interface()
	err := serializer.Deserialize(msg.Data, data)
	if err != nil {
		log.Error("deserialize error: %s", err.Error())
		return
	}

	ret := m.method.Func.Call([]reflect.Value{s.rcvr, reflect.ValueOf(session), reflect.ValueOf(data)})
	if len(ret) > 0 {
		err := ret[0].Interface()
		if err != nil {
			log.Error(err.(error).Error())
		}
	}
}

// current message handle in remote server
func (handler *handlerService) remoteProcess(session *Session, route *network.Route, msg *message.Message) {
	switch msg.Type {
	case message.Request:
		session.reqId = msg.ID
		remote.request(rpc.Sys, route, session, msg.Data)
	case message.Notify:
		session.reqId = 0
		remote.request(rpc.Sys, route, session, msg.Data)
	default:
		log.Error("invalid message type")
	}
}

// Register publishes in the service the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the first argument is *starx.Session
//	- the second argument is []byte
func (handler *handlerService) register(rcvr Component) error {
	if handler.serviceMap == nil {
		handler.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		return errors.New("handler.Register: no service name for type " + s.typ.String())
	}
	if !utils.IsExported(sname) {
		return errors.New("handler.Register: type " + sname + " is not exported")

	}
	if _, present := handler.serviceMap[sname]; present {
		return errors.New("handler: service already defined: " + sname)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "handler.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "handler.Register: type " + sname + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}
	handler.serviceMap[s.name] = s
	return nil
}

// suitableMethods returns suitable methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if utils.IsHandlerMethod(method) {
			methods[mname] = &methodType{method: method, dataType: mtype.In(2)}
		}
	}
	return methods
}

func (handler *handlerService) dumpServiceMap() {
	for sname, s := range handler.serviceMap {
		for mname, _ := range s.method {
			log.Info("registered service: %s.%s", sname, mname)
		}
	}
}
