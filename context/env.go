package context

import (
	"net"
	"reflect"
	"sync"
	"time"
)

// Env should used as a map, but it has an inherited mode, overlay liked.
// If you set, the value should be store at local.
// If you get, the value should be get from local first, otherwise from parent
type Env interface {
	// Fork return an inherited sub Env, and I am it's parent
	Fork() Env

	// Set always set key & value at local storage
	Set(key, value interface{})
	// Get always check local, if the key not exists, then check parent
	Get(key interface{}) (value interface{}, ok bool)
	Has(key interface{}) (ok bool)
	Keys() []interface{}

	GetInt(key interface{}) int
	GetInt64(key interface{}) int64
	GetUint(key interface{}) uint
	GetUint64(key interface{}) uint64
	GetBool(key interface{}) bool
	GetFloat(key interface{}) float64
	GetString(key interface{}) string
	GetIP(key interface{}) net.IP
	GetAddr(key interface{}) net.Addr
	GetTime(key interface{}) time.Time
	GetDuration(key interface{}) time.Duration
}

type env struct {
	parent *env

	vals sync.Map
}

var _ Env = &env{}

// NewEnv return a simple Env, use sync.Map as it's storage
func NewEnv() Env {
	return &env{}
}

func (e *env) fork() *env {
	return &env{
		parent: e,
	}
}

func (e *env) Fork() Env {
	return e.fork()
}

func (e *env) Set(key, value interface{}) {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	e.vals.Store(key, value)
}

func (e *env) Get(key interface{}) (value interface{}, ok bool) {
	// from local
	if value, ok := e.vals.Load(key); ok {
		return value, ok
	}
	// otherwise from parent
	if e.parent != nil {
		return e.parent.Get(key)
	}
	return nil, false
}

func (e *env) Has(key interface{}) (ok bool) {
	_, ok = e.Get(key)
	return
}

func (e *env) keys() map[interface{}]struct{} {
	var keys map[interface{}]struct{}
	if e.parent != nil {
		keys = e.parent.keys()
	} else {
		keys = make(map[interface{}]struct{})
	}
	e.vals.Range(func(k, v interface{}) bool {
		keys[k] = struct{}{}
		return true
	})
	return keys
}

func (e *env) Keys() []interface{} {
	var (
		idx  = e.keys()
		keys = make([]interface{}, 0, len(idx))
	)
	for k := range idx {
		keys = append(keys, k)
	}
	return keys
}

func (e *env) GetInt(key interface{}) int {
	if value, ok := e.Get(key); ok {
		return value.(int)
	}
	return 0
}

func (e *env) GetInt64(key interface{}) int64 {
	if value, ok := e.Get(key); ok {
		return value.(int64)
	}
	return 0
}

func (e *env) GetUint(key interface{}) uint {
	if value, ok := e.Get(key); ok {
		return value.(uint)
	}
	return 0
}

func (e *env) GetUint64(key interface{}) uint64 {
	if value, ok := e.Get(key); ok {
		return value.(uint64)
	}
	return 0
}

func (e *env) GetBool(key interface{}) bool {
	if value, ok := e.Get(key); ok {
		return value.(bool)
	}
	return false
}

func (e *env) GetFloat(key interface{}) float64 {
	if value, ok := e.Get(key); ok {
		return value.(float64)
	}
	return 0.0
}

func (e *env) GetString(key interface{}) string {
	if value, ok := e.Get(key); ok {
		return value.(string)
	}
	return ""
}

func (e *env) GetIP(key interface{}) net.IP {
	if value, ok := e.Get(key); ok {
		return value.(net.IP)
	}
	return nil
}

func (e *env) GetAddr(key interface{}) net.Addr {
	if value, ok := e.Get(key); ok {
		return value.(net.Addr)
	}
	return nil
}

func (e *env) GetTime(key interface{}) time.Time {
	if value, ok := e.Get(key); ok {
		return value.(time.Time)
	}
	return time.Time{}
}

func (e *env) GetDuration(key interface{}) time.Duration {
	if value, ok := e.Get(key); ok {
		return value.(time.Duration)
	}
	return 0
}
