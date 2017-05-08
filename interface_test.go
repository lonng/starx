package starx

import "testing"

func TestSetServerID(t *testing.T) {
	SetServerID("test")
	if serverID != "test" {
		t.Fail()
	}
}

func TestSetAppConfig(t *testing.T) {
	SetAppConfig("testapp.json")
	if appConfigPath != "testapp.json" {
		t.Fail()
	}
}

func TestSetMasterConfig(t *testing.T) {
	SetMasterConfig("testmaster.json")
	if masterConfigPath != "testmaster.json" {
		t.Fail()
	}
}

func TestSetServersConfig(t *testing.T) {
	SetServersConfig("testservers.json")
	if serversConfigPath != "testservers.json" {
		t.Fail()
	}
}
