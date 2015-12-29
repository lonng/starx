package starx

import (
	"encoding/json"
)

type Manager struct {
	Name    string
	Counter int
}

func (this *Manager) Setup() {
	this.Name = "ManagerComponenet"
	Info("manager component initialized")
}

func (m *Manager) UpdateServer(session *Session, data []byte) error {
	var newServerInfo ServerConfig
	err := json.Unmarshal(data, &newServerInfo)
	if err != nil {
		return err
	}
	updateServer(newServerInfo)
	return nil
}

func (m *Manager) RegisterServer(session *Session, data []byte) error {
	var newServerInfo ServerConfig
	err := json.Unmarshal(data, &newServerInfo)
	if err != nil {
		return err
	}
	Info("new server connected in")
	registerServer(newServerInfo)
	return nil
}

func (m *Manager) RemoveServer(session *Session, data []byte) error {
	var srvId string
	err := json.Unmarshal(data, &srvId)
	if err != nil {
		return err
	}
	removeServer(srvId)
}
