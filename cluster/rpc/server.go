package rpc

import (
	"errors"
	"strconv"
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	ErrNilResponse = errors.New("nil response")
)

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
