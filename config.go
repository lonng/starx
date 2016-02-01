package starx

import (
	"encoding/json"
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
	App               *starxApp // starx application
	AppPath           string
	workPath          string
	appConfigPath     string
	serverConfigPath  string
	masterConfigPath  string
	cluster           *clusterService                                    // cluster service
	settings          map[string][]func()                                // all settings
	remote            *remoteService                                     // remote service
	handler           *handlerService                                    // handler service
	defaultNetService *netService                                        // net service
	TimerManager      Timer                                              // timer component
	route             map[string]func(*Session) string                   // server route function
	channelServive    *ChannelServive                                    // channel service component
	ConnectionService *connectionService                                 // connection service component
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

func init() {
	App = newApp()
	cluster = newClusterService()
	settings = make(map[string][]func())
	Log = log.New(os.Stdout, "", log.LstdFlags)
	remote = newRemote()
	handler = newHandler()
	defaultNetService = newNetService()
	route = make(map[string]func(*Session) string)
	TimerManager = newTimer()
	channelServive = newChannelServive()
	ConnectionService = newConnectionService()
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
	// initialize app config
	if !utils.FileExists(appConfigPath) {
		panic(fmt.Sprintf("%s not found", appConfigPath))
	} else {
		type appConfig struct {
			AppName    string `json:"AppName"`
			Standalone bool   `json:"Standalone"`
		}
		f, _ := os.Open(appConfigPath)
		defer f.Close()
		reader := json.NewDecoder(f)
		var cfg appConfig
		for {
			if err := reader.Decode(&cfg); err == io.EOF {
				break
			} else if err != nil {
				Error(err.Error())
			}
		}
		App.AppName = cfg.AppName
		App.Standalone = cfg.Standalone
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
				cluster.registerServer(svr)
			}
		}
		cluster.dumpSvrTypeMaps()
	}

	if App.Standalone {
		if len(os.Args) < 2 {
			panic("server running in standalone mode, but not found server id argument")
		}
		serverId := os.Args[1]
		App.Config = cluster.svrIdMaps[serverId]
		if App.Config == nil {
			panic(fmt.Sprintf("%s infomation not found in %s", serverId, serverConfigPath))
		}
	} else {
		// if server running in cluster mode, master server config require
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
			cluster.registerServer(master)
		}
		if App.Master == nil {
			panic(fmt.Sprintf("wrong master server config file(%s)", masterConfigPath))
		}
		if len(os.Args) == 1 { // not pass server id, running in master mode
			App.Config = App.Master
		} else { // other server
			serverId := os.Args[1]
			App.Config = cluster.svrIdMaps[serverId]
			if App.Config == nil {
				panic(fmt.Sprintf("%s infomation not found in %s", serverId, serverConfigPath))
			}
		}
	}
}
