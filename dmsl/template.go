package dmsl

import (
	"bytes"
	"fmt"
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

type Template struct {
	html   *template.Template
	Result string
}

func New() *Template {
	t := &Template{}
	t.html = template.New("").Delims(LeftDelim, RightDelim)
	return t
}

func Parse(src []byte) *Template {
	t := New()
	t.Parse(src)
	return t
}

func ParseString(src string) *Template {
	return Parse([]byte(src))
}

func (t *Template) Parse(src []byte) (*Template, error) {
	//s := parse.Parse(src, TemplateDir)
	s, err := ParserParse(src)
	if Debug {
		fmt.Println(s)
	}
	t.Result = DocType + s
	t.html.Parse(t.Result)
	return t, err
}

func ParseFile(filename string) (*Template, error) {
	t := New()
	_, err := t.ParseFile(filename)
	return t, err
}

func (t *Template) ParseFile(filename string) (*Template, error) {
	b, err := ioutil.ReadFile(filepath.Join(TemplateDir, filename))
	if err != nil {
		return t, err
	}
	_, err = t.Parse(b)
	return t, err
}

func (t *Template) Execute(data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := t.html.Execute(buf, data)
	return buf.String(), err
}
