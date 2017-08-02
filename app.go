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
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/lonnng/starx/log"
)

func welcomeMsg() {
	fmt.Println(asciiLogo)
}

func startup() {
	startupComps()

	go func() {
		if app.config.IsWebsocket {
			listenAndServeWS()
		} else {
			listenAndServe()
		}
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT)

	// stop server
	select {
	case <-env.die:
		log.Infof("The app will shutdown in a few seconds")
	case s := <-sg:
		log.Infof("got signal: %v", s)
	}

	log.Infof("server: " + app.config.Id + " is stopping...")

	// shutdown all components registered by application, that
	// call by reverse order against register
	shutdownComps()
}

// Enable current server accept connection
func listenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.config.Host, app.config.Port))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Infof("listen at %s:%d(%s)", app.config.Host, app.config.Port, app.config.String())

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		if app.config.IsFrontend {
			go handler.handle(conn)
		} else {
			go remote.handle(conn)
		}
	}
}

func listenAndServeWS() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     env.checkOrigin,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			return
		}

		handler.HandleWS(conn)
	})

	addr := fmt.Sprintf("%s:%d", app.config.Host, app.config.Port)
	log.Infof("listen at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err.Error())
	}
}
