package starx

import (
	"encoding/json"
)

type Manager struct {
	Name    string
	Counter int
}

// Component interface methods
func (this *Manager) Init() {
	this.Name = "ManagerComponenet"
	Info("manager component initialized")
}
func (this *Manager) AfterInit()      {}
func (this *Manager) BeforeShutdown() {}
func (this *Manager) Shutdown()       {}

// attachment methods
func (m *Manager) UpdateServer(session *Session, data []byte) error {
	var newServerInfo ServerConfig
	err := json.Unmarshal(data, &newServerInfo)
	if err != nil {
		return err
	}
	cluster.updateServer(newServerInfo)
	return nil
}

func (m *Manager) RegisterServer(session *Session, data []byte) error {
	var newServerInfo ServerConfig
	err := json.Unmarshal(data, &newServerInfo)
	if err != nil {
		return err
	}
	Info("new server connected in")
	cluster.registerServer(newServerInfo)
	return nil
}

func (m *Manager) RemoveServer(session *Session, data []byte) error {
	var srvId string
	err := json.Unmarshal(data, &srvId)
	if err != nil {
		return err
	}
	cluster.removeServer(srvId)
	return nil
}
