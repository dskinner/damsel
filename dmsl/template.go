package dmsl

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
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

/*
func js(root string, files ...string) template.HTML {
	s := ""
	for _, v := range files {
		s += "<script type=\"text/javascript\" src=\"" + root + v + "\"/>"
	}
	return template.HTML(s)
}

func css(root string, files ...string) template.HTML {
	s := ""
	for _, v := range files {
		s += "<link type=\"text/css\" rel=\"stylesheet\" href=\"" + root + v + "\">"
	}
	return template.HTML(s)
}
*/

var TemplateFuncMap = template.FuncMap{
	"js":  js,
	"css": css,
}

func include(s string) template.HTML {
	return template.HTML("include('" + s + "')")
}

type Template struct {
	html *template.Template
}

func New() *Template {
	t := &Template{}
	t.html = template.New("").Delims(LeftDelim, RightDelim)//.Funcs(TemplateFuncMap)
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

func (t *Template) Parse(src []byte) *Template {
	//s := parse.Parse(src, TemplateDir)
	s := LexerParse(src, TemplateDir).String()
	if Debug {
		fmt.Println(s)
	}
	t.html.Parse(s)
	return t
}

func ParseFile(filename string) *Template {
	t := New()
	t.ParseFile(filename)
	return t
}

func (t *Template) ParseFile(filename string) *Template {
	b, err := ioutil.ReadFile(filepath.Join(TemplateDir, filename))
	if err != nil {
		log.Fatal(err)
	}
	t.Parse(b)
	return t
}

func (t *Template) Execute(data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := t.html.Execute(buf, data)
	return DocType + buf.String(), err
}
