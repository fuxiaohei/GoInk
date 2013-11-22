package app

import (
	"reflect"
	"errors"
	"fmt"
)

type InkFilter struct {
	isPrint bool
	prefix  string
	callers map[string]map[string]reflect.Value
}

// add event filter function
func (this *InkFilter) Add(eventName, filterName string, fn interface {}) error {
	reflectFn := reflect.ValueOf(fn)
	// only support function
	if reflectFn.Kind() != reflect.Func {
		return errors.New("filter supports function only")
	}
	if this.callers[this.prefix + "." + eventName] == nil {
		this.callers[this.prefix + "." + eventName] = make(map[string]reflect.Value)
	}
	this.callers[this.prefix + "." + eventName][filterName] = reflectFn
	return nil
}

// call event filter function by event and filter name
func (this *InkFilter) Filter(event string, name string, args...interface {}) ([]interface {}, error) {
	if this.isPrint {
		fmt.Println(event, "@", name, args)
	}
	// get function
	fn := this.callers[this.prefix + "." + event][name]
	if !fn.IsValid() {
		return nil, errors.New("invalid filter name as " + event)
	}
	// check pass-in number
	inNum := fn.Type().NumIn()
	if len(args) > inNum {
		return nil, errors.New("too many filter caller arguments")
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
	return result, nil
}

// run all filers in event
func (this *InkFilter) FilterAll(event string, args...interface {}) []error {
	key := this.prefix + "." + event
	if this.isPrint {
		fmt.Println(event, "@", "all", args)
	}
	if len(this.callers[key]) < 1 {
		return nil
	}
	e := make([]error, 0)
	for name, _ := range this.callers[key] {
		_, ei := this.Filter(event, name, args...)
		if ei != nil {
			e = append(e, ei)
		}
	}
	return e
}

// set print or not
func (this *InkFilter) EnablePrint(print bool) {
	this.isPrint = print
}

// create new filter object
func NewFilter(prefix string, isPrint bool) *InkFilter {
	filter := &InkFilter{}
	filter.prefix = prefix
	filter.callers = make(map[string]map[string]reflect.Value)
	filter.isPrint = isPrint
	return filter
}

