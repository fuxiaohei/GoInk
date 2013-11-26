package app

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type InkContext struct {
	// global objects
	Ink      *InkApp
	Request  *http.Request
	Response http.ResponseWriter

	// send status
	isSend bool

	// request params
	params    []string
	headers  map[string]string
	Ip        string
	Path      string
	Host      string
	Xhr       bool
	Protocol  string
	URI       string
	Method    string
	Refer     string
	UserAgent string
	// response params
	status  int
	content []byte

	// flash data
	flash map[string]string
}

// get params by index number.
// if out of range, return empty string
func (this *InkContext) Params(index int) string {
	if index + 1 > len(this.params) {
		return ""
	}
	return this.params[index]
}

// get all input data
func (this *InkContext) Input() map[string]string {
	data := make(map[string]string)
	for key, v := range this.Request.Form {
		data[key] = v[0]
	}
	return data
}

// get form string slice
func (this *InkContext) Strings(key string) []string {
	return this.Request.Form[key]
}

// get query string value
func (this *InkContext) String(key string) string {
	return this.Request.FormValue(key)
}

// get query string value with replacer value
func (this *InkContext) StringOr(key string, def string) string {
	value := this.String(key)
	if value == "" {
		return def
	}
	return value
}

// get query int value
func (this *InkContext) Int(key string) int {
	str := this.String(key)
	i, _ := strconv.Atoi(str)
	return i
}

// get query int value with replacer
func (this *InkContext) IntOr(key string, def int) int {
	i := this.Int(key)
	if i == 0 {
		return def
	}
	return i
}

// get query float value
func (this *InkContext) Float(key string) float64 {
	str := this.String(key)
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

// get query float value with replacer
func (this *InkContext) FloatOr(key string, def float64) float64 {
	f := this.Float(key)
	if f == 0.0 {
		return def
	}
	return f
}

// get query bool value
func (this *InkContext) Bool(key string) bool {
	str := this.String(key)
	b, _ := strconv.ParseBool(str)
	return b
}

// cookie getter and setter.
// if only key, get cookie in request.
// if set key,value and expire(string), set cookie in response
// expire time is in second
func (this *InkContext) Cookie(key string, value ...string) string {
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
func (this *InkContext) Flash(key string, value ...string) string {
	if len(value) > 0 {
		this.flash[key] = value[0]
		return ""
	}
	return this.flash[key]
}

// determine request suffix
// @todo add mime-type check
func (this *InkContext) Is(sfx string) bool {
	return path.Ext(this.Request.URL.Path) == "." + sfx
}

// get header info from request
func (this *InkContext) Get(key string) string {
	return this.Request.Header.Get(key)
}

// put header value into response
func (this *InkContext) Set(key string, value string) {
	this.headers[key] = value
}

// set status to response
func (this *InkContext) Status(status int) {
	this.status = status
}

// set redirect to response.
// do not redirect in this method, response is done in method "Send"
func (this *InkContext) Redirect(url string, status int) {
	this.Set("Location", url)
	this.status = status
}

// set content type to response
func (this *InkContext) ContentType(contentType string) {
	this.Set("Content-Type", contentType)
}

// set json data to response
func (this *InkContext) Json(data interface{}) {
	bytes, e := json.MarshalIndent(data, "", "    ")
	if e != nil {
		panic(e)
	}
	this.ContentType("application/json;charset=UTF-8")
	this.content = bytes
}

// set view rendered data to response
// render function depends on InkRender interface
func (this *InkContext) Render(tpl string, data map[string]interface{}) {
	var e error
	if data == nil {
		data = make(map[string]interface{})
	}
	data["Ink"] = this.Ink.storage
	data["Flash"] = this.flash
	data["Input"] = this.Input()
	this.Ink.Trigger("context.render.before", &tpl, &data)
	this.content, e = this.Ink.view.Render(tpl, data)
	if e != nil {
		panic(e)
	}
	this.Ink.Trigger("context.render.after", &this.content)
}

// determine response is sent or not
func (this *InkContext) IsEnd() bool {
	return this.isSend
}

// send response with content or status
// if content is empty, do not assign to response body content
func (this *InkContext) Send(content string, status ...int) {
	if this.isSend {
		return
	}
	if len(content) > 0 {
		this.content = []byte(content)
	}
	if len(status) > 0 {
		this.status = status[0]
	}
	this.Ink.Trigger("context.send.before", this)
	for name, value := range this.headers {
		this.Response.Header().Set(name, value)
	}
	// write direct context string
	this.Response.WriteHeader(this.status)
	this.Response.Write(this.content)
	this.isSend = true
	this.logContextSend()
}

// log context send out
func (this *InkContext) logContextSend() {
	if this.Ink.logger != nil {
		logData := []interface{}{
			this.Request.RemoteAddr,
			"- -",
					"[" + time.Now().Format(time.RFC822Z) + "]",
									`"` + this.Method + " " + this.URI + " " + this.Protocol + `"`,
			this.status,
			len(this.content),
			this.Request.UserAgent(),
		}
		if this.status >= 500 {
			this.Ink.logger.Error(logData...)
			return
		}
		this.Ink.logger.Log(logData...)
	}
}

// create new context object
func NewContext(app *InkApp, request *http.Request, response http.ResponseWriter) *InkContext {
	context := &InkContext{}
	context.Ink = app
	context.Request = request
	context.Response = response
	context.isSend = false
	// set params without empty value
	params := strings.Split(strings.Replace(request.URL.Path, path.Ext(request.URL.Path), "", -1), "/")
	context.params = []string{}
	for _, v := range params {
		if len(v) > 0 {
			context.params = append(context.params, v)
		}
	}
	// parse form always
	request.ParseForm()
	// assign request properties
	context.Ip = strings.Split(request.RemoteAddr, ":")[0]
	context.Path = request.URL.Path
	context.Host = strings.Split(request.Host, ":")[0]
	context.Xhr = context.Get("X-Requested-With") == "XMLHttpRequest"
	context.Protocol = request.Proto
	context.URI = request.RequestURI
	context.Method = request.Method
	context.Refer = request.Referer()
	context.UserAgent = request.UserAgent()
	// init response properties
	context.headers = map[string]string{"Content-Type": "text/html;charset=UTF-8"}
	context.status = 200
	// init flash data
	context.flash = make(map[string]string)
	context.Ink.Trigger("context.new", context)
	return context
}
