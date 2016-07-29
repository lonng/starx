package cluster

import "fmt"

type ServerConfig struct {
	Type        string
	Id          string
	Host        string
	Port        int32
	IsFrontend  bool
	IsMaster    bool
	IsWebsocket bool
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
