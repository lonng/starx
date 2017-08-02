package main

import (
	"net/http"

	"github.com/lonnng/starx"
	"github.com/lonnng/starx/component"
	"github.com/lonnng/starx/log"
	"github.com/lonnng/starx/serialize/json"
	"github.com/lonnng/starx/session"
)

type Room struct {
	component.Base
	group *starx.Group
}

type UserMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type JoinResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

func NewRoom() *Room {
	return &Room{
		group: starx.NewGroup("room"),
	}
}

func (r *Room) Join(s *session.Session, msg []byte) error {
	s.Bind(s.ID)   // binding session uid
	r.group.Add(s) // add session to group
	return s.Response(JoinResponse{Result: "sucess"})
}

func (r *Room) Message(s *session.Session, msg *UserMessage) error {
	return r.group.Broadcast("onMessage", msg)
}

func main() {
	starx.SetServersConfig("configs/servers.json")
	starx.Register(NewRoom())

	starx.SetServerID("demo-server-1")
	starx.SetSerializer(json.NewSerializer())

	log.SetLevel(log.LevelDebug)

	starx.SetCheckOriginFunc(func(_ *http.Request) bool { return true })
	starx.Run()
}
