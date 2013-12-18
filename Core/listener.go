package Core

import (
	"reflect"
	"errors"
	"sync"
	"fmt"
)

type Listener struct {
	callers map[string]map[string]reflect.Value
	sync.Mutex
}

// get all listening event
func (this *Listener) Listening() []string {
	s := make([]string, len(this.callers))
	if len(s) < 1 {
		return s
	}
	i := 0
	for name, _ := range this.callers {
		s[i] = name
	}
	return s
}

// add event listener function
func (this *Listener) AddListener(eventName, name string, fn interface {}) error {
	this.Lock()
	defer this.Unlock()
	reflectFn := reflect.ValueOf(fn)
	// only support function
	if reflectFn.Kind() != reflect.Func {
		return errors.New("listener supports function only")
	}
	if this.callers[eventName] == nil {
		this.callers[eventName] = make(map[string]reflect.Value)
	}
	this.callers[eventName][name] = reflectFn
	return nil
}

func (this *Listener) RemoveListener(eventName string, name ...string) {
	if this.callers[eventName] == nil {
		return
	}
	this.Lock()
	defer this.Unlock()
	if len(name) < 1 {
		delete(this.callers, eventName)
		return
	}
	for _, n := range name {
		delete(this.callers[eventName], n)
	}
}

// call event listener function by event and listener name
func (this *Listener) Emit(event string, name string, args...interface {}) ([]interface {}, error) {
	// println(event+" @ "+name)
	// get function
	fn := this.callers[event][name]
	if !fn.IsValid() {
		return nil, errors.New("invalid listener name as "+event)
	}
	// check pass-in number
	inNum := fn.Type().NumIn()
	if len(args) > inNum {
		return nil, errors.New("too many listener caller arguments")
	}
	// call function
	reflectArgs := make([]reflect.Value, inNum)
	for i, _ := range reflectArgs {
		reflectArgs[i] = reflect.ValueOf(args[i])
	}
	reflectResult := fn.Call(reflectArgs)
	result := make([]interface {}, len(reflectResult))
	for i, _ := range result {
		result[i] = reflectResult[i].Interface()
	}
	if IsDev() {
		fmt.Println("[Core.Event] emit listener : "+event+"@"+name)
	}
	return result, nil
}

// run all listener in event
func (this *Listener) EmitAll(event string, args...interface {}) (map[string][]interface {}, map[string]error) {
	// println(event+" @ all")
	if len(this.callers[event]) < 1 {
		return nil, nil
	}
	e := make(map[string]error, 0)
	r := make(map[string][]interface {}, 0)
	for name, _ := range this.callers[event] {
		res, ei := this.Emit(event, name, args...)
		r[name] = res
		e[name] = ei
	}
	return r, e
}

// create new listener object
func NewListener() *Listener {
	listener := &Listener{}
	listener.callers = make(map[string]map[string]reflect.Value)
	return listener
}

