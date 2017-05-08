package main

import (
	"net/http"

	"github.com/chrislonng/starx"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/serialize/json"
	"github.com/chrislonng/starx/session"
)

type Room struct {
	component.Base
	channel *starx.Channel
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
		channel: starx.ChannelService.NewChannel("room"),
	}
}

func (r *Room) Join(s *session.Session, msg []byte) error {
	s.Bind(s.ID)     // binding session uid
	r.channel.Add(s) // add session to channel
	return s.Response(JoinResponse{Result: "sucess"})
}

func (r *Room) Message(s *session.Session, msg *UserMessage) error {
	return r.channel.Broadcast("onMessage", msg)
}

func main() {
	starx.SetAppConfig("configs/app.json")
	starx.SetServersConfig("configs/servers.json")
	starx.Register(NewRoom())

	starx.SetServerID("demo-server-1")
	starx.SetSerializer(json.NewSerializer())

	log.SetLevel(log.LevelDebug)

	starx.SetCheckOriginFunc(func(_ *http.Request) bool { return true })
	starx.Run()
}
