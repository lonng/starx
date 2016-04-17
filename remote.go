package starx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chrislonng/starx/rpc"
	"net"
	"reflect"
)

type rpcStatus int32

const (
	_RPC_STATUS_UNINIT rpcStatus = iota
	_RPC_STATUS_INITED
)

type remoteService struct {
	name   string
	route  map[string]func(string) uint32
	status rpcStatus
}

type unhandledRequest struct {
	bs *acceptor
	rr *rpc.Request
}

func newRemote() *remoteService {
	return &remoteService{
		name:   "RpcComponent",
		route:  make(map[string]func(string) uint32),
		status: _RPC_STATUS_UNINIT}
}

func (rs *remoteService) register(rpcKind rpc.RpcKind, comp Component) {
	comp.Init()
	if rpcKind == rpc.SysRpc {
		rpc.SysRpcServer.Register(comp)
	} else if rpcKind == rpc.UserRpc {
		rpc.UserRpcServer.Register(comp)
	} else {
		Error("invalid rpc kind")
	}
}

// Server handle request
func (rs *remoteService) handle(conn net.Conn) {
	defer conn.Close()
	// message buffer
	requestChan := make(chan *unhandledRequest, packetBufferSize)
	endChan := make(chan bool, 1)
	// all user logic will be handled in single goroutine
	// synchronized in below routine
	go func() {
		for {
			select {
			case r := <-requestChan:
				rs.processRequest(r.bs, r.rr)
			case <-endChan:
				close(requestChan)
				return
			}
		}
	}()

	acceptor := defaultNetService.createAcceptor(conn)
	defaultNetService.dumpAcceptor()
	tmp := make([]byte, 0) // save truncated data
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("session closed(" + err.Error() + ")")
			defaultNetService.dumpAgents()
			acceptor.close()
			endChan <- true
			break
		}
		tmp = append(tmp, buf[:n]...)
		var rr *rpc.Request // save decoded packet
		// TODO
		// Refactor this loop
		// read all request from buffer, and send to handle queue
		for len(tmp) > headLength {
			if rr, tmp = readRequest(tmp); rr != nil {
				requestChan <- &unhandledRequest{acceptor, rr}
			} else {
				break
			}
		}
	}
}

func readRequest(data []byte) (*rpc.Request, []byte) {
	var length uint
	var offset = 0
	for i := 0; i < len(data); i++ {
		b := data[i]
		length += (uint(b&0x7F) << uint(7*(i)))
		if b < 128 {
			offset = i + 1
			break
		}
	}
	request := rpc.Request{}
	err := json.Unmarshal(data[offset:(offset+int(length))], &request)
	if err != nil {
		//TODO
	}
	return &request, data[(offset + int(length)):]
}

func writeResponse(bs *acceptor, response *rpc.Response) {
	if response == nil {
		return
	}
	resp, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	buf := make([]byte, 0)
	length := len(resp)
	for {
		b := byte(length % 128)
		length >>= 7
		if length != 0 {
			buf = append(buf, b+128)
		} else {
			buf = append(buf, b)
			break
		}
	}
	buf = append(buf, resp...)
	bs.socket.Write(buf)
}

func (rs *remoteService) processRequest(bs *acceptor, rr *rpc.Request) {
	if rr.Kind == rpc.SysRpc {
		fmt.Println(string(rr.Args))
		session := bs.GetUserSession(rr.Sid)
		returnValues, err := rpc.SysRpcServer.Call(rr.ServiceMethod, []reflect.Value{reflect.ValueOf(session), reflect.ValueOf(rr.Args)})
		response := &rpc.Response{}
		response.ServiceMethod = rr.ServiceMethod
		response.Seq = rr.Seq
		response.Sid = rr.Sid
		response.Kind = rpc.RemoteResponse
		if err != nil {
			response.Error = err.Error()
		} else {
			// handler method encounter error
			errInter := returnValues[0].Interface()
			if errInter != nil {
				response.Error = errInter.(error).Error()
			}
		}
		writeResponse(bs, response)
	} else if rr.Kind == rpc.UserRpc {
		var args interface{}
		var params = []reflect.Value{}
		json.Unmarshal(rr.Args, &args)
		switch args.(type) {
		case []interface{}:
			for _, arg := range args.([]interface{}) {
				params = append(params, reflect.ValueOf(arg))
			}
		default:
			fmt.Println("invalid rpc argument")
		}
		rets, err := rpc.UserRpcServer.Call(rr.ServiceMethod, params)
		response := &rpc.Response{}
		response.ServiceMethod = rr.ServiceMethod
		response.Seq = rr.Seq
		response.Sid = rr.Sid
		response.Kind = rpc.RemoteResponse
		if err != nil {
			response.Error = err.Error()
		} else {
			// handler method encounter error
			errInter := rets[1].Interface()
			if errInter != nil {
				response.Error = errInter.(error).Error()
			} else {
				response.Data = rets[0].Bytes()
			}
		}
		writeResponse(bs, response)
	} else {
		Error("invalid rpc namespace")
	}
}

func (rs *remoteService) asyncRequest(route *routeInfo, session *Session, args ...interface{}) {

}

// Client send request
// First argument is namespace, can be set `user` or `sys`
func (this *remoteService) request(rpcKind rpc.RpcKind, route *routeInfo, session *Session, args []byte) ([]byte, error) {
	client, err := cluster.getClientByType(route.serverType, session)
	if err != nil {
		Info(err.Error())
		return nil, err
	}
	reply := new([]byte)
	err = client.Call(rpcKind, route.service, route.method, session.entityID, reply, args)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return *reply, nil
}
