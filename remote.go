package starx

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"starx/rpc"
	"strings"
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

func newRemote() *remoteService {
	return &remoteService{
		Name:         "RpcComponent",
		ClientIdMaps: make(map[string]*rpc.Client),
		Route:        make(map[string]func(string) uint32),
		Status:       RPC_STATUS_UNINIT}
}

func (this *remoteService) register(comp RpcComponent) {
	comp.Setup()
	rpc.Register(comp)
}

func (this *remoteService) handle(conn net.Conn) {
	defer conn.Close()
	rpc.ServeConn(conn)
}

func (this *remoteService) request(route string, session *Session, data []byte) {
	routeArgs := strings.Split(route, ".")
	if len(routeArgs) != 3 {
		Error(fmt.Sprintf("wrong route: `%s`", route))
	}
	client, err := this.getClientByType(routeArgs[0], session)
	if err != nil {
		Info(err.Error())
		return
	}
	req := "hello"
	var rep int
	e := client.Call(routeArgs[1]+"."+routeArgs[2], &req, &rep)
	Info(fmt.Sprint("reply value: %d", rep))
	if e != nil {
		Info(e.Error())
	}
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

func (this *remoteService) close() {
	// close rpc clients
	Info("close all of socket connections")
	for svrId, _ := range this.ClientIdMaps {
		this.closeClient(svrId)
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
