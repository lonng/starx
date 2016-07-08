package network

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfBytes = reflect.TypeOf(([]byte)(nil))

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func IsExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// IsHandlerMethod
// decide a method is suitable handler method
func isHandlerMethod(method reflect.Method) bool {
	mtype := method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}

	// Method needs three ins: receiver, *Session, []byte or pointer.
	if mtype.NumIn() != 3 {
		return false
	}
	// Method needs one outs: error
	if mtype.NumOut() != 1 {
		return false
	}

	if arg1Type := mtype.In(1); arg1Type.Kind() != reflect.Ptr || arg1Type.Elem().Name() != "Session" {
		return false
	}

	if mtype.In(2).Kind() != reflect.Ptr || mtype.Out(0) != typeOfError {
		return false
	}
	return true
}

// IsRemoteMethod
// decide a method is suitable remote methd
func isRemoteMethod(method reflect.Method) bool {
	mtype := method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}
	// Method needs one outs: []byte, error
	if mtype.NumOut() != 2 {
		return false
	}

	if mtype.Out(0) != typeOfBytes || mtype.Out(1) != typeOfError {
		return false
	}

	return true
}

// suitableMethods returns suitable methods of typ, it will report
// error using log if reportErr is true.
func suitableHandlerMethods(typ reflect.Type, reportErr bool) map[string]*handlerMethod {
	methods := make(map[string]*handlerMethod)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if isHandlerMethod(method) {
			methods[mname] = &handlerMethod{method: method, dataType: mtype.In(2)}
		}
	}
	return methods
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableRemoteMethods(typ reflect.Type, reportErr bool) map[string]*remoteMethod {
	methods := make(map[string]*remoteMethod)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mname := method.Name
		if isRemoteMethod(method) {
			methods[mname] = &remoteMethod{method: method}
		}
	}
	return methods
}
