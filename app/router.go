package app

import "strings"

type InkRouter struct {
	routes map[string]func (context *InkContext) interface {}
}

// add route rule
func (this *InkRouter) Add(url string, method string, fn func (context *InkContext)interface {}) {
	this.routes[strings.ToUpper(method) + ":" + url] = fn
}

// get matched route function
// get the function by the longest matched pattern
func (this *InkRouter) Match(pattern string) func (context *InkContext)interface {} {
	maxLength, matchPattern := 0, ""
	for p, _ := range this.routes {
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
	return this.routes[matchPattern]
}

// show all register routes
// only show pattern
func (this *InkRouter) Routes() []string {
	routes, i := make([]string, len(this.routes)), 0
	for k, _ := range this.routes {
		routes[i] = k
	}
	return routes
}

// create new router
func NewRouter() *InkRouter {
	router := &InkRouter{}
	router.routes = make(map[string]func (context *InkContext)interface {})
	return router
}


