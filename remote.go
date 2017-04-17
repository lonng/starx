package starx

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net"
	"os"
	"reflect"
	"runtime/debug"

	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/route"
)

var remote = newRemote()

type remoteService struct {
	serviceMap map[string]*component.Service // all handler service
}

type unhandledRequest struct {
	bs *acceptor
	rr *rpc.Request
}

func newRemote() *remoteService {
	return &remoteService{
		serviceMap: make(map[string]*component.Service),
	}
}

func (rs *remoteService) register(rcvr component.Component) error {
	if rs.serviceMap == nil {
		rs.serviceMap = make(map[string]*component.Service)
	}

	s := &component.Service{
		Type: reflect.TypeOf(rcvr),
		Rcvr: reflect.ValueOf(rcvr),
	}
	s.Name = reflect.Indirect(s.Rcvr).Type().Name()
	if _, present := rs.serviceMap[s.Name]; present {
		return errors.New("remote: service already defined: " + s.Name)
	}

	if err := s.ScanHandler(); err != nil {
		return err
	}

	if err := s.ScanRemote(); err != nil {
		return err
	}
	rs.serviceMap[s.Name] = s
	return nil
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
			log.Infof("session closed(" + err.Error() + ")")
			defaultNetService.dumpAcceptor()
			acceptor.Close()
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
	var session = ac.Session(rr.Sid)

	// session closed notify request
	if isSessionClosedRequest(rr) {
		defaultNetService.closeSession(session)
		return
	}

	var (
		err      error
		service  *component.Service
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
		log.Errorf(err.Error())
		response.Error = err.Error()
		goto WRITE_RESPONSE
	}

	service, ok = rs.serviceMap[route.Service]
	if !ok || service == nil {
		str := "remote: servive " + route.Service + " does not exists"
		log.Errorf(str)
		response.Error = str
		goto WRITE_RESPONSE
	}

	switch rr.Kind {
	case rpc.Sys:
		m, ok := service.HandlerMethods[route.Method]
		if !ok || m == nil {
			str := "remote: service " + route.Service + "does not contain method: " + route.Method
			log.Errorf(str)
			response.Error = str
			goto WRITE_RESPONSE
		}
		var data interface{}
		if m.Raw {
			data = rr.Data
		} else {
			data = reflect.New(m.Type.Elem()).Interface()
			err := serializer.Deserialize(rr.Data, data)
			if err != nil {
				str := "deserialize error: " + err.Error()
				log.Errorf(str)
				response.Error = str
				goto WRITE_RESPONSE
			}
		}

		ret, err := rs.call(m.Method, []reflect.Value{
			service.Rcvr,
			reflect.ValueOf(session),
			reflect.ValueOf(data)})
		if err != nil {
			log.Errorf(err.Error())
			response.Error = err.Error()
		} else {
			// handler method encounter error
			if err := ret[0].Interface(); err != nil {
				log.Errorf(err.(error).Error())
				response.Error = err.(error).Error()
			}
		}
	case rpc.User:
		var args []interface{}
		var params = []reflect.Value{service.Rcvr}
		//json.Unmarshal(rr.Data, &args)
		gob.NewDecoder(bytes.NewReader(rr.Data)).Decode(&args)

		for _, arg := range args {
			params = append(params, reflect.ValueOf(arg))
		}

		m, ok := service.RemoteMethods[route.Method]
		if !ok || m == nil {
			response.Error = "remote: service " + route.Service + " does not contain method: " + route.Method
			goto WRITE_RESPONSE
		}
		ret, err := rs.call(m.Method, params)
		if err != nil {
			response.Error = err.Error()
		} else {
			// handler method encounter error
			if err := ret[1].Interface(); err != nil {
				response.Error = err.(error).Error()
			} else {
				buf := bytes.NewBuffer([]byte(nil))
				if err := gob.NewEncoder(buf).Encode(ret[0].Interface()); err != nil {
					response.Error = err.Error()
					goto WRITE_RESPONSE
				}
				response.Data = buf.Bytes()
			}
		}
	default:
		log.Errorf("invalid rpc namespace")
		return
	}

WRITE_RESPONSE:
	if err := rpc.WriteResponse(ac.socket, response); err != nil {
		log.Errorf(err.Error())
	}
}

func (rs *remoteService) call(method reflect.Method, args []reflect.Value) (rets []reflect.Value, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Errorf("rpc call error: %+v", rec)
			os.Stderr.Write(debug.Stack())
			if s, ok := rec.(string); ok {
				err = errors.New(s)
			} else {
				err = errors.New("rpc call internal error")
			}
		}
	}()
	rets = method.Func.Call(args)
	return rets, nil
}

func (rs *remoteService) dumpServiceMap() {
	for sn, s := range rs.serviceMap {
		for mn := range s.HandlerMethods {
			log.Infof("registered service: %s.%s", sn, mn)
		}

		for mn := range s.RemoteMethods {
			log.Infof("registered service: %s.%s", sn, mn)
		}
	}
}
