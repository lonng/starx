package starx

import (
	"reflect"
	"testing"
)

func BenchmarkPointerReflectNewValue(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(&T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t.Elem())
	}

	b.ReportAllocs()
}

func BenchmarkPointerReflectNewInterface(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(&T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t.Elem()).Interface()
	}

	b.ReportAllocs()
}
func BenchmarkReflectNewValue(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t)
	}

	b.ReportAllocs()
}

func BenchmarkReflectNewInterface(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t).Interface()
	}

	b.ReportAllocs()
}
