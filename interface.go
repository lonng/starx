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
	"net/http"
	"strings"
	"time"

	"github.com/lonnng/starx/cluster"
	"github.com/lonnng/starx/component"
	"github.com/lonnng/starx/session"
)

// Run server
func Run() {
	// output welcome message
	// welcomeMsg()

	// load servers config from $env.serversConfigPath
	loadServers()

	// init cluster servers config
	initSetting()

	// initialize current server specified by $env.serverId
	// execute initialize function registered by application
	initServer()

	// startup current server, and loading all components that
	// registered in server initialize function
	startup()
}

// Set special server initial function, starx.Set("oneServerType | anotherServerType", func(){})
func Set(svrTypes string, fn func()) {
	var types = strings.Split(strings.TrimSpace(svrTypes), "|")
	for _, t := range types {
		t = strings.TrimSpace(t)
		env.settings[t] = append(env.settings[t], fn)
	}
}

func SetRouter(svrType string, fn func(*session.Session) string) {
	cluster.Router(svrType, fn)
}

func Register(c component.Component) {
	comps = append(comps, c)
}

func SetServerID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		panic("empty server id")
	}
	env.serverId = id
}

// Set the path of servers.json
func SetServersConfig(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		panic("empty app path")
	}
	env.serversConfigPath = path
}

// Set heartbeat time internal
func SetHeartbeatInternal(d time.Duration) {
	env.heartbeatInternal = d
}

// SetCheckOriginFunc set the function that check `Origin` in http headers
func SetCheckOriginFunc(fn func(*http.Request) bool) {
	env.checkOrigin = fn
}

// EnableCluster enable cluster mode
func EnableCluster() {
	app.standalone = false
}

// SetMasterServerID set master server id, config must be contained
// in servers.json master server id must be set when cluster mode
// enabled
func SetMasterServerID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		panic("empty master server id")
	}
	env.masterServerId = id
}

func Shutdown() {
	close(env.die)
}
