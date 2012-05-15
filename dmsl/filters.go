package dmsl

import (
	"strings"
)

type filterFn func(*Filter) string

func js(filter *Filter) string {
	s := ""
	for _, v := range filter.Content {
		s += "<script type=\"text/javascript\" src=\"" + string(filter.Args) + string(v) + "\"/>"
	}
	return s
}

func css(filter *Filter) string {
	s := ""
	for _, v := range filter.Content {
		s += "<link type=\"text/css\" rel=\"stylesheet\" href=\"" + string(filter.Args) + string(v) + "\">"
	}
	return s
}

func extends(filter *Filter) string {
	bytes := Open(string(filter.Args), TemplateDir)
	return string(bytes)
}

func include(filter *Filter) string {
	ws := ""
	for i := 0; i < filter.Whitespace; i++ {
		ws += "\t"
	}
	bytes := Open(string(filter.Args), TemplateDir)
	s := strings.Split(string(bytes), "\n")
	for i, l := range s {
		if i == 0 { // gets inserted at original whitespace so no need to prepend
			continue
		}
		s[i] = ws + l
	}
	return strings.Join(s, "\n")
}
