package cluster

import "fmt"

type ServerConfig struct {
	Type        string `json:"type"`
	Id          string `json:"id"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	IsFrontend  bool   `json:"is_frontend"`
	IsMaster    bool   `json:"is_master"`
	IsWebsocket bool   `json:"is_websocket"`
}

func (c *ServerConfig) String() string {
	return fmt.Sprintf("Type: %s, Id: %s, Host: %s, Port: %d, IsFrontend: %t, IsMaster: %t, IsWebsocket: %t",
		c.Type,
		c.Id,
		c.Host,
		c.Port,
		c.IsFrontend,
		c.IsMaster,
		c.IsWebsocket)
}
