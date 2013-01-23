package parse

import (
	"fmt"
)

// CountWs is only called for appropriate emitted tokens that are known to be
// whitespace. The token length is multiplied by two to account for inlined tags
// that occupy ws+1
func CountWs(t Token) int {
	return (t.end - t.start) * 2
}

type ActionError struct {
	Value string
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("damsel: action \"%s\" unknown.", e.Value)
}

type Action struct {
	name       []byte
	start      int
	contentWs  int
	Args       []byte
	Content    [][]byte
	whitespace int
}

func (a *Action) Whitespace() string {
	ws := ""
	for i := 0; i < a.whitespace/2; i++ {
		ws += "\t"
	}
	return ws
}

type ActionFn func(*Action) string

type FuncMap map[string]ActionFn

var DefaultFuncMap = map[string]ActionFn{}

type ActionParser struct {
	lex     *lexer
	action  *Action
	funcMap FuncMap
}

func ActionParse(bytes []byte) ([]byte, error) {
	p := new(ActionParser)
	p.funcMap = DefaultFuncMap
	p.lex = NewLexer(p)
	p.lex.bytes = bytes
	p.lex.Run()
	// TODO real errors
	return p.lex.bytes, nil
}

func (p *ActionParser) handleActionEnd(t Token) {
	name := string(p.action.name)

	if p.funcMap[name] == nil {
		panic(&ActionError{name})
	}
	// TODO actionFn should return possible error
	result := p.funcMap[name](p.action)

	// TODO just use []byte
	b := []byte(result)

	// need to evaluate actionFn result against normal lexing
	p.lex.bytes = append(p.lex.bytes[:p.action.start], append(b, p.lex.bytes[t.start-1:]...)...)

	// reset pos and start to delete/insert point for lexer
	p.lex.pos = p.action.start
	p.lex.start = p.action.start
}

func (p *ActionParser) ReceiveToken(t Token) {
	switch t.typ {
	case TokenActionStart:
		p.action = &Action{start: t.start, whitespace: CountWs(t)}
		break
	case TokenActionName:
		p.action.name = p.lex.bytes[t.start:t.end]
		break
	case TokenActionArgs:
		p.action.Args = p.lex.bytes[t.start:t.end]
		break
	case TokenActionContentWs:
		if p.action.contentWs == 0 {
			p.action.contentWs = CountWs(t)
		}
	case TokenActionContent:
		// BUG tmp fix for blank token
		if t.start == t.end {
			break
		}
		// TODO work on contentWs/2
		p.action.Content = append(p.action.Content, p.lex.bytes[(t.start+p.action.contentWs/2):t.end])
		break
	case TokenActionEnd:
		p.handleActionEnd(t)
	case TokenEOF:
		// TODO
		break
	}
}

type DocParser struct {
	lex     *lexer
	root    *Elem
	curElem *Elem
	prevWs  int
	curWs   int
	textWs  int
	ids     map[string][]*Elem
	cache   []*Elem
	action  []byte
}

func DocParse(bytes []byte) (result string, err error) {
	p := new(DocParser)
	p.root = new(Elem)
	p.root.tag = []byte("root")
	p.ids = make(map[string][]*Elem)

	p.lex = NewLexer(p)
	p.lex.bytes = bytes

	/*
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
	*/

	p.lex.Run()

	// combine #ids
	for _, elems := range p.ids {
		if len(elems) > 1 {
			for i := 1; i < len(elems); i++ {

				isSuper := false

				// check for [super] attr
				for k, v := range elems[i].attr {
					if string(v[0]) == "super" {
						isSuper = true
						elems[i].attr = append(elems[i].attr[:k], elems[i].attr[k+1:]...)
						break
					}
				}

				if isSuper {
					elems[0].children = append(elems[0].children, elems[i].children...)
				} else {
					elems[0].children = elems[i].children
				}
			}
		}
	}

	// BUG(d) DOCTYPE check is horrid and could potentially result in panic for non-conformant or bug-ridden dmsl docs.
	if p.root.children[0].isComment {
		result = p.root.children[0].String()
		result += p.root.children[1].String()
	} else {
		result = p.root.children[0].String()
	}
	return result, err
}

