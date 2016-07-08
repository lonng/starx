package network

import "testing"

func TestDecodeRoute(t *testing.T) {
	if _, err := DecodeRoute("a.b.c"); err != nil {
		t.Error(err.Error())
	}

	if _, err := DecodeRoute("a.b.c.d"); err == nil {
		t.Fail()
	}

	if _, err := DecodeRoute("a.b."); err == nil {
		t.Fail()
	}

	if _, err := DecodeRoute(".b."); err == nil {
		t.Fail()
	}

	if _, err := DecodeRoute(".."); err == nil {
		t.Fail()
	}

	if _, err := DecodeRoute("a.b"); err != nil {
		t.Error(err.Error())
	}
}
