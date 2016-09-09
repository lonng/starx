package route

import "testing"

func TestDecodeRoute(t *testing.T) {
	if _, err := Decode("a.b.c"); err != nil {
		t.Error(err.Error())
	}

	if _, err := Decode("a.b.c.d"); err == nil {
		t.Fail()
	}

	if _, err := Decode("a.b."); err == nil {
		t.Fail()
	}

	if _, err := Decode(".b."); err == nil {
		t.Fail()
	}

	if _, err := Decode(".."); err == nil {
		t.Fail()
	}

	if _, err := Decode("a.b"); err != nil {
		t.Error(err.Error())
	}
}
