package starx

import (
	"encoding/json"
	"fmt"
	"github.com/chrislonng/starx/utils"
	"io"
	"log"
	"os"
	"path/filepath"
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
	ChannelServive    *channelServive                                    // channel service component
	connections       *connectionService                                 // connection service component
	heartbeatInternal time.Duration                    = time.Second * 8 // beatheart time internal, second unit
	heartbeat         *heartbeatService                                  // beatheart service
	endRunning        chan bool                                          // wait for end application
	handlers          []Component                                        // all register handler service
	remotes           []Component                                        // all register remote process call service
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
	ChannelServive = newChannelServive()
	connections = newConnectionService()
	heartbeat = newHeartbeatService()
	endRunning = make(chan bool, 1)

	workPath, _ = os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath = filepath.Join(AppPath, "configs", "app.json")
	serverConfigPath = filepath.Join(AppPath, "configs", "servers.json")
	masterConfigPath = filepath.Join(AppPath, "configs", "master.json")
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "configs", "app.json")
		}

		if utils.FileExists(serverConfigPath) {
			os.Chdir(AppPath)
		} else {
			serverConfigPath = filepath.Join(workPath, "configs", "servers.json")
		}

		if utils.FileExists(masterConfigPath) {
			os.Chdir(AppPath)
		} else {
			masterConfigPath = filepath.Join(workPath, "configs", "master.json")
		}
	}
}

func parseConfig() {
	// initialize app config
	if !utils.FileExists(appConfigPath) {
		Info("%s not found", appConfigPath)
		os.Exit(-1)
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
		Info("%s not found", serverConfigPath)
		os.Exit(-1)
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
			Info("server running in standalone mode, but not found server id argument")
			os.Exit(-1)
		}
		serverId := os.Args[1]
		App.Config = cluster.svrIdMaps[serverId]
		if App.Config == nil {
			Info("%s infomation not found in %s", serverId, serverConfigPath)
			os.Exit(-1)
		}
	} else {
		// if server running in cluster mode, master server config require
		// initialize master server config
		if !utils.FileExists(masterConfigPath) {
			Info("%s not found", masterConfigPath)
			os.Exit(-1)
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
			Info("wrong master server config file(%s)", masterConfigPath)
			os.Exit(-1)
		}
		if len(os.Args) == 1 {
			// not pass server id, running in master mode
			App.Config = App.Master
		} else {
			// other server
			serverId := os.Args[1]
			App.Config = cluster.svrIdMaps[serverId]
			if App.Config == nil {
				Info("%s infomation not found in %s", serverId, serverConfigPath)
				os.Exit(-1)
			}
		}
	}
}
