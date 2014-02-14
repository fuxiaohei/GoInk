package GoInk

import (
	"bytes"
	"html/template"
	"os"
	"path"
)

// View instance provides simple template render.
type View struct {
	// template directory
	Dir string
	// view functions map
	FuncMap template.FuncMap
}

// Render renders template with data.
// Tpl is the filename under template directory.
func (v *View) Render(tpl string, data map[string]interface{}) ([]byte, error) {
	t := template.New(path.Base(tpl))
	t.Funcs(v.FuncMap)
	var (
		e    error
		file = path.Join(v.Dir, tpl)
	)
	t, e = t.ParseFiles(file)
	if e != nil {
		return nil, e
	}
	var buf bytes.Buffer
	e = t.Execute(&buf, data)
	if e != nil {
		return nil, e
	}
	return buf.Bytes(), nil
}

// Has checks the template file existing.
func (v *View) Has(tpl string) bool {
	f := path.Join(v.Dir, tpl)
	_, e := os.Stat(f)
	return e == nil
}

// NewView returns view instance with directory.
// It contains bundle template function HTML(convert string to template.HTML).
func NewView(dir string) *View {
	v := new(View)
	v.Dir = dir
	v.FuncMap = make(template.FuncMap)
	v.FuncMap["Html"] = func(str string) template.HTML {
		return template.HTML(str)
	}
	return v
}
