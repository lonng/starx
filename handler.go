package starx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Unhandled message buffer size
// Every connection has an individual message channel buffer
const (
	messageBufferSize = 256
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

type HandlerService struct {
	serviceMap   map[string]*service
	routeMap     map[string]uint
	routeCodeMap map[uint]string
}

func NewHandler() *HandlerService {
	return &HandlerService{
		serviceMap: make(map[string]*service)}
}

// Handle network connection
// Read data from Socket file descriptor and decode it, handle message in
// individual logic routine
func (handler *HandlerService) Handle(conn net.Conn) {
	defer conn.Close()
	// message buffer
	messageChan := make(chan *unhandledMessage, messageBufferSize)
	// all user logic will be handled in single goroutine
	// synchronized in below routine
	go func() {
		for {
			cmsg := <-messageChan
			handler.processMessage(cmsg.session, cmsg.message)
		}
	}()
	// register new session when new connection connected in
	session := sessionService.RegisterSession(conn)
	sessionService.dumpSessions() // TODO delete this line
	tmp := make([]byte, 0)        // save truncated data
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("session closed(" + err.Error() + ")")
			session.status = SS_CLOSED
			sessionService.RemoveSession(session)
			sessionService.dumpSessions()
			break
		}

		var pkg *Packet // save decoded packet
		pkg, tmp = unpack(append(tmp, buf[:n]...))
		if pkg != nil {
			switch pkg.Type {
			case PACKET_HANDSHAKE:
				{
					session.status = SS_HANDSHAKING
					Info(pkg.String())
					data, err := json.Marshal(map[string]interface{}{"code": 200, "sys": map[string]float64{"heartbeat": heartbeatInternal.Seconds()}})
					if err != nil {
						Info(err.Error())
					}
					conn.Write(pack(PACKET_HANDSHAKE, data))
				}
			case PACKET_HANDSHAKE_ACK:
				{
					session.status = SS_WORKING
				}
			case PACKET_HEARTBEAT:
				{
					session.heartbeat()
				}
			case PACKET_DATA:
				{
					session.heartbeat()
					msg := decodeMessage(pkg.Body)
					if msg != nil {
						messageChan <- &unhandledMessage{session, msg}
					}
				}
			}
		}
	}
}

func encodeMessage(m *Message) []byte {
	temp := make([]byte, 0)
	flag := byte(m.Type) << 1
	if m.isCompress {
		flag |= 0x01
	}
	temp = append(temp, flag)
	// response message
	if m.Type == MT_RESPONSE {
		n := m.ID
		for {
			b := byte(n % 128)
			n >>= 7
			if n != 0 {
				temp = append(temp, b+128)
			} else {
				temp = append(temp, b)
				break
			}
		}
		fmt.Println("%+v", temp)
	} else if m.Type == MT_PUSH {
		if m.isCompress {
			temp = append(temp, byte((m.RouteCode>>8)&0xFF))
			temp = append(temp, byte(m.RouteCode&0xFF))
		} else {
			temp = append(temp, byte(len(m.Route)))
			temp = append(temp, []byte(m.Route)...)
		}
	} else {
		Error("wrong message type")
	}
	temp = append(temp, m.Body...)
	return temp
}

func decodeMessage(data []byte) *Message {
	// filter invalid message
	if len(data) <= 3 {
		Info("invalid message")
		return nil
	}
	msg := NewMessage()
	flag := data[0]
	// set offset to 1, because 1st byte will always be flag
	offset := 1
	msg.Type = MessageType((flag >> 1) & MSG_TYPE_MASK)
	if msg.Type == MT_REQUEST {
		id := uint(0)
		// little end byte order
		// WARNING: must can be stored in 64 bits integer
		for i := offset; i < len(data); i++ {
			b := data[i]
			id += (uint(b&0x7F) << uint(7*(i-offset)))
			if b < 128 {
				offset = i + 1
				break
			}
		}
		msg.ID = id
	}
	if flag&MSG_ROUTE_COMPRESS_MASK == 1 {
		msg.isCompress = true
		msg.RouteCode = uint(bytesToInt(data[offset:(offset + 2)]))
		offset += 2
	} else {
		msg.isCompress = false
		rl := data[offset]
		offset += 1
		msg.Route = string(data[offset:(offset + int(rl))])
		offset += int(rl)
	}
	msg.Body = data[offset:]
	return msg
}

func (handler *HandlerService) processMessage(session *Session, msg *Message) {
	ri, err := decodeRouteInfo(msg.Route)
	if err != nil {
		return
	}
	if ri.server == App.CurSvrConfig.Type {
		handler.localProcess(session, ri, msg)
	} else {
		handler.remoteProcess(session, ri, msg)
	}
}

func decodeRouteInfo(route string) (*routeInfo, error) {
	parts := strings.Split(route, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid route")
	}
	return newRouteInfo(parts[0], parts[1], parts[2]), nil
}

// TODO: implement request protocol
func (handler *HandlerService) localProcess(session *Session, ri *routeInfo, msg *Message) {
	if msg.Type == MT_REQUEST {
		// TODO
	} else if msg.Type == MT_NOTIFY {
		if s, present := handler.serviceMap[ri.service]; present {
			if m, ok := s.method[ri.method]; ok {
				m.method.Func.Call([]reflect.Value{s.rcvr, reflect.ValueOf(session), reflect.ValueOf(msg.Body)})
			} else {
				Info("method: " + ri.method + " not found")
			}
		} else {
			Info("service: " + ri.service + " not found")
		}
	} else {
		Info("unrecognize message type")
	}
}

// TODO: implemention
func (handler *HandlerService) remoteProcess(session *Session, ri *routeInfo, msg *Message) {
}

// Register publishes in the service the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the first argument is *starx.Session
//	- the second argument is []byte
func (handler *HandlerService) Register(rcvr HandlerComponent) {
	rcvr.Setup()
	handler.register(rcvr)
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

func (handler *HandlerService) register(rcvr HandlerComponent) {
	if handler.serviceMap == nil {
		handler.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		s := "handler.Register: no service name for type " + s.typ.String()
		Info(s)
		return
	}
	if !isExported(sname) {
		s := "handler.Register: type " + sname + " is not exported"
		Info(s)
		return
	}
	if _, present := handler.serviceMap[sname]; present {
		Info("handler: service already defined: " + sname)
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
			str = "handler.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "handler.Register: type " + sname + " has no exported methods of suitable type"
		}
		Info(str)
	}
	handler.serviceMap[s.name] = s
	handler.dumpServiceMap()
}

// suitableMethods returns suitable methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *Session, []byte.
		if mtype.NumIn() != 3 {
			continue
		}
		// First arg need not be *Session.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				fmt.Println(mname, "argument type not exported:", argType)
			}
			continue
		}
		if argType.Kind() != reflect.Ptr || argType.Elem().Name() != "Session" {
			if reportErr {
				fmt.Println("method", mname, " first argument must be a Session pointer:", argType)
			}
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Slice {
			if reportErr {
				fmt.Println("method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				fmt.Println("method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		methods[mname] = &methodType{method: method, Arg1Type: argType, Arg2Type: replyType}
	}
	return methods
}

func (handler *HandlerService) dumpServiceMap() {
	for sname, s := range handler.serviceMap {
		for mname, _ := range s.method {
			Info(fmt.Sprintf("registered service: %s.%s", sname, mname))
		}
	}
}
