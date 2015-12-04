package starx

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"starx/rpc"
)

type RpcStatus int32

const (
	RPC_STATUS_UNINIT RpcStatus = iota
	RPC_STATUS_INITED
)

type remoteService struct {
	Name         string
	ClientIdMaps map[string]*rpc.Client
	Route        map[string]func(string) uint32
	Status       RpcStatus
}

type unhandledRequest struct {
	bs *remoteSession
	rr *rpc.Request
}

func newRemote() *remoteService {
	return &remoteService{
		Name:         "RpcComponent",
		ClientIdMaps: make(map[string]*rpc.Client),
		Route:        make(map[string]func(string) uint32),
		Status:       RPC_STATUS_UNINIT}
}

func (rs *remoteService) register(comp RpcComponent) {
	comp.Setup()
	rpc.Register(comp)
}

// Server handle request
func (rs *remoteService) handle(conn net.Conn) {
	defer conn.Close()
	// message buffer
	requestChan := make(chan *unhandledRequest, packetBufferSize)
	// all user logic will be handled in single goroutine
	// synchronized in below routine
	go func() {
		for r := range requestChan {
			rs.processRequest(r.bs, r.rr)
		}
	}()

	bs := Net.createRemoteSession(conn)
	Net.dumpRemoteSessions()
	tmp := make([]byte, 0) // save truncated data
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("session closed(" + err.Error() + ")")
			bs.status = SS_CLOSED
			Net.closeSession(bs.userSession)
			Net.dumpHandlerSessions()
			break
		}
		tmp = append(tmp, buf[:n]...)
		var rr *rpc.Request // save decoded packet
		// TODO
		// Refactor this loop
		// read all request from buffer, and send to handle queue
		for len(tmp) > headLength {
			if rr, tmp = readRequest(tmp); rr != nil {
				fmt.Println("readRequest loop")
				requestChan <- &unhandledRequest{bs, rr}
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

func (rs *remoteService) processRequest(bs *remoteSession, rr *rpc.Request) {
	fmt.Printf("%+v\n", rr)
	if rr.Namespace == "sys" {
		fmt.Println(string(rr.Args))
	} else if rr.Namespace == "user" {
		var args interface{}
		json.Unmarshal(rr.Args, &args)
		fmt.Printf("%#v\n", args)
	} else {
		Error("invalid rpc namespace")
	}
}

func (rs *remoteService) asyncRequest(route *routeInfo, session *Session, args ...interface{}) {

}

// Client send request
// First argument is namespace, can be set `user` or `sys`
func (this *remoteService) request(ns string, route *routeInfo, session *Session, args []byte) ([]byte, error) {
	client, err := this.getClientByType(route.server, session)
	if err != nil {
		Info(err.Error())
		return nil, err
	}
	var reply *[]byte
	err = client.Call(ns, route.service, route.method, session.Id, reply, args)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return *reply, nil
}

func (this *remoteService) closeClient(svrId string) {
	if client, ok := this.ClientIdMaps[svrId]; ok {
		delete(this.ClientIdMaps, svrId)
		client.Close()
	} else {
		Info(fmt.Sprintf("%s not found in rpc client list", svrId))
	}

	Info(fmt.Sprintf("%s rpc client has been removed.", svrId))
	this.dumpClientIdMaps()
}

func (rs *remoteService) close() {
	// close rpc clients
	Info("close all of socket connections")
	for svrId, _ := range rs.ClientIdMaps {
		rs.closeClient(svrId)
	}
}

// TODO: add another argment session, to select a exact server when the
// server type has more than one server
// all established `rpc.Client` will be disconnected in `App.Stop()`
func (this *remoteService) getClientByType(svrType string, session *Session) (*rpc.Client, error) {
	if svrType == App.Config.Type {
		return nil, errors.New(fmt.Sprintf("current server has the same type(Type: %s)", svrType))
	}
	svrIds := SvrTypeMaps[svrType]
	if nums := len(svrIds); nums > 0 {
		if fn := Route[svrType]; fn != nil {
			// try to get user-define route function
			return this.getClientById(fn(session))
		} else {
			// if can not abtain user-define route function,
			// select a random server establish rpc connection
			random := rand.Intn(nums)
			return this.getClientById(svrIds[random])
		}
	}
	return nil, errors.New("not found rpc client")
}

// Get rpc client by server id(`connector-server-1`), return correspond rpc
// client if remote server connection has established already, or try to
// connect remote server when remote server network connectoin has not made
// by now, and return a nil value when server id not found or target machine
// refuse it.
func (this *remoteService) getClientById(svrId string) (*rpc.Client, error) {
	client := this.ClientIdMaps[svrId]
	if client != nil {
		Info("already exists")
		return client, nil
	}
	if svr, ok := SvrIdMaps[svrId]; ok && svr != nil {
		if svr.Id == App.Config.Id {
			return nil, errors.New(svr.Id + " is current server")
		}
		if svr.IsFrontend {
			return nil, errors.New(svr.Id + " is frontend server, can handle rpc request")
		}
		client, err := rpc.Dial("tcp4", fmt.Sprintf("%s:%d", svr.Host, svr.Port))
		if err != nil {
			return nil, err
		}
		this.ClientIdMaps[svr.Id] = client
		Info(fmt.Sprintf("%s establish rpc client successful.", svr.Id))
		this.dumpClientIdMaps()
		return client, nil
	}
	return nil, errors.New(fmt.Sprintf("server id does not exists(Id: %s)", svrId))
}

// Dump all clients that has established netword connection with remote server
func (this *remoteService) dumpClientIdMaps() {
	for id, _ := range this.ClientIdMaps {
		Info(fmt.Sprintf("[%s] is contained in rpc client list", id))
	}
}
