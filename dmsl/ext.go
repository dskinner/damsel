package dmsl

import "strings"

func ActionGoTemplate(s string) ActionType {
	switch {
	case strings.HasPrefix(s, "range "):
		return ActionStart
	case strings.HasPrefix(s, "if "):
		return ActionStart
	case strings.HasPrefix(s, "with "):
		return ActionStart
	case s == "end":
		return ActionEnd
	}
	return ActionIgnore
}
