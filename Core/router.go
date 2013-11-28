package Core

import "strings"

type Router struct {
	rules map[string]func(context *Context)interface {}
}

func (this *Router) Get(pattern string, fn func(context *Context)interface {}) {
	this.rules["GET:"+pattern] = fn
}

func (this *Router) Post(pattern string, fn func(context *Context)interface {}) {
	this.rules["POST:"+pattern] = fn
}

func (this *Router) Delete(pattern string, fn func(context *Context)interface {}) {
	this.rules["DELETE:"+pattern] = fn
}

func (this *Router) Put(pattern string, fn func(context *Context)interface {}) {
	this.rules["PUT:"+pattern] = fn
}

func (this *Router) Match(pattern string) func(context *Context)interface {} {
	maxLength, matchPattern := 0, ""
	for p, _ := range this.rules {
		if strings.HasPrefix(pattern, p) {
			if len(p) > maxLength {
				maxLength = len(p)
				matchPattern = p
			}
		}
	}
	if maxLength < 1 {
		return nil
	}
	return this.rules[matchPattern]
}

func NewRouter() *Router {
	router := new(Router)
	router.rules = make(map[string]func(context *Context)interface {})
	return router
}

