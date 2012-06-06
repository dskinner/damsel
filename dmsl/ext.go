package dmsl

import (
	"strings"
	"fmt"
)

/* html/template */

var GoActionEnd Action = Action{ActionStart, "end"}

func ActionGoTemplate(s string) Action {
	switch {
	case strings.HasPrefix(s, "range "):
		return GoActionEnd
	case strings.HasPrefix(s, "if "):
		return GoActionEnd
	case strings.HasPrefix(s, "with "):
		return GoActionEnd
	case s == "end":
		return Action{ActionEnd, ""}
	}
	return Action{ActionIgnore, ""}
}


/* moustache */

func ActionMustache(s string) Action {
	fmt.Println(s)
	if strings.HasPrefix(s, "{#") {
		return Action{ActionStart, "{/"+s[2:]+"}"}
	}
	return Action{ActionIgnore, ""}
}