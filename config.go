package mello

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mello/utils"
	"os"
	"path/filepath"
	"time"
)

var VERSION = "0.0.1"

var (
	App              *MelloApp // mello application
	AppName          string
	AppPath          string
	workPath         string
	AppConfigPath    string
	ServerConfigPath string
	MasterConfigPath string
	StartTime        time.Time
	SvrConfigs       []*ServerConfig            // all servers config
	SvrTypes         []string                   // all server type
	SvrTypeMaps      map[string][]*ServerConfig // all servers type maps
	SvrIdMaps        map[string]*ServerConfig   // all servers id maps
	Settings         map[string][]func()
	Rpc              *MelloRpc                // rpc proxy
	Handler          *MelloHandler            // hander
	TimerManager     Timer                    // timer component
	Route            map[string]func() string // server route function
	SessionService   *MelloSessionService     // session service component
	ChannelServive   *MelloChannelServive     // channel service component
)

type ServerConfig struct {
	Type       string
	Id         string
	Host       string
	Port       int32
	ClientPort int32
	IsFrontend bool
	IsMaster   bool
}

func (this *ServerConfig) String() string {
	return fmt.Sprintf("Type: %s, Id: %s, Host: %s, Port: %d, ClientPort: %d, IsFrontend: %t, IsMaster: %t",
		this.Type,
		this.Id,
		this.Host,
		this.Port,
		this.ClientPort,
		this.IsFrontend,
		this.IsMaster)
}

func dumpSvrConfigs() {
	for _, v := range SvrConfigs {
		Info(v.String())
	}
}

func dumpSvrIdMaps() {
	for _, v := range SvrIdMaps {
		Info(fmt.Sprintf("id: %s(%s)", v.Id, v.String()))
	}
}

func dumpSvrTypeMaps() {
	for _, t := range SvrTypes {
		svrs := SvrTypeMaps[t]
		if len(svrs) == 0 {
			continue
		}
		for _, svr := range svrs {
			Info(fmt.Sprintf("%s(%s)", svr.Type, svr.String()))
		}
	}
}

func addServer(server ServerConfig) {
	// server exists
	if _, ok := SvrIdMaps[server.Id]; ok {
		Info(fmt.Sprintf("serverId: %s already existed(%s)", server.Id, server.String()))
		return
	}
	svr := &server
	SvrConfigs = append(SvrConfigs, svr)
	if len(SvrTypes) > 0 {
		for k, t := range SvrTypes {
			// duplicate
			if t == svr.Type {
				break
			}
			// arrive slice end
			if k == len(SvrTypes)-1 {
				SvrTypes = append(SvrTypes, svr.Type)
			}
		}
	} else {
		SvrTypes = append(SvrTypes, svr.Type)
	}
	SvrIdMaps[svr.Id] = svr
	SvrTypeMaps[svr.Type] = append(SvrTypeMaps[svr.Type], svr)
}

func removeServer(svrId string) {

	if _, ok := SvrIdMaps[svrId]; ok {
		// remove from Servers array
		var tempServers []*ServerConfig
		for idx, svr := range SvrConfigs {
			if svr.Id == svrId {
				tempServers = append(tempServers, SvrConfigs[:idx]...)
				tempServers = append(tempServers, SvrConfigs[(idx+1):]...)
				break
			}
		}
		SvrConfigs = tempServers
		// remove from ServerIdMaps map
		typ := SvrIdMaps[svrId].Type
		if svrs, ok := SvrTypeMaps[typ]; ok && len(svrs) > 0 {
			if len(svrs) == 1 { // array only one element, remove it directly
				delete(SvrTypeMaps, typ)
			} else {
				var tempSvrs []*ServerConfig
				for idx, svr := range svrs {
					if svr.Id == svrId {
						tempSvrs = append(tempSvrs, svrs[:idx]...)
						tempSvrs = append(tempSvrs, svrs[(idx+1):]...)
						break
					}
				}
				SvrTypeMaps[typ] = tempSvrs
			}
		}
		// remove from ServerIdMaps
		delete(SvrIdMaps, svrId)
		Rpc.CloseClient(svrId)
	} else {
		Info(fmt.Sprintf("serverId: %s not found", svrId))
	}
}

func init() {
	App = NewApp()
	SvrTypeMaps = make(map[string][]*ServerConfig)
	SvrIdMaps = make(map[string]*ServerConfig)
	Settings = make(map[string][]func())
	Log = log.New(os.Stdout, "", log.LstdFlags)
	Rpc = NewRpc()
	Handler = NewHandler()
	Route = make(map[string]func() string)
	SessionService = NewSesseionService()
	ChannelServive = NewChannelServive()

	workPath, _ = os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	AppConfigPath = filepath.Join(AppPath, "conf", "app.json")
	ServerConfigPath = filepath.Join(AppPath, "conf", "servers.json")
	MasterConfigPath = filepath.Join(AppPath, "conf", "master.json")
	if workPath != AppPath {
		if utils.FileExists(AppConfigPath) {
			os.Chdir(AppPath)
		} else {
			AppConfigPath = filepath.Join(workPath, "conf", "app.json")
		}

		if utils.FileExists(ServerConfigPath) {
			os.Chdir(AppPath)
		} else {
			ServerConfigPath = filepath.Join(workPath, "conf", "servers.json")
		}
		
		if utils.FileExists(MasterConfigPath) {
			os.Chdir(AppPath)
		} else {
			MasterConfigPath = filepath.Join(workPath, "conf", "master.json")
		}
	}
}

func ParseConfig() {
	// initialize master server config
	if !utils.FileExists(MasterConfigPath) {
		panic(fmt.Sprintf("%s not found", MasterConfigPath))
	} else {
		f, _ := os.Open(MasterConfigPath)
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
		addServer(master)
	}

	// initialize servers config
	if !utils.FileExists(ServerConfigPath) {
		panic(fmt.Sprintf("%s not found", ServerConfigPath))
	} else {
		f, _ := os.Open(ServerConfigPath)
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
				addServer(svr)
			}
		}
	}

	if App.Master == nil {
		panic(fmt.Sprintf("wrong master server config file(%s)", MasterConfigPath))
	}

	defaultServerId := "master-server-1"
	var serverId string
	flag.StringVar(&serverId, "s", defaultServerId, "server id")
	flag.Parse()
	if serverId == defaultServerId { // master server
		App.CurSvrConfig = App.Master
	} else { // other server
		App.CurSvrConfig = SvrIdMaps[serverId]
		if App.CurSvrConfig == nil {
			panic(fmt.Sprintf("%s infomation not found in %s", serverId, ServerConfigPath))
		}
	}
}
