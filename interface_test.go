package starx

import "testing"

func TestSetServerID(t *testing.T) {
	SetServerID("test")
	if env.serverId != "test" {
		t.Fail()
	}
}

func TestSetServersConfig(t *testing.T) {
	SetServersConfig("testservers.json")
	if env.serversConfigPath != "testservers.json" {
		t.Fail()
	}
}
