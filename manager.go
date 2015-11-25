package starx

import (
	"fmt"
)

type Manager struct {
	Name    string
	Counter int
}

func (this *Manager) Setup() {
	this.Name = "ManagerComponenet"
	Info("manager component initialized")
}

func (this *Manager) RemoveServer(svrId *string, replay *int) error {
	App.removeChan <- *svrId
	return nil
}

func (this *Manager) AddServer(svr *ServerConfig, replay *int) error {
	Info(App.Config.String())
	App.registChan <- svr
	return nil
}

func (this *Manager) Test(g string, replay *int) error {
	this.Counter++
	Info(fmt.Sprintf("%s, %d", g, this.Counter))
	return nil
}
