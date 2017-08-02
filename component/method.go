package component

import (
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/lonnng/starx/session"
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfBytes   = reflect.TypeOf(([]byte)(nil))
	typeOfSession = reflect.TypeOf(session.New(nil))
)

func isExported(name string) bool {
	w, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(w)
}

func isExportedOrBuiltinType(t reflect.Type) bool {
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
	mt := method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}

	// Method needs three ins: receiver, *Session, []byte or pointer.
	if mt.NumIn() != 3 {
		return false
	}

	// Method needs one outs: error
	if mt.NumOut() != 1 {
		return false
	}

	if t1 := mt.In(1); t1.Kind() != reflect.Ptr || t1 != typeOfSession {
		return false
	}

	if (mt.In(2).Kind() != reflect.Ptr && mt.In(2) != typeOfBytes) || mt.Out(0) != typeOfError {
		return false
	}
	return true
}

// IsRemoteMethod
// decide a method is suitable remote method
func isRemoteMethod(method reflect.Method) bool {
	mt := method.Type

	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}

	// Method needs one outs: []byte, error
	if mt.NumOut() != 2 {
		return false
	}

	if mt.Out(0).Kind() != reflect.Interface || mt.Out(1) != typeOfError {
		return false
	}

	return true
}

// suitableMethods returns suitable methods of typ, it will report
// error using log if reportErr is true.
func suitableHandlerMethods(typ reflect.Type, reportErr bool) map[string]*HandlerMethod {
	methods := make(map[string]*HandlerMethod)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mt := method.Type
		mn := method.Name
		if isHandlerMethod(method) {
			raw := false
			if mt.In(2) == typeOfBytes {
				raw = true
			}
			methods[mn] = &HandlerMethod{Method: method, Type: mt.In(2), Raw: raw}
		}
	}
	return methods
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableRemoteMethods(typ reflect.Type, reportErr bool) map[string]*RemoteMethod {
	methods := make(map[string]*RemoteMethod)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mn := method.Name
		if isRemoteMethod(method) {
			methods[mn] = &RemoteMethod{Method: method}
		}
	}
	return methods
}
