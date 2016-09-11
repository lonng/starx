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
	appConfigPath string
	serversConfigPath string
	masterConfigPath string
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

	appConfigPath = filepath.Join(appPath, "configs", "app.json")
	serversConfigPath = filepath.Join(appPath, "configs", "servers.json")
	masterConfigPath = filepath.Join(appPath, "configs", "master.json")
	if workPath != appPath {
		if fileExist(appConfigPath) {
			os.Chdir(appPath)
		} else {
			appConfigPath = filepath.Join(workPath, "configs", "app.json")
		}

		if fileExist(serversConfigPath) {
			os.Chdir(appPath)
		} else {
			serversConfigPath = filepath.Join(workPath, "configs", "servers.json")
		}

		if fileExist(masterConfigPath) {
			os.Chdir(appPath)
		} else {
			masterConfigPath = filepath.Join(workPath, "configs", "master.json")
		}
	}
}

func parseConfig() {
	// initialize app config
	if !fileExist(appConfigPath) {
		log.Fatalf("%s not found", appConfigPath)
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
				log.Errorf(err.Error())
			}
		}
		App.AppName = cfg.AppName
		App.Standalone = cfg.Standalone
		//log.SetLevelByName(cfg.LogLevel)
	}

	// initialize servers config
	if !fileExist(serversConfigPath) {
		log.Fatalf("%s not found", serversConfigPath)
	} else {
		f, _ := os.Open(serversConfigPath)
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
