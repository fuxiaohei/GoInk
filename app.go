package GoInk

import (
	"github.com/fuxiaohei/GoInk/Core"
	"os"
	"path"
	"net/http"
	"fmt"
	"errors"
)

type Simple struct {
	Core.Base
	dispatchFunc func(context *Core.Context)
	errorFunc func(context *Core.Context, errorStatus int, errorObj error)
}

func (this *Simple) bootstrap() {
	this.Listener.AddListener("core.context.new", "add_render_func", func(context *Core.Context) {
			context.RenderFunc = this.View.Render
		})
	this.Router.Get("/", func(context *Core.Context) interface {} {
			context.Body = []byte("It Works !")
			return nil
		})
}

func (this *Simple) HandleDefault(handler func(context *Core.Context)) {
	this.dispatchFunc = handler
}

func (this *Simple) HandleRecover(handler func(context *Core.Context, errorStatus int, errorObj error)) {
	this.errorFunc = handler
}

func (this *Simple) Run() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
			file := path.Join(this.Root, req.URL.Path)
			fi, e := os.Stat(file)
			if e == nil && !fi.IsDir() {
				this.Listener.EmitAll("server.static.before", file)
				http.ServeFile(res, req, file)
				this.Listener.EmitAll("server.static.after", file)
				return
			}
			context := Core.NewContext(res, req, this.Base)
			this.Listener.EmitAll("server.dynamic.before", context)
			defer func() {
				e := recover()
				if e == nil {
					return
				}
				err := errors.New(fmt.Sprint(e))
				this.Listener.EmitAll("server.error.before", context, err)
				if this.errorFunc != nil {
					this.errorFunc(context, http.StatusServiceUnavailable, err)
					return
				}
				http.Error(res, err.Error(), http.StatusServiceUnavailable)
				this.Listener.EmitAll("server.error.after", context, err)
			}()
			this.dispatchFunc(context)
		})
	this.Listener.EmitAll("server.run.before", this)
	e := http.ListenAndServe(this.Config.StringOr("server.addr", "localhost:8080"), nil)
	if e != nil {
		fmt.Println(e)
	}
}

func NewSimple(configFile string) (*Simple, error) {
	s := new(Simple)
	//-----------
	var e error
	s.Root, e = os.Getwd()
	if e != nil {
		return nil, e
	}
	//------------
	if configFile == "" {
		s.Config, e = Core.NewConfig([]byte("{}"), "json")
	}else {
		s.Config, e = Core.NewConfigFromFile(path.Join(s.Root, configFile), "json")
	}
	if e != nil {
		return nil, e
	}
	//-------------
	s.Router = Core.NewRouter()
	s.Listener = Core.NewListener()
	s.Logger = Core.NewLogger(s.Config.StringOr("log.dir", "log"), s.Config.IntOr("log.clock", 300))
	s.View = Core.NewView(s.Config.StringOr("view.dir", "view"))
	//------------
	s.HandleDefault(func(context *Core.Context) {
		fn := s.Router.Match(context.Method+":"+context.Url)
		if fn == nil {
			s.Listener.EmitAll("server.notfound.before")
			http.Error(context.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			s.Listener.EmitAll("server.notfound.after")
			return
		}
		result := fn(context)
		if !context.IsSend {
			context.Send()
		}
		s.Listener.EmitAll("server.dynamic.after", context, result)
	})
	//-------------
	s.bootstrap()
	return s, nil
}

