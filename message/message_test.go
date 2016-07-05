package message

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	m1 := &Message{
		Type:  Request,
		ID:    100,
		Route: "test.test.test",
		Data:  []byte(`hello world`),
	}
	dm1 := Decode(m1.Encode())

	if !reflect.DeepEqual(m1, dm1) {
		t.Fail()
	}

	m2 := &Message{
		Type:       Request,
		ID:         100,
		RouteCode:  1000,
		IsCompress: true,
		Data:       []byte(`hello world`),
	}
	dm2 := Decode(m2.Encode())

	if !reflect.DeepEqual(m2, dm2) {
		t.Fail()
	}

	m3 := &Message{
		Type:  Response,
		ID:    100,
		Route: "test.test.test",
		Data:  []byte(`hello world`),
	}
	dm3 := Decode(m3.Encode())

	if !reflect.DeepEqual(m3, dm3) {
		t.Fail()
	}

	m4 := &Message{
		Type:       Response,
		ID:         100,
		RouteCode:  1000,
		IsCompress: true,
		Data:       []byte(`hello world`),
	}
	dm4 := Decode(m4.Encode())

	if !reflect.DeepEqual(m4, dm4) {
		t.Fail()
	}

	m5 := &Message{
		Type:  Notify,
		Route: "test.test.test",
		Data:  []byte(`hello world`),
	}
	dm5 := Decode(m5.Encode())

	if !reflect.DeepEqual(m5, dm5) {
		t.Fail()
	}

	m6 := &Message{
		Type:       Notify,
		RouteCode:  1000,
		IsCompress: true,
		Data:       []byte(`hello world`),
	}
	dm6 := Decode(m6.Encode())

	if !reflect.DeepEqual(m6, dm6) {
		t.Fail()
	}

	m7 := &Message{
		Type:  Push,
		Route: "test.test.test",
		Data:  []byte(`hello world`),
	}
	dm7 := Decode(m7.Encode())

	if !reflect.DeepEqual(m7, dm7) {
		t.Fail()
	}

	m8 := &Message{
		Type:       Push,
		RouteCode:  1000,
		IsCompress: true,
		Data:       []byte(`hello world`),
	}
	dm8 := Decode(m8.Encode())

	if !reflect.DeepEqual(m8, dm8) {
		t.Fail()
	}
}
