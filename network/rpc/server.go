package rpc

import (
	"errors"
	"reflect"
	stddebug "runtime/debug"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/utils"
	"os"
)

type ResponseKind byte

const (
	_               ResponseKind = iota
	HandlerResponse              // handler session response
	HandlerPush                  // handler session push
	RemoteResponse               // remote request normal response, represent whether rpc call successfully
	RemotePush                   // using remote server push message to current server
)

type RpcKind byte

const (
	_       RpcKind = iota
	SysRpc          // sys namespace rpc
	UserRpc         // user namespace rpc
)

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

// Request is a header written before every RPC call.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Request struct {
	ServiceMethod string  // format: "Service.Method"
	Seq           uint64  // sequence number chosen by client
	Sid           uint64  // frontend session id
	Args          []byte  // for args
	Kind          RpcKind // namespace
}

// Response is a header written before every RPC return.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Response struct {
	Kind          ResponseKind // rpc response type
	ServiceMethod string       // echoes that of the Request
	Seq           uint64       // echoes that of the request
	Sid           uint64       // frontend session id
	Data          []byte       // save response value
	Error         string       // error, if any.
	Route         string       // exists when ResponseType equal RPC_HANDLER_PUSH
}

// Server represents an RPC Server.
type Server struct {
	Kind       RpcKind             // rpc kind, either SysRpc or UserRpc
	mu         sync.RWMutex        // protects the serviceMap
	serviceMap map[string]*service // all service
}

// NewServer returns a new Server.
func NewServer(kind RpcKind) *Server {
	return &Server{Kind: kind, serviceMap: make(map[string]*service)}
}

// SysRpcServer is the system namespace rpc instance of *Server.
var SysRpcServer = NewServer(SysRpc)

// UserRpcServer is the user namespace rpc instance of *Server
var UserRpcServer = NewServer(UserRpc)

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

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the second argument is a pointer
//	- one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		return errors.New("remote.Register: no service name for type " + s.typ.String())
	}
	if !isExported(sname) && !useName {
		return errors.New("remote.Register: type " + sname + " is not exported")
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("remote: service already defined: " + sname)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(server.Kind, s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(server.Kind, reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}
	server.serviceMap[s.name] = s
	return nil
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(kind RpcKind, typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	switch kind {
	case SysRpc:
		for m := 0; m < typ.NumMethod(); m++ {
			method := typ.Method(m)
			mtype := method.Type
			mname := method.Name
			if utils.IsHandlerMethod(method) {
				methods[mname] = &methodType{method: method, ArgType: mtype.In(1), ReplyType: mtype.In(2)}
			}
		}
	case UserRpc:
		for m := 0; m < typ.NumMethod(); m++ {
			method := typ.Method(m)
			mname := method.Name
			if utils.IsRemoteMethod(method) {
				methods[mname] = &methodType{method: method}
			}
		}
	}
	return methods
}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (server *Server) Call(serviceMethod string, args []reflect.Value) (r []reflect.Value, err error) {
	defer func() {
		if recov := recover(); recov != nil {
			log.Fatal("RpcCall Error: %+v", recov)
			os.Stderr.Write(stddebug.Stack())
			if s, ok := recov.(string); ok {
				err = errors.New(s)
			} else {
				err = errors.New("RpcCall internal error")
			}
		}
	}()
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return nil, errors.New("wrong route string: " + serviceMethod)
	}

	s, m := parts[0], parts[1]

	service, ok := server.serviceMap[s]
	if !ok || service == nil {
		return nil, errors.New("remote: servive " + s + " does not exists")
	}

	method, ok := service.method[m]
	if !ok || method == nil {
		return nil, errors.New("remote: service " + s + "does not contain method: " + m)
	}
	args = append([]reflect.Value{service.rcvr}, args...)
	rets := method.method.Func.Call(args)
	return rets, nil
}

var rpcResponseKindNames = []string{
	HandlerResponse: "HandlerResponse",
	HandlerPush:     "HandlerPush",
	RemoteResponse:  "RemoteResponse",
}

func (k ResponseKind) String() string {
	if int(k) < len(rpcResponseKindNames) {
		return rpcResponseKindNames[k]
	}
	return strconv.Itoa(int(k))
}

var rpcKindNames = []string{
	SysRpc:  "SysRpc",  // system rpc
	UserRpc: "UserRpc", // user rpc
}

func (k RpcKind) String() string {
	if int(k) < len(rpcKindNames) {
		return rpcKindNames[k]
	}
	return strconv.Itoa(int(k))
}
