package starx

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/network"
	"strings"
)

var VERSION = "0.0.1"

var (
	App              *starxApp // starx application
	AppPath          string
	workPath         string
	appConfigPath    string
	serverConfigPath string
	masterConfigPath string
	settings         map[string][]func() // all settings
	endRunning       chan bool           // wait for end application
)

func init() {
	App = newApp()
	settings = make(map[string][]func())
	endRunning = make(chan bool, 1)

	workPath, _ = os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath = filepath.Join(AppPath, "configs", "app.json")
	serverConfigPath = filepath.Join(AppPath, "configs", "servers.json")
	masterConfigPath = filepath.Join(AppPath, "configs", "master.json")
	if workPath != AppPath {
		if fileExist(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "configs", "app.json")
		}

		if fileExist(serverConfigPath) {
			os.Chdir(AppPath)
		} else {
			serverConfigPath = filepath.Join(workPath, "configs", "servers.json")
		}

		if fileExist(masterConfigPath) {
			os.Chdir(AppPath)
		} else {
			masterConfigPath = filepath.Join(workPath, "configs", "master.json")
		}
	}
}

func parseConfig() {
	// initialize app config
	if !fileExist(appConfigPath) {
		log.Fatal("%s not found", appConfigPath)
		os.Exit(-1)
	} else {
		type appConfig struct {
			AppName    string `json:"AppName"`
			Standalone bool   `json:"Standalone"`
			LogLevel   string `json:"LogLevel"`
		}
		f, _ := os.Open(appConfigPath)
		defer f.Close()
		reader := json.NewDecoder(f)
		var cfg appConfig
		for {
			if err := reader.Decode(&cfg); err == io.EOF {
				break
			} else if err != nil {
				log.Error(err.Error())
			}
		}
		App.AppName = cfg.AppName
		App.Standalone = cfg.Standalone
		log.SetLevelByName(cfg.LogLevel)
	}

	// initialize servers config
	if !fileExist(serverConfigPath) {
		log.Fatal("%s not found", serverConfigPath)
		os.Exit(-1)
	} else {
		f, _ := os.Open(serverConfigPath)
		defer f.Close()

		reader := json.NewDecoder(f)
		var servers map[string][]*cluster.ServerConfig
		for {
			if err := reader.Decode(&servers); err == io.EOF {
				break
			} else if err != nil {
				log.Error(err.Error())
			}
		}

		for svrType, svrs := range servers {
			for _, svr := range svrs {
				svr.Type = svrType
				cluster.Register(svr)
			}
		}
		cluster.DumpSvrTypeMaps()
	}

	var serverId string
	flag.StringVar(&serverId, "server-id", "", "starx server id")
	flag.Parse()

	if App.Standalone {
		if strings.TrimSpace(serverId) == "" {
			log.Fatal("server running in standalone mode, but not found server id argument")
			os.Exit(-1)
		}

		cfg, err := cluster.Server(serverId)
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(-1)
		}

		App.Config = cfg
	} else {
		// if server running in cluster mode, master server config require
		// initialize master server config
		if !fileExist(masterConfigPath) {
			log.Fatal("%s not found", masterConfigPath)
			os.Exit(-1)
		} else {
			f, _ := os.Open(masterConfigPath)
			defer f.Close()

			reader := json.NewDecoder(f)
			var master *cluster.ServerConfig
			for {
				if err := reader.Decode(master); err == io.EOF {
					break
				} else if err != nil {
					log.Error(err.Error())
				}
			}

			master.Type = "master"
			master.IsMaster = true
			App.Master = master
			cluster.Register(master)
		}
		if App.Master == nil {
			log.Fatal("wrong master server config file(%s)", masterConfigPath)
			os.Exit(-1)
		}

		if strings.TrimSpace(serverId) == "" {
			// not pass server id, running in master mode
			App.Config = App.Master
		} else {
			cfg, err := cluster.Server(serverId)
			if err != nil {
				log.Fatal(err.Error())
				os.Exit(-1)
			}

			App.Config = cfg
		}
	}

	// dependencies initialization
	network.SetAppConfig(App.Config)
	cluster.SetAppConfig(App.Config)
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
