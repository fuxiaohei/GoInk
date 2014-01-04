package Core

import (
	"bytes"
	"errors"
	"html/template"
	"path"
	"strings"
	"fmt"
)

type layout struct {
	tpl  *template.Template
	file  string
	files []string
}

type View struct {
	dir     string
	layouts map[string]*layout
	funcMap template.FuncMap
}

func (this *View) NewLayout(name string, files ...string) error {
	templates := files
	templateName := path.Base(templates[0])
	for i, tp := range templates {
		templates[i] = path.Join(this.dir, tp)
	}
	t := template.New(templateName)
	t = t.Funcs(this.funcMap)
	t, e := t.ParseFiles(templates...)
	if e != nil {
		return e
	}
	layout := new(layout)
	layout.tpl = t
	layout.file = templateName
	layout.files = templates
	this.layouts[name] = layout
	if IsDev() {
		fmt.Println("[Core.View] register layout '"+name + "` :", files)
	}
	return nil
}

func (this *View) NewFunc(name string, fn interface{}) {
	this.funcMap[name] = fn
}

func (this *View) renderFile(tpl string, data map[string]interface{}) (string, error) {
	templates := strings.Split(tpl, ",")
	name := path.Base(templates[0])
	for i, tp := range templates {
		templates[i] = path.Join(this.dir, tp)
	}
	t := template.New(name)
	t = t.Funcs(this.funcMap)
	t, e := t.ParseFiles(templates...)
	if e != nil {
		return "", e
	}
	var buffer bytes.Buffer
	e = t.ExecuteTemplate(&buffer, name, data)
	if e != nil {
		return "", e
	}
	return buffer.String(), nil
}

func (this *View) renderLayout(name string, data map[string]interface{}) (string, error) {
	if this.layouts[name] == nil {
		return "", errors.New("no layout "+name)
	}
	layout := this.layouts[name]
	var buffer bytes.Buffer
	if IsDev() {
		t := template.New(layout.file)
		t = t.Funcs(this.funcMap)
		t, _ = t.ParseFiles(layout.files...)
		layout.tpl = t
	}
	e := layout.tpl.ExecuteTemplate(&buffer, layout.file, data)
	if e != nil {
		return "", e
	}
	return buffer.String(), nil
}

func (this *View) Render(tpl string, data map[string]interface{}) ([]byte, error) {
	tplKeys := strings.Split(tpl, ":")
	if len(tplKeys) > 1 {
		tplString, e := this.renderFile(tplKeys[1], data)
		if e != nil {
			return []byte{}, e
		}
		layoutString, e := this.renderLayout(tplKeys[0], data)
		if e != nil {
			return []byte{}, e
		}
		return []byte(strings.Replace(layoutString, "{@Content}", tplString, -1)), nil
	}
	tplString, e := this.renderFile(tplKeys[0], data)
	return []byte(tplString), e
}

// create new view with directory
func NewView(dir string) *View {
	view := new(View)
	view.dir = dir
	view.layouts = make(map[string]*layout)
	view.funcMap = make(template.FuncMap)
	view.funcMap["Html"] = func(str string) template.HTML {
		return template.HTML(str)
	}
	return view
}
