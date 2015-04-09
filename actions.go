package damsel

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"dasa.cc/damsel/parse"
)

func open(filename string, dir string) []byte {
	b, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func js(action *parse.Action) string {
	ws := action.Whitespace()
	s := ""
	for _, v := range action.Content {
		s += ws + "%script[type=\"text/javascript\"][src=\"" + string(action.Args) + string(v) + "\"]\n"
	}
	return s
}

func css(action *parse.Action) string {
	ws := action.Whitespace()
	s := ""
	for _, v := range action.Content {
		s += ws + "%link[rel=stylesheet][href=\"" + string(action.Args) + string(v) + "\"]\n"
	}
	return s
}

func extends(action *parse.Action) string {
	bytes := open(string(action.Args), TemplateDir)
	return string(bytes)
}

func include(action *parse.Action) string {
	ws := action.Whitespace()
	bytes := open(string(action.Args), TemplateDir)
	s := strings.Split(string(bytes), "\n")
	for i, l := range s {
		s[i] = ws + l
	}
	return strings.Join(s, "\n")
}
