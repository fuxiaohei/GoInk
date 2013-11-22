package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

const (
	MODE_DEBUG = "debug"
	MODE_PRO   = "pro"
)

type InkApp struct {
	config    *InkConfig
	storage   map[string]string
	router    *InkRouter
	filter    *InkFilter
	staticDir string
	view      InkRender
	mode      string
	logger    *InkLogger
}

func (this *InkApp) Mode(mode ...string) string {
	if len(mode) > 0 {
		this.mode = mode[0]
		return ""
	}
	return this.mode
}

func (this *InkApp) debugMode() {
	this.Filter().EnablePrint(true)
}

// crash app with error
// print error and debug trace
// exit app
func (this *InkApp) Crash(e error) {
	fmt.Println(e)
	debug.PrintStack()
	os.Exit(0)
}

// app storage setter and getter
func (this *InkApp) Store(key string, value ...string) string {
	if len(value) > 0 {
		this.storage[key] = value[0]
		return ""
	}
	return this.storage[key]
}

// get config string value
func (this *InkApp) String(key string) string {
	return this.config.String(key)
}

// get config int value
func (this *InkApp) Int(key string) int {
	return this.config.Int(key)
}

// get config float value
func (this *InkApp) Float(key string) float64 {
	return this.config.Float(key)
}

// get config bool value
func (this *InkApp) Bool(key string) bool {
	return this.config.Bool(key)
}

// get config object
func (this *InkApp) Config() *InkConfig {
	return this.config
}

// set GET route func
func (this *InkApp) GET(url string, fn func (context *InkContext) interface{}) {
	this.router.Add(url, "GET", fn)
}

// set POST route func
func (this *InkApp) POST(url string, fn func (context *InkContext) interface{}) {
	this.router.Add(url, "POST", fn)
}

// set DELETE route func
func (this *InkApp) DELETE(url string, fn func (context *InkContext) interface{}) {
	this.router.Add(url, "DELETE", fn)
}

// set PUT route func
func (this *InkApp) PUT(url string, fn func (context *InkContext) interface{}) {
	this.router.Add(url, "PUT", fn)
}

// get router object
func (this *InkApp) Router() *InkRouter {
	return this.router
}

// get registered route rules
func (this *InkApp) Routes() []string {
	return this.router.Routes()
}

// bind on event function
func (this *InkApp) On(event string, fn interface{}) error {
	tmp := strings.Split(event, "@")
	if len(tmp) < 2 {
		return errors.New("invalid filter event name: " + event)
	}
	this.filter.Add(tmp[0], tmp[1], fn)
	return nil
}

// trigger event function
// if event_name@func_named, only trigger this named func, return result or error
// if event_name, do all funcs in this event, nil return
func (this *InkApp) Trigger(event string, args ...interface{}) ([]interface{}, error) {
	tmp := strings.Split(event, "@")
	if len(tmp) < 2 {
		// if no-named event function, call all functions in this event
		e := this.filter.FilterAll(event, args...)
		if len(e) > 0 {
			return nil, e[0]
		}
		return nil, nil
	}
	return this.filter.Filter(tmp[0], tmp[1], args...)
}

// wake event function
// run event functions in goroutine
// no return
func (this *InkApp) Wake(event string, args ...interface{}) {
	go this.Trigger(event, args...)
}

// get filter object
func (this *InkApp) Filter() *InkFilter {
	return this.filter
}

// set static directory
// if files are not int this directory, can't be read and then do routing
func (this *InkApp) Static(dir string) {
	this.staticDir = dir
}

// get view render object
func (this *InkApp) View() InkRender {
	return this.view
}

// log something
func (this *InkApp) Log(v... interface {}) {
	this.logger.Log(v...)
}

// log some errors
func (this *InkApp) LogErr(v... interface {}) {
	this.logger.Error(v...)
}

// listen server
func (this *InkApp) Listen() {
	if this.mode == MODE_DEBUG {
		this.debugMode()
	}
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			// determine in static dir, if file is in dir, do server file
			if len(this.staticDir) > 0 {
				if strings.HasPrefix(r.URL.Path[1:], this.staticDir) {
					http.ServeFile(rw, r, r.URL.Path[1:])
					return
				}
			}
			context := NewContext(this, r, rw)
			// do recover if panic when executing route function
			defer func() {
				e := recover()
				if e == nil {
					return
				}
				this.Trigger("router.run.error", context, errors.New(fmt.Sprint(e)))
				if context.IsEnd() {
					return
				}
				context.Status(http.StatusServiceUnavailable)
				context.Send(fmt.Sprintln(e) + string(debug.Stack()))
				if context.Ink.Mode() == MODE_DEBUG {
					debug.PrintStack()
				}
			}()
			// get matched route function
			fn := this.router.Match(r.Method + ":" + r.URL.Path)
			if fn != nil {
				// filter of before running
				this.Trigger("router.run.before", context)
				if context.IsEnd() {
					return
				}
				result := fn(context)
				// filter of after running
				this.Trigger("router.run.after", context, result)
				context.Send("")
				return
			}
			// if no matched route function, call "run null" filter
			this.Trigger("router.run.null", context)
			if context.IsEnd() {
				return
			}
			context.Send("", http.StatusNotFound)
		})
	addr := this.String("server.addr")
	this.Trigger("server.listen.before", &addr)
	e := http.ListenAndServe(addr, nil)
	if e != nil {
		this.Crash(e)
	}
}

// create new app with config file
func NewApp(configFile string, configType int) *InkApp {
	app := &InkApp{}
	var e error
	app.config, e = NewConfig(configFile, configType)
	if e != nil {
		app.Crash(e)
	}
	app.storage = make(map[string]string)
	app.router = NewRouter()
	app.filter = NewFilter("inkApp", true)
	app.view = NewView(app.String("view.dir"))
	app.mode = MODE_DEBUG
	app.logger = NewLogger(app.Config().StringOr("log.dir", "log"), app.Config().IntOr("log.clock", 300), app.Mode() == MODE_DEBUG)
	return app
}
