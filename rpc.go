package mello

import (
	"errors"
	"fmt"
	"mello/rpc"
	"net"
	"strings"
	"time"
)

type RpcStatus int32

const (
	RPC_STATUS_UNINIT RpcStatus = iota
	RPC_STATUS_INITED
)

type RpcService struct {
	Name         string
	ClientIdMaps map[string]*rpc.Client
	Route        map[string]func(string) uint32
	Status       RpcStatus
}

type RpcComponent interface {
	Setup()
}

func NewRpc() *RpcService {
	return &RpcService{
		Name:         "RpcComponent",
		ClientIdMaps: make(map[string]*rpc.Client),
		Route:        make(map[string]func(string) uint32),
		Status:       RPC_STATUS_UNINIT}
}

func (this *RpcService) Register(comp RpcComponent) {
	comp.Setup()
	rpc.Register(comp)
}

func (this *RpcService) Handle(conn net.Conn) {
	defer conn.Close()
	rpc.ServeConn(conn)
}

func (this *RpcService) Request(route string) {
	routeArgs := strings.Split(route, ".")
	if len(routeArgs) != 3 {
		Error(fmt.Sprintf("wrong route: `%s`", route))
	}
	client, err := this.getClientByType(routeArgs[0])
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

func (this *RpcService) CloseClient(svrId string) {
	if client, ok := this.ClientIdMaps[svrId]; ok {
		delete(this.ClientIdMaps, svrId)
		client.Close()
	} else {
		Info(fmt.Sprintf("serverId: %s not found in rpc list", svrId))
	}

	Info(fmt.Sprintf("ServerId: %s rpc client has removed.", svrId))
	this.dumpClientIdMaps()
}

func (this *RpcService) Close() {
	// close rpc clients
	Info("close all of socket connections")
	for svrId, _ := range this.ClientIdMaps {
		this.CloseClient(svrId)
	}
}

// TODO: add another argment session, to select a exact server when the
// server type has more than one server
// all established `rpc.Client` will be disconnected in `App.Stop()`
func (this *RpcService) getClientByType(svrType string) (*rpc.Client, error) {
	if svrType == App.CurSvrConfig.Type {
		return nil, errors.New(fmt.Sprintf("current server has the same type(Type: %s)", svrType))
	}
	svrIds := SvrTypeMaps[svrType]
	if nums := len(svrIds); nums > 0 {
		if fn := Route[svrType]; fn != nil {
			return this.getClientById(fn())
		} else {
			curTime := time.Now().Unix()
			idx := curTime % int64(nums)
			return this.getClientById(svrIds[idx])
		}
	}
	return nil, errors.New("not found rpc client")
}

func (this *RpcService) getClientById(svrId string) (*rpc.Client, error) {
	client := this.ClientIdMaps[svrId]
	if client != nil {
		Info("already exists")
		return client, nil
	}
	if svr, ok := SvrIdMaps[svrId]; ok && svr != nil {
		if svr.Id == App.CurSvrConfig.Id {
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
		Info(fmt.Sprintf("ServerId: %s establish rpc client successful.", svr.Id))
		this.dumpClientIdMaps()
		return client, nil
	}
	return nil, errors.New(fmt.Sprintf("serverId does not exists(Id: %s)", svrId))
}

func (this *RpcService) dumpClientIdMaps() {
	for id, _ := range this.ClientIdMaps {
		Info(fmt.Sprintf("Rpc.ClientIdMaps[%s] is contained in", id))
	}
}
