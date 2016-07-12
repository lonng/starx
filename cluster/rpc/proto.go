package rpc

//go:generate msgp
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
	Data          []byte  // for args
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
