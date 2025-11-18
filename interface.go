package goapi

import (
	"reflect"
	"sync"
)

type ParamSpec struct {
	Name        string
	Required    bool
	Description string
}

type ParamType interface {
	Spec() ParamSpec
}

type ParamWithFormat interface {
	Format() string
}

type ParamWithIn interface {
	In() ParamIn
}

var (
	paramTypeInterface       = reflect.TypeOf((*ParamType)(nil)).Elem()
	paramWithFormatInterface = reflect.TypeOf((*ParamWithFormat)(nil)).Elem()
)

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

func resolveInterfaceInstance(t reflect.Type, iface reflect.Type) (any, bool) {
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

func extractSpec(t reflect.Type) *ParamSpec {
	if value, ok := resolveInterfaceInstance(t, paramTypeInterface); ok {
		spec := value.(ParamType).Spec()
		return &spec
	}
	return nil
}

func extractFormat(Type reflect.Type) string {
	if value, ok := resolveInterfaceInstance(Type, paramWithFormatInterface); ok {
		return value.(ParamWithFormat).Format()
	}
	return ""
}
