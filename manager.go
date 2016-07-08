package starx

import (
	"encoding/json"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/session"
)

type Manager struct {
	Name    string
	Counter int
}

// Component interface methods
func (this *Manager) Init() {
	this.Name = "ManagerComponenet"
	log.Info("manager component initialized")
}
func (this *Manager) AfterInit()      {}
func (this *Manager) BeforeShutdown() {}
func (this *Manager) Shutdown()       {}

// attachment methods
func (m *Manager) UpdateServer(session *session.Session, data []byte) error {
	var newServerInfo *cluster.ServerConfig
	err := json.Unmarshal(data, newServerInfo)
	if err != nil {
		return err
	}
	cluster.UpdateServer(newServerInfo)
	return nil
}

func (m *Manager) RegisterServer(session *session.Session, data []byte) error {
	var newServerInfo *cluster.ServerConfig
	err := json.Unmarshal(data, newServerInfo)
	if err != nil {
		return err
	}
	log.Info("new server connected in")
	cluster.Register(newServerInfo)
	return nil
}

func (m *Manager) RemoveServer(session *session.Session, data []byte) error {
	var srvId string
	err := json.Unmarshal(data, &srvId)
	if err != nil {
		return err
	}
	cluster.RemoveServer(srvId)
	return nil
}
