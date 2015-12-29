package starx

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"starx/utils"
	"time"
)

var VERSION = "0.0.1"

var (
	App               *_app // starx application
	AppName           string
	AppPath           string
	workPath          string
	appConfigPath     string
	serverConfigPath  string
	masterConfigPath  string
	StartTime         time.Time
	svrTypes          []string                                           // all server type
	svrTypeMaps       map[string][]string                                // all servers type maps
	svrIdMaps         map[string]*ServerConfig                           // all servers id maps
	settings          map[string][]func()                                // all settiings
	remote            *remoteService                                     // remote service
	handler           *handlerService                                    // hander
	netService        *_netService                                       // net service
	TimerManager      Timer                                              // timer component
	route             map[string]func(*Session) string                   // server route function
	channelServive    *ChannelServive                                    // channel service component
	ConnectionService *connectionService                                 // connection service component
	protocolState     ProtocolState                                      // current protocol state
	heartbeatInternal time.Duration                    = time.Second * 8 // beatheart time internal, second unit
	heartbeatService  *HeartbeatService                                  // beatheart service
	endRunning        chan bool                                          // wait for end application
)

type ServerConfig struct {
	Type       string
	Id         string
	Host       string
	Port       int32
	IsFrontend bool
	IsMaster   bool
}

func (this *ServerConfig) String() string {
	return fmt.Sprintf("Type: %s, Id: %s, Host: %s, Port: %d, IsFrontend: %t, IsMaster: %t",
		this.Type,
		this.Id,
		this.Host,
		this.Port,
		this.IsFrontend,
		this.IsMaster)
}

func dumpSvrIdMaps() {
	for _, v := range svrIdMaps {
		Info(fmt.Sprintf("id: %s(%s)", v.Id, v.String()))
	}
}

func dumpSvrTypeMaps() {
	for _, t := range svrTypes {
		svrs := svrTypeMaps[t]
		if len(svrs) == 0 {
			continue
		}
		for _, svrId := range svrs {
			Info(svrId)
		}
	}
}

func registerServer(server ServerConfig) {
	// server exists
	if _, ok := svrIdMaps[server.Id]; ok {
		Info(fmt.Sprintf("serverId: %s already existed(%s)", server.Id, server.String()))
		return
	}
	svr := &server
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

func removeServer(svrId string) {

	if _, ok := svrIdMaps[svrId]; ok {
		// remove from ServerIdMaps map
		typ := svrIdMaps[svrId].Type
		if svrs, ok := svrTypeMaps[typ]; ok && len(svrs) > 0 {
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
		}
		// remove from ServerIdMaps
		delete(svrIdMaps, svrId)
		remote.closeClient(svrId)
	} else {
		Info(fmt.Sprintf("serverId: %s not found", svrId))
	}
}

func updateServer(newSvr ServerConfig) {
	if srv, ok := svrIdMaps[newSvr.Id]; ok && srv != nil {
		svrIdMaps[srv.Id] = &newSvr
	} else {
		Error(newSvr.Id + " not exists")
	}
}

func init() {
	App = newApp()
	svrTypeMaps = make(map[string][]string)
	svrIdMaps = make(map[string]*ServerConfig)
	settings = make(map[string][]func())
	StartTime = time.Now()
	Log = log.New(os.Stdout, "", log.LstdFlags)
	remote = newRemote()
	handler = newHandler()
	netService = newNetService()
	route = make(map[string]func(*Session) string)
	TimerManager = newTimer()
	channelServive = newChannelServive()
	ConnectionService = newConnectionService()
	protocolState = PROTOCOL_START
	heartbeatService = newHeartbeatService()
	endRunning = make(chan bool, 1)

	workPath, _ = os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath = filepath.Join(AppPath, "conf", "app.json")
	serverConfigPath = filepath.Join(AppPath, "conf", "servers.json")
	masterConfigPath = filepath.Join(AppPath, "conf", "master.json")
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "conf", "app.json")
		}

		if utils.FileExists(serverConfigPath) {
			os.Chdir(AppPath)
		} else {
			serverConfigPath = filepath.Join(workPath, "conf", "servers.json")
		}

		if utils.FileExists(masterConfigPath) {
			os.Chdir(AppPath)
		} else {
			masterConfigPath = filepath.Join(workPath, "conf", "master.json")
		}
	}
}

func parseConfig() {
	// initialize master server config
	if !utils.FileExists(masterConfigPath) {
		panic(fmt.Sprintf("%s not found", masterConfigPath))
	} else {
		f, _ := os.Open(masterConfigPath)
		defer f.Close()

		reader := json.NewDecoder(f)
		var master ServerConfig
		for {
			if err := reader.Decode(&master); err == io.EOF {
				break
			} else if err != nil {
				Error(err.Error())
			}
		}

		master.Type = "master"
		master.IsMaster = true
		App.Master = &master
		registerServer(master)
	}

	// initialize servers config
	if !utils.FileExists(serverConfigPath) {
		panic(fmt.Sprintf("%s not found", serverConfigPath))
	} else {
		f, _ := os.Open(serverConfigPath)
		defer f.Close()

		reader := json.NewDecoder(f)
		var servers map[string][]ServerConfig
		for {
			if err := reader.Decode(&servers); err == io.EOF {
				break
			} else if err != nil {
				Error(err.Error())
			}
		}

		for svrType, svrs := range servers {
			for _, svr := range svrs {
				svr.Type = svrType
				registerServer(svr)
			}
		}
		dumpSvrTypeMaps()
	}

	if App.Master == nil {
		panic(fmt.Sprintf("wrong master server config file(%s)", masterConfigPath))
	}

	defaultServerId := "master-server-1"
	var serverId string
	flag.StringVar(&serverId, "s", defaultServerId, "server id")
	flag.Parse()
	if serverId == defaultServerId { // master server
		App.Config = App.Master
	} else { // other server
		App.Config = svrIdMaps[serverId]
		if App.Config == nil {
			panic(fmt.Sprintf("%s infomation not found in %s", serverId, serverConfigPath))
		}
	}
}
