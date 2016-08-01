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
		log.Fatalf("%s not found", appConfigPath)
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
				log.Errorf(err.Error())
			}
		}
		App.AppName = cfg.AppName
		App.Standalone = cfg.Standalone
		log.SetLevelByName(cfg.LogLevel)
	}

	// initialize servers config
	if !fileExist(serverConfigPath) {
		log.Fatalf("%s not found", serverConfigPath)
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
