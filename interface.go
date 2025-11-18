package goapi

import (
	"reflect"
	"sync"
)

type Spec struct {
	Name        string
	Required    bool
	Description string
}

type paramType interface {
	Spec() Spec
}

type paramWithFormat interface {
	Format() string
}

type paramWithIn interface {
	In() ParamIn
}

func getInterface[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

var interfaceCache = struct {
	sync.RWMutex
	cache map[reflect.Type]any
}{
	cache: make(map[reflect.Type]any),
}

func getCachedInterface(t reflect.Type) any {
	interfaceCache.RLock()
	if value, has := interfaceCache.cache[t]; has {
		interfaceCache.RUnlock()
		return value
	}
	interfaceCache.RUnlock()

	// Create instance WITHOUT holding any locks
	var instance any
	if t.Kind() == reflect.Pointer {
		instance = reflect.New(t.Elem()).Interface()
	} else {
		instance = reflect.New(t).Elem().Interface()
	}

	// Now acquire write lock and store (check again in case another goroutine added it)
	interfaceCache.Lock()
	if existing, has := interfaceCache.cache[t]; has {
		interfaceCache.Unlock()
		return existing
	}
	interfaceCache.cache[t] = instance
	interfaceCache.Unlock()

	return instance
}

func resolveInterfaceInstance[T any](t reflect.Type) (any, bool) {
	iface := getInterface[T]()

	if t.Implements(iface) {
		return getCachedInterface(t), true
	}

	if t.Kind() != reflect.Pointer {
		ptr := reflect.PointerTo(t)
		if ptr.Implements(iface) {
			return getCachedInterface(ptr), true
		}
	}
	return nil, false
}

func extractSpec(t reflect.Type) *Spec {
	if value, ok := resolveInterfaceInstance[paramType](t); ok {
		spec := value.(paramType).Spec()
		return &spec
	}
	return nil
}

func extractFormat(Type reflect.Type) string {
	if value, ok := resolveInterfaceInstance[paramWithFormat](Type); ok {
		return value.(paramWithFormat).Format()
	}
	return ""
}

func extractIn(t reflect.Type) ParamIn {
	if value, ok := resolveInterfaceInstance[paramWithIn](t); ok {
		return value.(paramWithIn).In()
	}

	return ParamUndefined
}
