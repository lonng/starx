package network

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"reflect"
	"runtime/debug"

	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network/route"
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
			defaultNetService.dumpAcceptor()
			acceptor.close()
			endChan <- true
			break
		}
		tmp = append(tmp, buf[:n]...)
		// TODO
		// Refactor this loop
		// read all request from buffer, and send to handle queue
		for {
			rr := &rpc.Request{} // save decoded packet
			if tmp, err = rr.UnmarshalMsg(tmp); err != nil {
				break
			} else {
				requestChan <- &unhandledRequest{acceptor, rr}
			}
		}
	}
}

func isSessionClosedRequest(rr *rpc.Request) bool {
	return rr.ServiceMethod == sessionClosedRoute
}

func (rs *remoteService) processRequest(ac *acceptor, rr *rpc.Request) {
	// session closed notify request
	if isSessionClosedRequest(rr) {
		defaultNetService.closeSession(ac.Session(rr.Sid))
		return
	}

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
		log.Error(err.Error())
		response.Error = err.Error()
		goto WRITE_RESPONSE
	}

	service, ok = rs.serviceMap[route.Service]
	if !ok || service == nil {
		str := "remote: servive " + route.Service + " does not exists"
		log.Error(str)
		response.Error = str
		goto WRITE_RESPONSE
	}

	switch rr.Kind {
	case rpc.Sys:
		session := ac.Session(rr.Sid)
		m, ok := service.handlerMethod[route.Method]
		if !ok || m == nil {
			str := "remote: service " + route.Service + "does not contain method: " + route.Method
			log.Error(str)
			response.Error = str
			goto WRITE_RESPONSE
		}
		var data interface{}
		if m.raw {
			data = rr.Data
		} else {
			data = reflect.New(m.dataType.Elem()).Interface()
			err := serializer.Deserialize(rr.Data, data)
			if err != nil {
				str := "deserialize error: " + err.Error()
				log.Error(str)
				response.Error = str
				goto WRITE_RESPONSE
			}
		}

		ret, err := rs.call(m.method, []reflect.Value{
			service.rcvr,
			reflect.ValueOf(session),
			reflect.ValueOf(data)})
		if err != nil {
			log.Error(err.Error())
			response.Error = err.Error()
		} else {
			// handler method encounter error
			if err := ret[0].Interface(); err != nil {
				log.Error(err.(error).Error())
				response.Error = err.(error).Error()
			}
		}
	case rpc.User:
		var args interface{}
		var params = []reflect.Value{}
		json.Unmarshal(rr.Data, &args)
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
	if err := rpc.WriteResponse(ac.socket, response); err != nil {
		log.Error(err.Error())
	}
}

func (rs *remoteService) call(method reflect.Method, args []reflect.Value) (rets []reflect.Value, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Error("rpc call error: %+v", rec)
			os.Stderr.Write(debug.Stack())
			if s, ok := rec.(string); ok {
				err = errors.New(s)
			} else {
				err = errors.New("RpcCall internal error")
			}
		}
	}()
	rets = method.Func.Call(args)
	return rets, nil
}

func (rs *remoteService) dumpServiceMap() {
	for sname, s := range rs.serviceMap {
		for mname, _ := range s.handlerMethod {
			log.Info("registered service: %s.%s", sname, mname)
		}

		for mname, _ := range s.remoteMethod {
			log.Info("registered service: %s.%s", sname, mname)
		}
	}
}
