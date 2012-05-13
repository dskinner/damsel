package dmsl

import (
	"strings"
)

type filterFn func([]byte, ...[]byte) string

func js(root []byte, files ...[]byte) string {
	s := ""
	for _, v := range files {
		s += "<script type=\"text/javascript\" src=\"" + string(root) + string(v) + "\"/>"
	}
	//return template.HTML(s)
	return s
}

func css(root []byte, files ...[]byte) string {
	s := ""
	for _, v := range files {
		s += "<link type=\"text/css\" rel=\"stylesheet\" href=\"" + string(root) + string(v) + "\">"
	}
	//return template.HTML(s)
	return s
}

func extends(filename []byte, content ...[]byte) string {
	bytes := Open(string(filename), TemplateDir)
	return string(bytes)
}

func include(filename []byte, n int) string {
	ws := ""
	for i := 0; i < n; i++ {
		ws += "\t"
	}
	bytes := Open(string(filename), TemplateDir)
	s := strings.Split(string(bytes), "\n")
	for i, l := range s {
		if i == 0 { // gets inserted at original whitespace so no need to prepend
			continue
		}
		s[i] = ws + l
	}
	return strings.Join(s, "\n")
}