func (p *DocParser) ReadPos(pos int) rune {
	if pos >= len(p.lex.bytes) {
		return eof
	}
	return rune(p.lex.bytes[pos])
}

func (p *DocParser) extendCache(el *Elem) {
	for i := el.ws - len(p.cache) + 1; i > 0; i-- {
		p.cache = append(p.cache, nil)
	}
	p.cache[el.ws] = el
}

func (p *DocParser) NewElem() {
	if p.curWs == 0 {
		p.curElem = p.root.SubElement()
	} else if p.curWs > p.prevWs {
		p.curElem = p.cache[p.prevWs].SubElement()
	} else if p.curWs == p.prevWs {
		p.curElem = p.cache[p.prevWs].parent.SubElement()
	} else if p.curWs < p.prevWs {
		p.curElem = p.cache[p.curWs].parent.SubElement()
		// TODO is this some old archaic bit not needed anymore?
		// original issue was assuring that ident line up didn't mis match, so
		// 0
		//   1
		//     2
		//       3
		//   4
		//       5
		// where 5 wouldn't match anything except 4
		p.cache = p.cache[:p.curWs+1]
	}

	p.curElem.ws = p.curWs
	// TODO can i just setup a findByWhitespace on Elem and call on p.root instead of maintaining a slice
	p.extendCache(p.curElem)
}

func (p *DocParser) AppendAttrKey(t Token) {
	p.curElem.attr = append(p.curElem.attr, [][][]byte{[][]byte{p.lex.bytes[t.start:t.end], nil}}...)
}

func (p *DocParser) AppendText(t Token) {
	if p.textWs == 0 || p.textWs > p.curWs {
		p.curElem.text = append(p.curElem.text, p.lex.bytes[t.start:t.end])
	} else if p.textWs == p.curWs {
		p.curElem.tail = append(p.curElem.tail, p.lex.bytes[t.start:t.end])
	} else if p.textWs < p.curWs {
		p.cache[p.textWs].tail = append(p.cache[p.textWs].tail, p.lex.bytes[t.start:t.end])
	}
}

func (p *DocParser) ReceiveToken(t Token) {
	switch t.typ {
	case TokenElement:
		if t.start == 0 || rune(p.lex.bytes[t.start-1]) == '\n' {
			p.prevWs = p.curWs
			p.curWs = CountWs(t)
		} else { // handle inline
			p.prevWs = p.curWs
			p.curWs++
		}
		p.NewElem()
		break
	case TokenHashTag:
		p.curElem.tag = p.lex.bytes[t.start:t.end]
		break
	case TokenHashId:
		p.curElem.id = p.lex.bytes[t.start:t.end]
		p.ids[string(p.curElem.id)] = append(p.ids[string(p.curElem.id)], p.curElem)
		break
	case TokenHashClass:
		p.curElem.class = append(p.curElem.class, p.lex.bytes[t.start:t.end])
		break
	case TokenAttrKey:
		p.AppendAttrKey(t)
		break
	case TokenAttrValue:
		switch p.lex.bytes[t.start] {
		case '\'', '"':
			t.start++
			t.end--
			break
		}
		// TODO remove escapes
		p.curElem.attr[len(p.curElem.attr)-1][1] = p.lex.bytes[t.start:t.end]
		break
	case TokenText:
		p.AppendText(t)
		break
	case TokenTextWs:
		if t.start != 0 && rune(p.lex.bytes[t.start-1]) != '\n' {
			p.textWs = 0
		} else { // multiline text
			p.textWs = CountWs(t)
		}
		break
	case TokenComment:
		p.curElem.isComment = true
		break
	case TokenEOF:
		// TODO
		break
	}
}
