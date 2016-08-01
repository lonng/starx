package cluster

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/chrislonng/starx/cluster/rpc"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/session"
)

var (
	svrLock     sync.RWMutex             // protect server collections operations
	svrTypes    []string                 // all server type
	svrTypeMaps map[string][]string      // all servers type maps
	svrIdMaps   map[string]*ServerConfig // all servers id maps

	mutex        sync.RWMutex           // protect ClientIdMaps
	clientIdMaps map[string]*rpc.Client // all rpc clients
	appConfig    *ServerConfig          // current app config

	sessionManger SessionManager //get session instance
)

var (
	ErrServerNotFound = errors.New("server config not found")
)

type SessionManager interface {
	Session(sid uint64) (*session.Session, error)
}

func init() {
	svrTypeMaps = make(map[string][]string)
	svrIdMaps = make(map[string]*ServerConfig)
	clientIdMaps = make(map[string]*rpc.Client)
}

func DumpSvrIdMaps() {
	svrLock.RLock()
	defer svrLock.RUnlock()

	for _, v := range svrIdMaps {
		log.Infof("id: %s(%s)", v.Id, v.String())
	}
}

func DumpSvrTypeMaps() {
	svrLock.RLock()
	defer svrLock.RUnlock()

	for _, t := range svrTypes {
		svrs := svrTypeMaps[t]
		if len(svrs) == 0 {
			continue
		}
		for _, svrId := range svrs {
			log.Infof("server type: %s, id: %s", t, svrId)
		}
	}
}

func Register(server *ServerConfig) {
	svrLock.Lock()
	defer svrLock.Unlock()

	// server exists
	if _, ok := svrIdMaps[server.Id]; ok {
		log.Infof("serverId: %s already existed(%s)", server.Id, server.String())
		return
	}

	svr := server
	if len(svrTypes) > 0 {
		for k, t := range svrTypes {
			// duplicate
			if t == svr.Type {
				break
			}

			// arrive slice end
			if k == len(svrTypes)-1 {
				svrTypes = append(svrTypes, svr.Type)
			}
		}
	} else {
		svrTypes = append(svrTypes, svr.Type)
	}

	svrIdMaps[svr.Id] = svr
	svrTypeMaps[svr.Type] = append(svrTypeMaps[svr.Type], svr.Id)
}

func RemoveServer(svrId string) {
	svrLock.Lock()
	defer svrLock.Unlock()

	svr, ok := svrIdMaps[svrId]
	if !ok || svr == nil {
		log.Infof("serverId: %s not found", svrId)
		return
	}

	// remove from ServerIdMaps map
	typ := svr.Type
	svrs, ok := svrTypeMaps[typ]

	if !ok || len(svrs) == 0 {
		log.Infof("server type: %s has not instance", typ)
		return
	}

	if len(svrs) == 1 { // array only one element, remove it directly
		delete(svrTypeMaps, typ)
	} else {
		var tempSvrs []string
		for idx, id := range svrs {
			if id == svrId {
				tempSvrs = append(tempSvrs, svrs[:idx]...)
				tempSvrs = append(tempSvrs, svrs[(idx+1):]...)
				break
			}
		}
		svrTypeMaps[typ] = tempSvrs
	}

	// remove from ServerIdMaps
	delete(svrIdMaps, svrId)
	CloseClient(svrId)
}

func Server(id string) (*ServerConfig, error) {
	svrLock.RLock()
	defer svrLock.RUnlock()

	svr, ok := svrIdMaps[id]
	if !ok {
		return nil, ErrServerNotFound
	}

	return svr, nil
}

func UpdateServer(newSvr *ServerConfig) {
	svrLock.Lock()
	defer svrLock.Unlock()

	svr, ok := svrIdMaps[newSvr.Id]
	if !ok || svr == nil {
		log.Errorf(newSvr.Id + " not exists")
		return
	}

	svrIdMaps[svr.Id] = newSvr
}

func CloseClient(svrId string) {
	mutex.Lock()
	defer mutex.Unlock()

	client, ok := clientIdMaps[svrId]
	if !ok {
		log.Infof("%s not found in rpc client list", svrId)
		return
	}

	delete(clientIdMaps, svrId)
	client.Close()

	log.Infof("%s rpc client has been removed.", svrId)
	DumpClientIdMaps()
}

func ClientByType(svrType string, session *session.Session) (*rpc.Client, error) {
	if svrType == appConfig.Type {
		return nil, errors.New(fmt.Sprintf("current server has the same type(Type: %s)", svrType))
	}

	// fast mode
	if id := session.ServerID(svrType); id != "" {
		return Client(id)
	}

	// slow mode
	svrIds := svrTypeMaps[svrType]
	if n := len(svrIds); n > 0 {
		var id string
		if fn := router[svrType]; fn != nil {
			// try to get user-define router function
			id = fn(session)
		} else {
			// select a random server when could not found user-define router
			r := rand.Intn(n)
			id = svrIds[r]
		}

		session.SetServerID(svrType, id)
		return Client(id)
	}

	return nil, errors.New("not found rpc client")
}

// Get RPC client by server id(`connector-server-1`), and return the client if
// remote server connection has established already, or try to connect the
// remote server when remote server network connections have not made by now,
// and return a nil value when server id not found or target machine refuse it.
func Client(svrId string) (*rpc.Client, error) {
	mutex.RLock()
	client, ok := clientIdMaps[svrId]
	mutex.RUnlock()

	if ok && client != nil {
		return client, nil
	}

	svr, ok := svrIdMaps[svrId]
	if !ok || svr == nil {
		return nil, errors.New(fmt.Sprintf("server id does not exists(Id: %s)", svrId))

	}

	// current server
	if svr.Id == appConfig.Id {
		return nil, errors.New(svr.Id + " is current server")
	}

	// frontend server
	if svr.IsFrontend {
		return nil, errors.New(svr.Id + " is frontend server, can handle rpc request")
	}

	client, err := rpc.Dial("tcp4", fmt.Sprintf("%s:%d", svr.Host, svr.Port))
	if err != nil {
		return nil, err
	}
	log.Infof("%s establish rpc client successful.", svr.Id)

	// on client shutdown
	client.OnShutdown(func() {
		RemoveServer(svr.Id)
	})

	mutex.Lock()
	clientIdMaps[svr.Id] = client
	mutex.Unlock()

	// handle sys rpc push/response
	go func() {
		for resp := range client.ResponseChan {
			s, err := sessionManger.Session(resp.Sid)
			if err != nil {
				log.Errorf(err.Error())
				continue
			}

			switch resp.Kind {
			case rpc.HandlerPush:
				s.Push(resp.Route, resp.Data)
			case rpc.HandlerResponse:
				s.Response(resp.Data)
			default:
				log.Errorf("invalid response kind")
			}
		}
	}()

	return client, nil
}

// Dump all clients that has established netword connection with remote server
func DumpClientIdMaps() {
	mutex.RLock()
	defer mutex.RUnlock()

	for id, _ := range clientIdMaps {
		log.Infof("[%s] is contained in rpc client list", id)
	}
}

func Close() {
	mutex.Lock()
	mutex.Unlock()

	// close all RPC clients
	log.Infof("close all of socket connections")
	for id, client := range clientIdMaps {
		delete(clientIdMaps, id)
		client.Close()
	}
}

func SetAppConfig(c *ServerConfig) {
	appConfig = c
}

func SetSessionManager(s SessionManager) {
	if s == nil {
		panic("nil session manager")
	}
	sessionManger = s
}
