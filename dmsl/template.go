package dmsl

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"path/filepath"
)

var Debug = false
var TemplateDir = ""
var DocType = "<!DOCTYPE html>"
var LeftDelim = "{"
var RightDelim = "}"

func Delims(l, r string) {
	LeftDelim = l
	RightDelim = r
}

func Mod(i, n int) bool {
	if (i % n) == 0 {
		return true
	}
	return false
}

var funcMap = template.FuncMap{
	"Mod": Mod,
}

type Template struct {
	Html   *template.Template
	Result string
}

func New() *Template {
	t := &Template{}
	t.Html = template.New("").Delims(LeftDelim, RightDelim).Funcs(funcMap)
	return t
}

func Parse(src []byte) (*Template, error) {
	t := New()
	err := t.Parse(src)
	return t, err
}

func (t *Template) Parse(src []byte) error {
	s, err := FilterParse(src)
	if err != nil {
		return err
	}
	_, err = t.Html.Parse(string(s))
	return err
}

func ParseString(src string) (*Template, error) {
	t := New()
	err := t.ParseString(src)
	return t, err
}

func (t *Template) ParseString(src string) error {
	return t.Parse([]byte(src))
}

func ParseFile(filename string) (*Template, error) {
	t := New()
	err := t.ParseFile(filename)
	return t, err
}

func (t *Template) ParseFile(filename string) error {
	b, err := ioutil.ReadFile(filepath.Join(TemplateDir, filename))
	if err != nil {
		return err
	}
	err = t.Parse(b)
	return err
}

func (t *Template) Execute(data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := t.Html.Execute(buf, data)
	if err != nil {
		return "", err
	}
	r, err := DocParse(buf.Bytes())
	return DocType + r, err
}
