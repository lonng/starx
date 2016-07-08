package rpc

import (
	"strconv"
	"sync"
	"unicode"
	"unicode/utf8"
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
	_    RpcKind = iota
	Sys          // sys namespace rpc
	User         // user namespace rpc
)

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
	Kind RpcKind      // rpc kind, either SysRpc or UserRpc
	mu   sync.RWMutex // protects the serviceMap
}

// NewServer returns a new Server.
func NewServer(kind RpcKind) *Server {
	return &Server{Kind: kind}
}

// SysRpcServer is the system namespace rpc instance of *Server.

// UserRpcServer is the user namespace rpc instance of *Server

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
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
	Sys:  "SysRpc",  // system rpc
	User: "UserRpc", // user rpc
}

func (k RpcKind) String() string {
	if int(k) < len(rpcKindNames) {
		return rpcKindNames[k]
	}
	return strconv.Itoa(int(k))
}
