package damsel

import (
	"dasa.cc/damsel/parse"
	"io/ioutil"
	"path/filepath"
)

var Debug = false
var TemplateDir = ""

func SetPprint(b bool) {
	parse.Pprint = b
}

type Template struct {
	DocType string
	result  []byte
}

func New() *Template {
	t := &Template{}
	return t
}

func Parse(src []byte) (*Template, error) {
	t := New()
	err := t.Parse(src)
	return t, err
}

func (t *Template) Parse(src []byte) error {
	s, err := parse.ActionParse(src)
	if err != nil {
		return err
	}
	t.result = s
	return nil
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

func (t *Template) Result() (string, error) {
	r, err := parse.DocParse(t.result)
	if err != nil {
		return "", err
	}
	return t.DocType + r, err
}

func init() {
	parse.DefaultFuncMap = map[string]parse.ActionFn{
		"js":      js,
		"css":     css,
		"extends": extends,
		"include": include,
	}
}
