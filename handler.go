package starx

import (
	"encoding/json"
	"errors"
	"github.com/chrislonng/starx/rpc"
	"github.com/chrislonng/starx/utils"
	"net"
	"reflect"
	"sync"
)

// Unhandled message buffer size
// Every connection has an individual message channel buffer
const (
	packetBufferSize = 256
)

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	Arg1Type   reflect.Type
	Arg2Type   reflect.Type
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
			Info("session closed(" + err.Error() + ")")
			agent.close()
			endChan <- true
			break
		}
		tmp = append(tmp, buf[:n]...)
		var pkg *packet // save decoded packet
		for len(tmp) >= headLength {
			if pkg, tmp = unpack(tmp); pkg != nil {
				packetChan <- &unhandledPacket{agent, pkg}
			} else {
				break
			}
		}
	}
}

func (handler *handlerService) processPacket(fs *agent, pkg *packet) {
	switch pkg.kind {
	case _PACKET_HANDSHAKE:
		fs.status = _STATUS_HANDSHAKING
		data, err := json.Marshal(map[string]interface{}{"code": 200, "sys": map[string]float64{"heartbeat": heartbeatInternal.Seconds()}})
		if err != nil {
			Info(err.Error())
		}
		fs.send(pack(_PACKET_HANDSHAKE, data))
	case _PACKET_HANDSHAKE_ACK:
		fs.status = _STATUS_WORKING
	case _PACKET_HEARTBEAT:
		go fs.heartbeat()
	case _PACKET_DATA:
		go fs.heartbeat()
		if msg := decodeMessage(pkg.body); msg != nil {
			handler.processMessage(fs.session, msg)
		}
	default:
		Info("invalid packet type")
		fs.close()
	}
}

func (handler *handlerService) processMessage(session *Session, msg *message) {
	defer func() {
		if err := recover(); err != nil {
			Error("processMessage Error: %+v", err)
		}
	}()
	Info("Route: %s, Length: %d", msg.route, len(msg.body))
	ri, err := decodeRouteInfo(msg.route)
	if err != nil {
		Error(err.Error())
		return
	}
	// if serverType equal nil, message handle in local server
	if ri.serverType == "" || ri.serverType == App.Config.Type {
		handler.localProcess(session, ri, msg)
	} else {
		handler.remoteProcess(session, ri, msg)
	}
}

// current message handle in local server
func (handler *handlerService) localProcess(session *Session, ri *routeInfo, msg *message) {
	if msg.kind == _MT_REQUEST {
		session.reqId = msg.id
	} else if msg.kind == _MT_NOTIFY {
		session.reqId = 0
	} else {
		Error("invalid message type")
		return
	}
	if s, present := handler.serviceMap[ri.service]; present {
		if m, ok := s.method[ri.method]; ok {
			ret := m.method.Func.Call([]reflect.Value{s.rcvr, reflect.ValueOf(session), reflect.ValueOf(msg.body)})
			if len(ret) > 0 {
				err := ret[0].Interface()
				if err != nil {
					Error(err.(error).Error())
				}
			}
		} else {
			Info("handler: " + ri.service + " does not contain method: " + ri.method)
		}
	} else {
		Info("handler: service: " + ri.service + " not found")
	}
}

// current message handle in remote server
func (handler *handlerService) remoteProcess(session *Session, ri *routeInfo, msg *message) {
	if msg.kind == _MT_REQUEST {
		session.reqId = msg.id
		remote.request(rpc.SysRpc, ri, session, msg.body)
	} else if msg.kind == _MT_NOTIFY {
		session.reqId = 0
		remote.request(rpc.SysRpc, ri, session, msg.body)
	} else {
		Info("invalid message type")
		return
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
			methods[mname] = &methodType{method: method, Arg1Type: mtype.In(1), Arg2Type: mtype.In(2)}
		}
	}
	return methods
}

func (handler *handlerService) dumpServiceMap() {
	for sname, s := range handler.serviceMap {
		for mname, _ := range s.method {
			Info("registered service: %s.%s", sname, mname)
		}
	}
}
