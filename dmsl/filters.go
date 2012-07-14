package dmsl

import (
	"strings"
)

type filterFn func(*Filter) string

func js(filter *Filter) string {
	ws := ""
	for i := 0; i < filter.Whitespace/2; i++ {
		ws += "\t"
	}
	s := ws
	for _, v := range filter.Content {
		s += "<script type=\"text/javascript\" src=\"" + string(filter.Args) + string(v) + "\"/>"
	}
	return s
}

func css(filter *Filter) string {
	ws := ""
	for i := 0; i < filter.Whitespace/2; i++ {
		ws += "\t"
	}
	s := ""
	for _, v := range filter.Content {
		//s += "<link type=\"text/css\" rel=\"stylesheet\" href=\"" + string(filter.Args) + string(v) + "\">"
		s += ws + "%link[rel=stylesheet][href=\"" + string(filter.Args) + string(v) + "\"]\n"
	}
	return s
}

func extends(filter *Filter) string {
	bytes := Open(string(filter.Args), TemplateDir)
	return string(bytes)
}

func include(filter *Filter) string {
	ws := ""
	for i := 0; i < filter.Whitespace/2; i++ {
		ws += "\t"
	}
	bytes := Open(string(filter.Args), TemplateDir)
	s := strings.Split(string(bytes), "\n")
	for i, l := range s {
		s[i] = ws + l
	}
	return strings.Join(s, "\n")
}
