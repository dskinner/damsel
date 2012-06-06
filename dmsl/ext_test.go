package dmsl

import (
	"github.com/hoisie/mustache"
	"testing"
	"fmt"
)

func TestMustache(t *testing.T) {
	ActionHandler = ActionMustache
	s := `
%html %body
	{{#c}}
	%p hello {{name}}
`
	tpl := Parse([]byte(s))
	r := tpl.Result
	
	fmt.Println(r)
	
	m := make(map[string][]map[string]string)
	m["c"] = []map[string]string{
		map[string]string {
			"name": "Daniel",
		},
	}
	fmt.Println(m)

	data := mustache.Render(r, m)
	
	fmt.Println(data)
}