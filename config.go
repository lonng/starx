package starx

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
)

var VERSION = "0.0.1"

var (
	App              *starxApp // starx application
	appPath          string
	workPath         string
	AppConfigPath    string
	ServerConfigPath string
	MasterConfigPath string
	serverID         string              // current process server id
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
	appPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	AppConfigPath = filepath.Join(appPath, "configs", "app.json")
	ServerConfigPath = filepath.Join(appPath, "configs", "servers.json")
	MasterConfigPath = filepath.Join(appPath, "configs", "master.json")
	if workPath != appPath {
		if fileExist(AppConfigPath) {
			os.Chdir(appPath)
		} else {
			AppConfigPath = filepath.Join(workPath, "configs", "app.json")
		}

		if fileExist(ServerConfigPath) {
			os.Chdir(appPath)
		} else {
			ServerConfigPath = filepath.Join(workPath, "configs", "servers.json")
		}

		if fileExist(MasterConfigPath) {
			os.Chdir(appPath)
		} else {
			MasterConfigPath = filepath.Join(workPath, "configs", "master.json")
		}
	}
}

func parseConfig() {
	// initialize app config
	if !fileExist(AppConfigPath) {
		log.Fatalf("%s not found", AppConfigPath)
	} else {
		type appConfig struct {
			AppName    string `json:"AppName"`
			Standalone bool   `json:"Standalone"`
			LogLevel   string `json:"LogLevel"`
		}
		f, _ := os.Open(AppConfigPath)
		defer f.Close()
		reader := json.NewDecoder(f)
		var cfg appConfig
		for {
			if err := reader.Decode(&cfg); err == io.EOF {
				break
			} else if err != nil {
				log.Errorf(err.Error())
			}
		}
		App.AppName = cfg.AppName
		App.Standalone = cfg.Standalone
		//log.SetLevelByName(cfg.LogLevel)
	}

	// initialize servers config
	if !fileExist(ServerConfigPath) {
		log.Fatalf("%s not found", ServerConfigPath)
	} else {
		f, _ := os.Open(ServerConfigPath)
		defer f.Close()

		reader := json.NewDecoder(f)
		var servers map[string][]*cluster.ServerConfig
		for {
			if err := reader.Decode(&servers); err == io.EOF {
				break
			} else if err != nil {
				log.Errorf(err.Error())
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
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
