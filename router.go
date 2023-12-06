package GoInk

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// RouterMethod enum type for HTTP methods
type RouterMethod string

const (
	MethodGet     RouterMethod = "GET"
	MethodPost    RouterMethod = "POST"
	MethodPut     RouterMethod = "PUT"
	MethodDelete  RouterMethod = "DELETE"
	MethodPatch   RouterMethod = "PATCH"
	MethodOptions RouterMethod = "OPTIONS"
	MethodHead    RouterMethod = "HEAD"
)

// Router instance provides router pattern and handlers.
type Router struct {
	routeTrie      *node
	routeCache     map[string]map[RouterMethod]routeCacheItem
	globalHandlers []Handler
}

// NewRouter creates a new router instance.
func NewRouter() *Router {
	return &Router{
		routeTrie:  newNode(),
		routeCache: make(map[string]map[RouterMethod]routeCacheItem),
	}
}

// AddMiddleware adds a global middleware to the router.
func (rt *Router) AddMiddleware(middleware ...Handler) {
	rt.globalHandlers = append(rt.globalHandlers, middleware...)
}

// AddRoute registers handlers with the specified method and pattern.
func (rt *Router) AddRoute(method RouterMethod, pattern string, fn ...Handler) {
	handlers := append(rt.globalHandlers, fn...)
	rt.routeTrie.insert(method, pattern, handlers)

	// Create cache for the route pattern and method
	if rt.routeCache[pattern] == nil {
		rt.routeCache[pattern] = make(map[RouterMethod]routeCacheItem)
	}
	rt.routeCache[pattern][method] = routeCacheItem{handlers: handlers}
}

// Find finds a matched route and its associated handlers.
func (rt *Router) Find(method RouterMethod, url string) (map[string]string, []Handler, error) {
	url = sanitizeURL(url)

	// Check cache first
	if cacheItem, ok := rt.routeCache[url][method]; ok {
		return nil, cacheItem.handlers, nil
	}

	params, handlers := rt.routeTrie.search(method, url)
	if params != nil {
		// Cache the matched route
		rt.routeCache[url][method] = routeCacheItem{handlers: handlers}
	} else {
		return nil, nil, errors.New("route not found")
	}

	return params, handlers, nil
}

// Node represents a node in the routing trie.
type node struct {
	children    map[string]*node
	handlers    map[RouterMethod][]Handler
	wildcard    bool
	paramKey    string
	isEndOfPath bool
}

// newNode creates a new node.
func newNode() *node {
	return &node{
		children: make(map[string]*node),
		handlers: make(map[RouterMethod][]Handler),
	}
}

// insert adds a route handler to the trie.
func (n *node) insert(method RouterMethod, path string, handlers []Handler) {
	segments := strings.Split(path, "/")
	currNode := n

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		if _, ok := currNode.children[segment]; !ok {
			currNode.children[segment] = newNode()
		}

		currNode = currNode.children[segment]
	}

	currNode.isEndOfPath = true
	currNode.handlers[method] = handlers
}

// search finds the route handler in the trie.
func (n *node) search(method RouterMethod, path string) (map[string]string, []Handler) {
	segments := strings.Split(path, "/")
	params := make(map[string]string)
	currNode := n

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		if nextNode, ok := currNode.children[segment]; ok {
			currNode = nextNode
		} else if nextNode, ok := currNode.children["*"]; ok {
			currNode = nextNode
			currNode.wildcard = true
		} else {
			return nil, nil
		}
	}

	if !currNode.isEndOfPath {
		return nil, nil
	}

	if currNode.wildcard {
		params[currNode.paramKey] = strings.Join(segments, "/")
	}

	return params, currNode.handlers[method]
}

// routeCacheItem stores cached route handlers.
type routeCacheItem struct {
	handlers []Handler
}

// sanitizeURL sanitizes the URL path.
func sanitizeURL(urlPath string) string {
	sfx := path.Ext(urlPath)
	urlPath = strings.TrimSuffix(urlPath, sfx)

	// Fix path end slash
	if !strings.HasSuffix(urlPath, "/") && sfx == "" {
		urlPath += "/"
	}

	return url.QueryEscape(urlPath)
}

// Handler defines route handler, middleware handler type.
type Handler func(context *Context)

// Context represents the context of the request.
type Context struct {
	Request  *http.Request  // You can add more request-specific fields
	Response *http.Response // You can add more response-specific fields
	// Add other necessary request context fields here
}
