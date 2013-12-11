package Core

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	App      Base
	Request  *http.Request
	Response http.ResponseWriter
	Params   []string
	IsSend   bool
	flash    map[string]interface{}

	Base       string
	Url        string
	RequestUrl string
	Method     string
	Ip         string
	UserAgent  string
	Referer    string
	Host       string
	Ext        string
	IsSSL      bool
	IsAjax     bool

	Status int
	Header map[string]string
	Body   []byte
	RenderFunc func(tpl string, data map[string]interface {})([]byte, error)
}

// get params by index number.
// if out of range, return empty string
func (this *Context) Param(index int) string {
	if index+1 > len(this.Params) {
		return ""
	}
	return this.Params[index]
}

// get all input data
func (this *Context) Input() map[string]string {
	data := make(map[string]string)
	for key, v := range this.Request.Form {
		data[key] = v[0]
	}
	return data
}

// get form string slice
func (this *Context) Strings(key string) []string {
	return this.Request.Form[key]
}

// get query string value
func (this *Context) String(key string) string {
	return this.Request.FormValue(key)
}

// get query string value with replacer value
func (this *Context) StringOr(key string, def string) string {
	value := this.String(key)
	if value == "" {
		return def
	}
	return value
}

// get query int value
func (this *Context) Int(key string) int {
	str := this.String(key)
	i, _ := strconv.Atoi(str)
	return i
}

// get query int value with replacer
func (this *Context) IntOr(key string, def int) int {
	i := this.Int(key)
	if i == 0 {
		return def
	}
	return i
}

// get query float value
func (this *Context) Float(key string) float64 {
	str := this.String(key)
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

// get query float value with replacer
func (this *Context) FloatOr(key string, def float64) float64 {
	f := this.Float(key)
	if f == 0.0 {
		return def
	}
	return f
}

// get query bool value
func (this *Context) Bool(key string) bool {
	str := this.String(key)
	b, _ := strconv.ParseBool(str)
	return b
}

// cookie getter and setter.
// if only key, get cookie in request.
// if set key,value and expire(string), set cookie in response
// expire time is in second
func (this *Context) Cookie(key string, value ...string) string {
	if len(value) < 1 {
		c, e := this.Request.Cookie(key)
		if e != nil {
			return ""
		}
		return c.Value
	}
	if len(value) == 2 {
		t := time.Now()
		expire, _ := strconv.Atoi(value[1])
		t = t.Add(time.Duration(expire)*time.Second)
		cookie := &http.Cookie{
			Name:    key,
			Value:   value[0],
			Path:    "/",
			MaxAge:  expire,
			Expires: t,
		}
		http.SetCookie(this.Response, cookie)
		return ""
	}
	return ""
}

// get flash data
// flash data only available in this context
func (this *Context) Flash(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		this.flash[key] = value[0]
		return ""
	}
	return this.flash[key]
}

// get header info from request
func (this *Context) GetHeader(key string) string {
	return this.Request.Header.Get(key)
}

// set redirect to response.
// do not redirect in this method, response is done in method "Send"
func (this *Context) Redirect(url string, status... int) {
	this.Header["Location"] = url
	if len(status) > 0 {
		this.Status = status[0]
		return
	}
	this.Status = 302
}

// set content type to response
func (this *Context) ContentType(contentType string) {
	this.Header["Content-Type"] = contentType
}

// set json data to response
func (this *Context) Json(data interface{}) {
	bytes, e := json.MarshalIndent(data, "", "    ")
	if e != nil {
		panic(e)
	}
	this.ContentType("application/json;charset=UTF-8")
	this.Body = bytes
}

// render template
func (this *Context) Render(tpl string, data map[string]interface{}) {
	if this.RenderFunc != nil {
		var e error
		if data == nil {
			data = make(map[string]interface {})
		}
		data["Input"] = this.Input()
		data["Flash"] = this.flash
		this.Body, e = this.RenderFunc(tpl, data)
		if e != nil {
			panic(e)
		}
	}
}

// send response with content or status
func (this *Context) Send() {
	if this.IsSend {
		return
	}
	this.App.Listener.EmitAll("core.context.send.before", this)
	for name, value := range this.Header {
		this.Response.Header().Set(name, value)
	}
	// write direct context string
	this.Response.WriteHeader(this.Status)
	this.Response.Write(this.Body)
	this.IsSend = true
	this.App.Listener.EmitAll("core.context.send.after", this)
}

func NewContext(res http.ResponseWriter, req *http.Request, app Base) *Context {
	context := new(Context)
	req.ParseForm()
	context.Request = req
	context.Response = res
	//---------------
	params := strings.Split(strings.Replace(req.URL.Path, path.Ext(req.URL.Path), "", -1), "/")
	context.Params = make([]string, 0)
	for _, v := range params {
		if len(v) > 0 {
			context.Params = append(context.Params, v)
		}
	}
	//--------------
	context.Url = req.URL.Path
	context.RequestUrl = req.RequestURI
	context.Method = req.Method
	context.Ext = path.Ext(req.URL.Path)
	context.Host = req.Host
	context.Ip = strings.Split(req.RemoteAddr, ":")[0]
	context.IsAjax = req.Header.Get("X-Requested-With") == "XMLHttpRequest"
	context.IsSSL = req.TLS != nil
	context.Referer = req.Referer()
	context.UserAgent = req.UserAgent()
	context.Base = "://"+context.Host + "/"
	if context.IsSSL {
		context.Base = "https" + context.Base
	} else {
		context.Base = "http" + context.Base
	}
	//-------------
	context.Status = 200
	context.Header = make(map[string]string)
	context.Header["Content-Type"] = "text/html;charset=UTF-8"
	context.IsSend = false
	//-----------
	context.flash = make(map[string]interface{})
	context.App = app
	app.Listener.EmitAll("core.context.new", context)
	return context
}
