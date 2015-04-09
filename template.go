package damsel

import (
	"io/ioutil"
	"path/filepath"

	"dasa.cc/damsel/parse"
)

var Debug = false
var TemplateDir = ""

// SetPprint will force all document output to be pretty printed.
func SetPprint(b bool) {
	parse.Pprint = b
}

type Template struct {
	DocType string
	result  []byte
}

// New returns a new template with no data.
func New() *Template {
	t := &Template{}
	return t
}

// Parse returns a new template and initializes with the []byte content.
func Parse(src []byte) (*Template, error) {
	t := New()
	err := t.Parse(src)
	return t, err
}

// Parse initializes the template with the []byte content.
func (t *Template) Parse(src []byte) error {
	s, err := parse.ActionParse(src)
	if err != nil {
		return err
	}
	t.result = s
	return nil
}

// ParseString creates a new template and initializes with the string content.
func ParseString(src string) (*Template, error) {
	t := New()
	err := t.ParseString(src)
	return t, err
}

// ParseString initializes the template with the string content.
func (t *Template) ParseString(src string) error {
	return t.Parse([]byte(src))
}

// ParseFile creates a new template and initializes with the content of the given filename.
func ParseFile(filename string) (*Template, error) {
	t := New()
	err := t.ParseFile(filename)
	return t, err
}

// ParseFile initializes the template with the content of the given filename.
func (t *Template) ParseFile(filename string) error {
	b, err := ioutil.ReadFile(filepath.Join(TemplateDir, filename))
	if err != nil {
		return err
	}
	err = t.Parse(b)
	return err
}

// ParseResult returns intermediary result. Integration with other template engines such as html/template should use
// this as source, passing that result on to parse.DocParse.
func (t *Template) ParseResult() []byte {
	return t.result
}

// Result initiates the final parse phase and returns the document as a string.
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
