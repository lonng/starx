// Copyright (c) starx Author. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package starx

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lonnng/starx/cluster"
	"github.com/lonnng/starx/log"
	"github.com/lonnng/starx/timer"
)

var VERSION = "0.0.1"

var (
	// app represents the current server process
	app = &struct {
		master     *cluster.ServerConfig // master server config
		config     *cluster.ServerConfig // current server information
		name       string                // current application name
		standalone bool                  // current server is running in standalone mode
		startAt    time.Time             // startup time
	}{}

	// env represents the environment of the current process, includes
	// work path and config path etc.
	env = &struct {
		wd                string                      // working path
		serversConfigPath string                      // servers config path(default: $appPath/configs/servers.json)
		masterServerId    string                      // master server id
		serverId          string                      // current process server id
		settings          map[string][]ServerInitFunc // all settings
		heartbeatInternal time.Duration               // heartbeat internal
		die               chan bool                   // wait for end application

		checkOrigin func(*http.Request) bool // check origin when websocket enabled
	}{}
)

type ServerInitFunc func()

// init default configs
func init() {
	// register session manager for cluster
	cluster.SetSessionManager(transporter)

	// application initialize
	app.name = strings.TrimLeft(path.Base(os.Args[0]), "/")
	app.standalone = true
	app.startAt = time.Now()

	// environment initialize
	env.settings = make(map[string][]ServerInitFunc)
	env.die = make(chan bool)

	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		env.wd, _ = filepath.Abs(wd)

		// config file path
		serversConfigPath := filepath.Join(wd, "configs", "servers.json")

		if fileExists(serversConfigPath) {
			env.serversConfigPath = serversConfigPath
		}
	}
}

func loadServers() {
	// initialize servers config
	if !fileExists(env.serversConfigPath) {
		log.Fatalf("%s not found", env.serversConfigPath)
		return
	}

	// read config file
	f, _ := os.Open(env.serversConfigPath)
	defer f.Close()

	// load config from file
	reader := json.NewDecoder(f)
	var servers map[string][]*cluster.ServerConfig
	for {
		if err := reader.Decode(&servers); err == io.EOF {
			break
		} else if err != nil {
			log.Errorf(err.Error())
		}
	}

	// register server to cluster
	for typ, svrs := range servers {
		for _, svr := range svrs {
			svr.Type = typ
			cluster.Register(svr)
		}
	}
	cluster.DumpServers()
}

func initSetting() {
	// init
	if app.standalone {
		if strings.TrimSpace(env.serverId) == "" {
			log.Fatal("server running in standalone mode, but not found server id argument")
		}

		cfg, err := cluster.Server(env.serverId)
		if err != nil {
			log.Fatal(err.Error())
		}

		app.config = cfg
	} else {
		// if server running in cluster mode, master server config require
		// initialize master server config
		if env.masterServerId == "" {
			log.Fatalf("master server id must be set in cluster mode", env.masterServerId)
		}

		if server, err := cluster.Server(env.masterServerId); err != nil {
			log.Fatalf("wrong master server config file(%s)", env.masterServerId)
		} else {
			app.master = server
		}

		if strings.TrimSpace(env.serverId) == "" {
			// not pass server id, running in master mode
			app.config = app.master
		} else {
			cfg, err := cluster.Server(env.serverId)
			if err != nil {
				log.Fatal(err.Error())
			}

			app.config = cfg
		}
	}

	// dependencies initialization
	cluster.SetAppConfig(app.config)
}

func initServer() {
	setting, ok := env.settings[app.config.Type]
	if !ok {
		return
	}

	// call all init function
	for _, fn := range setting {
		fn()
	}

	// register heartbeat service
	if app.config.IsFrontend {
		timer.Register(env.heartbeatInternal, func() {
			transporter.heartbeat()
		})
	}
}
