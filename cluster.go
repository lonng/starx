package starx

import (
	"errors"
	"fmt"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network/rpc"
	"math/rand"
	"sync"
)

type clusterService struct {
	svrTypes     []string                 // all server type
	svrTypeMaps  map[string][]string      // all servers type maps
	svrIdMaps    map[string]*ServerConfig // all servers id maps
	lock         sync.RWMutex             // protect ClientIdMaps
	clientIdMaps map[string]*rpc.Client   // all rpc clients
}

func newClusterService() *clusterService {
	return &clusterService{
		svrTypeMaps:  make(map[string][]string),
		svrIdMaps:    make(map[string]*ServerConfig),
		clientIdMaps: make(map[string]*rpc.Client)}
}

func (c *clusterService) dumpSvrIdMaps() {
	for _, v := range c.svrIdMaps {
		log.Info("id: %s(%s)", v.Id, v.String())
	}
}

func (c *clusterService) dumpSvrTypeMaps() {
	for _, t := range c.svrTypes {
		svrs := c.svrTypeMaps[t]
		if len(svrs) == 0 {
			continue
		}
		for _, svrId := range svrs {
			log.Info("server type: %s, id: %s", t, svrId)
		}
	}
}

func (c *clusterService) registerServer(server ServerConfig) {
	// server exists
	if _, ok := c.svrIdMaps[server.Id]; ok {
		log.Info("serverId: %s already existed(%s)", server.Id, server.String())
		return
	}
	svr := &server
	if len(c.svrTypes) > 0 {
		for k, t := range c.svrTypes {
			// duplicate
			if t == svr.Type {
				break
			}
			// arrive slice end
			if k == len(c.svrTypes)-1 {
				c.svrTypes = append(c.svrTypes, svr.Type)
			}
		}
	} else {
		c.svrTypes = append(c.svrTypes, svr.Type)
	}
	c.svrIdMaps[svr.Id] = svr
	c.svrTypeMaps[svr.Type] = append(c.svrTypeMaps[svr.Type], svr.Id)
}

func (c *clusterService) removeServer(svrId string) {

	if _, ok := c.svrIdMaps[svrId]; ok {
		// remove from ServerIdMaps map
		typ := c.svrIdMaps[svrId].Type
		if svrs, ok := c.svrTypeMaps[typ]; ok && len(svrs) > 0 {
			if len(svrs) == 1 { // array only one element, remove it directly
				delete(c.svrTypeMaps, typ)
			} else {
				var tempSvrs []string
				for idx, id := range svrs {
					if id == svrId {
						tempSvrs = append(tempSvrs, svrs[:idx]...)
						tempSvrs = append(tempSvrs, svrs[(idx+1):]...)
						break
					}
				}
				c.svrTypeMaps[typ] = tempSvrs
			}
		}
		// remove from ServerIdMaps
		delete(c.svrIdMaps, svrId)
		c.closeClient(svrId)
	} else {
		log.Info("serverId: %s not found", svrId)
	}
}

func (c *clusterService) updateServer(newSvr ServerConfig) {
	if srv, ok := c.svrIdMaps[newSvr.Id]; ok && srv != nil {
		c.svrIdMaps[srv.Id] = &newSvr
	} else {
		log.Error(newSvr.Id + " not exists")
	}
}

func (c *clusterService) closeClient(svrId string) {
	if client, ok := c.clientIdMaps[svrId]; ok {
		c.lock.Lock()
		delete(c.clientIdMaps, svrId)
		c.lock.Unlock()
		client.Close()
	} else {
		log.Info("%s not found in rpc client list", svrId)
	}

	log.Info("%s rpc client has been removed.", svrId)
	c.dumpClientIdMaps()
}

// TODO: add another argment session, to select a exact server when the
// server type has more than one server
// all established `rpc.Client` will be disconnected in `App.Stop()`
func (c *clusterService) getClientByType(svrType string, session *Session) (*rpc.Client, error) {
	if svrType == App.Config.Type {
		return nil, errors.New(fmt.Sprintf("current server has the same type(Type: %s)", svrType))
	}
	svrIds := c.svrTypeMaps[svrType]
	if nums := len(svrIds); nums > 0 {
		if fn := route[svrType]; fn != nil {
			// try to get user-define route function
			return c.getClientById(fn(session))
		} else {
			// if can not abtain user-define route function,
			// select a random server establish rpc connection
			random := rand.Intn(nums)
			return c.getClientById(svrIds[random])
		}
	}
	return nil, errors.New("not found rpc client")
}

// Get rpc client by server id(`connector-server-1`), return correspond rpc
// client if remote server connection has established already, or try to
// connect remote server when remote server network connectoin has not made
// by now, and return a nil value when server id not found or target machine
// refuse it.
func (c *clusterService) getClientById(svrId string) (*rpc.Client, error) {
	c.lock.RLock()
	client := c.clientIdMaps[svrId]
	c.lock.RUnlock()
	if client != nil {
		return client, nil
	}
	if svr, ok := c.svrIdMaps[svrId]; ok && svr != nil {
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
		// on client shutdown
		client.OnShutdown(func() {
			c.removeServer(svr.Id)
		})
		c.lock.Lock()
		c.clientIdMaps[svr.Id] = client
		c.lock.Unlock()
		// handle sys rpc push/response
		go func() {
			for resp := range client.ResponseChan {
				agent, err := defaultNetService.getAgent(resp.Sid)
				if err != nil {
					log.Error(err.Error())
					continue
				}
				if resp.Kind == rpc.HandlerPush {
					agent.session.Push(resp.Route, resp.Data)
				} else if resp.Kind == rpc.HandlerResponse {
					agent.session.Response(resp.Data)
				} else if resp.Kind == rpc.RemotePush {
					// TODO
					// remote server push data
				} else {
					log.Error("invalid response kind")
				}
			}
		}()
		log.Info("%s establish rpc client successful.", svr.Id)
		c.dumpClientIdMaps()
		return client, nil
	}
	return nil, errors.New(fmt.Sprintf("server id does not exists(Id: %s)", svrId))
}

// Dump all clients that has established netword connection with remote server
func (c *clusterService) dumpClientIdMaps() {
	for id, _ := range c.clientIdMaps {
		log.Info("[%s] is contained in rpc client list", id)
	}
}

func (c *clusterService) close() {
	// close rpc clients
	log.Info("close all of socket connections")
	for svrId, _ := range c.clientIdMaps {
		c.closeClient(svrId)
	}
}
