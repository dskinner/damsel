package damsel

import (
	"bytes"
	"html/template"

	"dasa.cc/damsel/parse"
)

var (
	LeftDelim  = "{"
	RightDelim = "}"
	funcMap    = template.FuncMap{
		"Mod":   Mod,
		"StrEq": StrEq,
	}
)

func Delims(l, r string) {
	LeftDelim = l
	RightDelim = r
}

// Mod provides a modulus function to html/template.
func Mod(i, n int) bool {
	if (i % n) == 0 {
		return true
	}
	return false
}

// StrEq provides string equality check to html/template.
func StrEq(a, b string) bool {
	return a == b
}

// HtmlTemplate is an inefficient example of external template integration that is also used with tests
// using html/template.
type HtmlTemplate struct {
	Html *template.Template
	Dmsl *Template
}

func NewHtmlTemplate(tpl *Template) *HtmlTemplate {
	t := &HtmlTemplate{}
	t.Html = template.New("").Delims(LeftDelim, RightDelim).Funcs(funcMap)
	t.Dmsl = tpl
	return t
}

func (t *HtmlTemplate) Execute(data interface{}) (string, error) {
	_, err := t.Html.Parse(string(t.Dmsl.ParseResult()))
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := t.Html.Execute(buf, data); err != nil {
		return "", err
	}
	r, err := parse.DocParse(buf.Bytes())
	if err != nil {
		return "", err
	}
	return r, nil
}
