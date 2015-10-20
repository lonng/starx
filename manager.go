package mello

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
	App.SvrRemoveChan <- *svrId
	return nil
}

func (this *Manager) AddServer(svr *ServerConfig, replay *int) error {
	Info(App.CurSvrConfig.String())
	App.SvrAddChan <- *svr
	return nil
}

func (this *Manager) Test(g string, replay *int) error {
	this.Counter++
	Info(fmt.Sprintf("%s, %d", g, this.Counter))
	return nil
}
