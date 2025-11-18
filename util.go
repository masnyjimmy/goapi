package goapi

import "reflect"

func derefType(Type reflect.Type) reflect.Type {
	for Type.Kind() == reflect.Pointer {
		Type = Type.Elem()
	}

	return Type
}
