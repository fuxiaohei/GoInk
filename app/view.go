package app

import (
	"bytes"
	"html/template"
	"path"
	"strings"
)

type InkRender interface {
	Func(name string, fn interface{})
	Render(tpl string, data map[string]interface{}) ([]byte, error)
}

type InkView struct {
	dir     string
	funcMap template.FuncMap
}

// add template function
func (this *InkView) Func(name string, fn interface{}) {
	this.funcMap[name] = fn
}

// render template file with data
func (this *InkView) Render(tpl string, data map[string]interface{}) ([]byte, error) {
	templates := strings.Split(tpl, ",")
	name := path.Base(templates[0])
	for i, tp := range templates {
		templates[i] = path.Join(this.dir, tp)
	}
	t := template.New(name)
	t = t.Funcs(this.funcMap)
	t, e := t.ParseFiles(templates...)
	if e != nil {
		return []byte{}, e
	}
	var buffer bytes.Buffer
	e = t.ExecuteTemplate(&buffer, name, data)
	if e != nil {
		return []byte{}, e
	}
	return buffer.Bytes(), nil
}

// create new view with directory
func NewView(dir string) *InkView {
	view := &InkView{}
	view.dir = dir
	view.funcMap = make(template.FuncMap)
	return view
}
