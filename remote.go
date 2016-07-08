package starx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network"
	"github.com/chrislonng/starx/network/rpc"
	"github.com/chrislonng/starx/packet"
)

type rpcStatus int32

const (
	rpcStatusUninit rpcStatus = iota
	rpcStatusInited
)

var (
	ErrNilResponse = errors.New("nil response")
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
		status: rpcStatusUninit}
}

func (rs *remoteService) register(rpcKind rpc.RpcKind, comp Component) {
	comp.Init()
	if rpcKind == rpc.Sys {
		rpc.SysRpcServer.Register(comp)
	} else if rpcKind == rpc.User {
		rpc.UserRpcServer.Register(comp)
	} else {
		log.Error("invalid rpc kind")
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
			log.Info("session closed(" + err.Error() + ")")
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
		for len(tmp) > packet.HeadLength {
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

func writeResponse(bs *acceptor, response *rpc.Response) error {
	if response == nil {
		return ErrNilResponse
	}
	resp, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err.Error())
		return err
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
	_, err = bs.socket.Write(buf)
	return err
}

func (rs *remoteService) processRequest(bs *acceptor, rr *rpc.Request) {
	response := &rpc.Response{
		ServiceMethod: rr.ServiceMethod,
		Seq:           rr.Seq,
		Sid:           rr.Sid,
		Kind:          rpc.RemoteResponse,
	}

	switch rr.Kind {
	case rpc.Sys:
		fmt.Println(string(rr.Args))
		session := bs.Session(rr.Sid)
		ret, err := rpc.SysRpcServer.Call(rr.ServiceMethod, []reflect.Value{reflect.ValueOf(session), reflect.ValueOf(rr.Args)})
		if err != nil {
			response.Error = err.Error()
		} else {
			// handler method encounter error

			if err := ret[0].Interface(); err != nil {
				response.Error = err.(error).Error()
			}
		}
	case rpc.User:
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
		ret, err := rpc.UserRpcServer.Call(rr.ServiceMethod, params)
		if err != nil {
			response.Error = err.Error()
		} else {
			// handler method encounter error
			if err := ret[1].Interface(); err != nil {
				response.Error = err.(error).Error()
			} else {
				response.Data = ret[0].Bytes()
			}
		}
	default:
		log.Error("invalid rpc namespace")
		return
	}
	writeResponse(bs, response)
}

func (rs *remoteService) asyncRequest(route *network.Route, session *Session, args ...interface{}) {

}

// Client send request
// First argument is namespace, can be set `user` or `sys`
func (this *remoteService) request(rpcKind rpc.RpcKind, route *network.Route, session *Session, args []byte) ([]byte, error) {
	client, err := cluster.getClientByType(route.ServerType, session)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	reply := new([]byte)
	err = client.Call(rpcKind, route.Service, route.Method, session.entityID, reply, args)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return *reply, nil
}
