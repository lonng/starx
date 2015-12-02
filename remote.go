package starx

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"starx/rpc"
	"encoding/json"
)

type RpcStatus int32

const (
	RPC_STATUS_UNINIT RpcStatus = iota
	RPC_STATUS_INITED
)

const (
	remoteRequestHeadLength = 2
	remoteResponseHeadLength = 2
)

type remoteService struct {
	Name         string
	ClientIdMaps map[string]*rpc.Client
	Route        map[string]func(string) uint32
	Status       RpcStatus
}

type remoteRequest struct {
	namespace string
	seq uint64
}

type remoteResponse struct {
	seq uint64
}

type unhandledRemoteRequest struct {
	bs *backendSession
	rr *remoteRequest
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

func (rs *remoteService) handle(conn net.Conn) {
	defer conn.Close()
	requestChan := make(chan *unhandledRemoteRequest, packetBufferSize)
	go func(){
		for urr := range requestChan {
			rs.processPacket(urr.bs, urr.rr)
		}
	}()

	bs := Net.createBackendSession(conn)
	Net.dumpBackendSessions()
	tmp := make([]byte, 512) // save truncated data
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("session closed(" + err.Error() + ")")
			bs.status = SS_CLOSED
			Net.closeSession(bs.userSession)
			Net.dumpFrontendSessions()
			break
		}
		tmp = append(tmp, buf[:n]...)
		var rr *remoteRequest // save decoded packet
		// TODO
		// Refactor this loop
		for len(tmp) > headLength {
			if rr, tmp = readRemoteRequest(tmp); rr != nil {
				requestChan <- &unhandledRemoteRequest{bs, rr}
			} else {
				break
			}
		}
	}
	//rpc.ServeConn(conn)
}

func readRemoteRequest(data []byte) (*remoteRequest, []byte){
	if len(data) < remoteRequestHeadLength {
		return nil, data
	}
	length := bytesToInt(data[:remoteRequestHeadLength])
	if len(data) < remoteRequestHeadLength + length {
		return nil, data
	}else {
		rr := new(remoteRequest)
		err := json.Unmarshal(data[remoteRequestHeadLength:(remoteRequestHeadLength + length)], rr)
		if err != nil {
			Error(err.Error())
			return nil, data[(remoteRequestHeadLength + length):]
		}
		return rr, data[(remoteRequestHeadLength + length):]
	}
}

func (rs *remoteService) processPacket(bs *backendSession, rr *remoteRequest){

}

func (rs *remoteService) asyncRequest(route *routeInfo, session *Session, args ... interface{}) {

}

func (this *remoteService) request(route *routeInfo, session *Session, args ...interface{}) {
	client, err := this.getClientByType(route.server, session)
	if err != nil {
		Info(err.Error())
		return
	}
	req := "hello"
	var rep int
	e := client.Call(route.service+"."+route.method, &req, &rep)
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
