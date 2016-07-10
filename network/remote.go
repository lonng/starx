package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"runtime/debug"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network/route"
	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/session"
	"github.com/chrislonng/starx/network/packet"
)

var (
	ErrNilResponse = errors.New("nil response")
)

var Remote = newRemote()

type remoteService struct {
	serviceMap map[string]*service // all handler service
}

type unhandledRequest struct {
	bs *acceptor
	rr *rpc.Request
}

func newRemote() *remoteService {
	return &remoteService{
		serviceMap: make(map[string]*service),
	}
}

func (remote *remoteService) Register(rcvr Component) error {
	s := &service{
		typ:  reflect.TypeOf(rcvr),
		rcvr: reflect.ValueOf(rcvr),
	}
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if !isExported(sname) {
		return errors.New("remote.Register: type " + sname + " is not exported")
	}
	if _, present := remote.serviceMap[sname]; present {
		return errors.New("remote: service already defined: " + sname)
	}
	s.name = sname

	// Install the handler methods
	s.handlerMethod = suitableHandlerMethods(s.typ, true)
	if len(s.handlerMethod) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableHandlerMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}

	// Install the remote methods
	s.remoteMethod = suitableRemoteMethods(s.typ, true)
	if len(s.handlerMethod) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableRemoteMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "remote.Register: type " + sname + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}
	remote.serviceMap[s.name] = s
	return nil
}

// Server handle request
func (rs *remoteService) Handle(conn net.Conn) {
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

func (rs *remoteService) processRequest(ac *acceptor, rr *rpc.Request) {
	var (
		err      error
		service  *service
		ok       bool
		response = &rpc.Response{
			ServiceMethod: rr.ServiceMethod,
			Seq:           rr.Seq,
			Sid:           rr.Sid,
			Kind:          rpc.RemoteResponse,
		}
	)

	route, err := route.Decode(rr.ServiceMethod)
	if err != nil {
		response.Error = err.Error()
		goto WRITE_RESPONSE
	}

	service, ok = rs.serviceMap[route.Service]
	if !ok || service == nil {
		response.Error = "remote: servive " + route.Service + " does not exists"
		goto WRITE_RESPONSE
	}

	switch rr.Kind {
	case rpc.Sys:
		fmt.Println(string(rr.Args))
		session := ac.Session(rr.Sid)
		m, ok := service.handlerMethod[route.Method]
		if !ok || m == nil {
			response.Error = "remote: service " + route.Service + "does not contain method: " + route.Method
			goto WRITE_RESPONSE
		}
		ret, err := rs.call(m.method, []reflect.Value{reflect.ValueOf(session), reflect.ValueOf(rr.Args)})
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
			log.Error("invalid rpc argument")
		}
		m, ok := service.handlerMethod[route.Method]
		if !ok || m == nil {
			response.Error = "remote: service " + route.Service + "does not contain method: " + route.Method
			goto WRITE_RESPONSE
		}
		ret, err := rs.call(m.method, params)
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
WRITE_RESPONSE:
	writeResponse(ac, response)
}

func (rs *remoteService) call(method reflect.Method, args []reflect.Value) (rets []reflect.Value, err error) {
	defer func() {
		if recov := recover(); recov != nil {
			log.Fatal("RpcCall Error: %+v", recov)
			os.Stderr.Write(debug.Stack())
			if s, ok := recov.(string); ok {
				err = errors.New(s)
			} else {
				err = errors.New("RpcCall internal error")
			}
		}
	}()
	rets = method.Func.Call(args)
	return rets, nil
}

func (rs *remoteService) asyncRequest(route *route.Route, session *session.Session, args ...interface{}) {

}

// Client send request
// First argument is namespace, can be set `user` or `sys`
func (rs *remoteService) request(rpcKind rpc.RpcKind, route *route.Route, session *session.Session, args []byte) ([]byte, error) {
	client, err := cluster.ClientByType(route.ServerType, session)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	reply := new([]byte)
	err = client.Call(rpcKind, route.Service, route.Method, session.Entity.ID(), reply, args)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return *reply, nil
}
