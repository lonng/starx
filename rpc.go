package mello

import (
	"errors"
	"fmt"
	"net"
	"mello/rpc"
	"strings"
)

type RpcStatus int32

const (
	RPC_STATUS_UNINIT RpcStatus = iota
	RPC_STATUS_INITED
)

type RpcService struct {
	Name           string
	ClientTypeMaps map[string][]*rpc.Client
	ClientIdMaps   map[string]*rpc.Client
	Route          map[string]func(string) uint32
	Status         RpcStatus
}

type RpcComponent interface {
	Setup()
}

func NewRpc() *RpcService {
	return &RpcService{
		Name:           "RpcComponent",
		ClientTypeMaps: make(map[string][]*rpc.Client),
		ClientIdMaps:   make(map[string]*rpc.Client),
		Route:          make(map[string]func(string) uint32),
		Status:         RPC_STATUS_UNINIT}
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
		Error(err.Error())
		return
	}
	req := "hello"
	e := client.Call("Manager.Test", &req, nil)
	if e != nil {
		Error(e.Error())
	}
}

func (this *RpcService) CloseClient(svrId string) {
	// TODO:
	if client, ok := this.ClientIdMaps[svrId]; ok {
		// delete client from `Rpc.ClientIdMaps`
		for _, c := range this.ClientIdMaps {
			if c == client {
				delete(this.ClientIdMaps, svrId)
				break
			}
		}

		isDeleted := false
		for svrType, clients := range this.ClientTypeMaps {
			if len(clients) == 0 {
				delete(this.ClientTypeMaps, svrType)
				continue
			}
			for idx, c := range clients {
				if c == client {
					this.ClientTypeMaps[svrType] = append(clients[:idx], clients[(idx+1):]...)
					isDeleted = true
					break
				}
			}

			if isDeleted {
				break
			}
		}
		// close client
		client.Close()
	} else {
		Info(fmt.Sprintf("serverId: %s not found in rpc list", svrId))
	}

	Info(fmt.Sprintf("ServerId: %s rpc client has removed.", svrId))
	this.dumpClientTypeMaps()
	this.dumpClientIdMaps()
}

func (this *RpcService) AddClient(svr *ServerConfig) {
	client, err := rpc.Dial("tcp4", fmt.Sprintf("%s:%d", svr.Host, svr.Port))
	if err != nil {
		Info(fmt.Sprintf("ServerId: %s establish rpc client failed.", svr.Id))
		return
	}
	this.ClientIdMaps[svr.Id] = client
	this.ClientTypeMaps[svr.Type] = append(this.ClientTypeMaps[svr.Type], client)
	Info(fmt.Sprintf("ServerId: %s establish rpc client successful.", svr.Id))
	this.dumpClientTypeMaps()
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

	if this.Status == RPC_STATUS_UNINIT {
		this.initClient()
	}

	if fn := Route[svrType]; fn != nil {
		return this.getClientById(fn())
	}

	clients := this.ClientTypeMaps[svrType]
	if len(clients) == 0 {
		this.dumpClientTypeMaps()
		Error(fmt.Sprintf("RpcClient: `type: %s` not found", svrType))
		return nil, errors.New("rpc client not found")
	}

	return this.ClientTypeMaps[svrType][0], nil
}

func (this *RpcService) getClientById(svrId string) (*rpc.Client, error) {
	if this.Status == RPC_STATUS_UNINIT {
		this.initClient()
	}
	client := this.ClientIdMaps[svrId]
	if client == nil {
		return nil, errors.New(fmt.Sprintf("serverId does not exists(Id: %s)", svrId))
	}
	return client, nil
}

// establish all server in `App.ServerIdMaps` rpc client
func (this *RpcService) initClient() {
	for _, svrPtr := range SvrIdMaps {
		if svrPtr.Id != App.CurSvrConfig.Id {
			this.AddClient(svrPtr)
		}
	}
	this.Status = RPC_STATUS_INITED
}

func (this *RpcService) dumpClientTypeMaps() {
	for t, cs := range this.ClientTypeMaps {
		if len(cs) == 0 {
			delete(this.ClientTypeMaps, t)
			continue
		}
		Info(fmt.Sprintf("Rpc.ClientTypeMaps[%s] current length: %d", t, len(cs)))
	}
}

func (this *RpcService) dumpClientIdMaps() {
	for id, _ := range this.ClientIdMaps {
		Info(fmt.Sprintf("Rpc.ClientIdMaps[%s] is contained in", id))
	}
}
