package dmsl

import (
	"fmt"
)

type FilterError struct {
	Value string
}

func (e *FilterError) Error() string {
	return fmt.Sprintf("damsel: filter \"%s\" unknown.", e.Value)
}

type ActionType int

const (
	ActionStart ActionType = iota
	ActionEnd
	ActionIgnore
)

type actionFn func(string) ActionType

func defActionHandler(s string) ActionType {
	return ActionIgnore
}

var ActionHandler actionFn = defActionHandler

type Filter struct {
	name       []byte
	start      int
	contentWs  int
	Args       []byte
	Content    [][]byte
	Whitespace int
}

type FuncMap map[string]filterFn

type Parser struct {
	lex     *lexer
	root    *Elem
	curElem *Elem
	prevWs  int
	curWs   int
	textWs  int
	ids     map[string][]*Elem
	cache   []*Elem
	filter  *Filter
	action  []byte
	funcMap FuncMap
}

func ParserParse(bytes []byte) (result string, err error) {
	p := new(Parser)
	p.funcMap = FuncMap{
		"js":      js,
		"css":     css,
		"extends": extends,
		"include": include,
	}
	p.root = new(Elem)
	p.root.tag = []byte("root")
	p.ids = make(map[string][]*Elem)

	p.lex = NewLexer(p)
	p.lex.bytes = bytes

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

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

	result = p.root.children[0].String()
	return result, err
}

func (p *Parser) ReadPos(pos int) rune {
	if pos >= len(p.lex.bytes) {
		return eof
	}
	return rune(p.lex.bytes[pos])
}

func (p *Parser) getBytes(t Token) []byte {
	return p.lex.bytes[t.start:t.end]
}

func (p *Parser) extendCache(el *Elem) {
	for i := el.ws - len(p.cache) + 1; i > 0; i-- {
		p.cache = append(p.cache, nil)
	}
	p.cache[el.ws] = el
}

func (p *Parser) NewElem() {
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

func (p *Parser) AppendAttrKey(t Token) {
	p.curElem.attr = append(p.curElem.attr, [][][]byte{[][]byte{p.lex.bytes[t.start:t.end], nil}}...)
}

func (p *Parser) AppendText(t Token) {
	if p.textWs == 0 || p.textWs > p.curWs {
		p.curElem.text = append(p.curElem.text, p.getBytes(t))
	} else if p.textWs == p.curWs {
		p.curElem.tail = append(p.curElem.tail, p.getBytes(t))
	} else if p.textWs < p.curWs {
		p.cache[p.textWs].tail = append(p.cache[p.textWs].tail, p.getBytes(t))
	}
}

func (p *Parser) handleAction(s string) {
	switch ActionHandler(s) {
	case ActionStart:
		p.curElem.actionEnds++
	case ActionEnd:
		if p.textWs > p.prevWs {
			p.cache[p.prevWs].actionEnds--
		} else if p.textWs == p.prevWs {
			p.cache[p.prevWs].parent.actionEnds--
		} else if p.textWs < p.prevWs { // TODO this may be wrong, refer to other areas in code for how this is done
			p.cache[p.textWs].parent.actionEnds--
		}
	}
}

func (p *Parser) handleFilterEnd(t Token) {
	name := string(p.filter.name)

	if p.funcMap[name] == nil {
		panic(&FilterError{name})
	}
	// TODO filterFn should return possible error
	result := p.funcMap[name](p.filter)

	//
	b := []byte(result)

	// need to evaluate filterFn result against normal lexing
	p.lex.bytes = append(p.lex.bytes[:p.filter.start], append(b, p.lex.bytes[t.start-1:]...)...)

	// reset pos to delete/insert point for lexer
	p.lex.pos = p.filter.start
	// reset start point to beginning of line
	// have to divide by 2 since CountWs returns * 2 due to inlines and ++ ws for them.
	//p.lex.start = p.lex.pos - (p.filter.Whitespace / 2)
	p.lex.start = p.filter.start
}

func (p *Parser) ReceiveToken(t Token) {
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
	case TokenFilterStart:
		p.filter = &Filter{start: t.start, Whitespace: CountWs(t)}
		break
	case TokenFilterName:
		p.filter.name = p.getBytes(t)
		break
	case TokenFilterArgs:
		p.filter.Args = p.getBytes(t)
		break
	case TokenFilterContentWs:
		if p.filter.contentWs == 0 {
			p.filter.contentWs = CountWs(t)
		}
	case TokenFilterContent:
		// TODO work on contentWs/2
		p.filter.Content = append(p.filter.Content, p.lex.bytes[(t.start+p.filter.contentWs/2):t.end])
		break
	case TokenFilterEnd:
		p.handleFilterEnd(t)
	case TokenAction:
		action := string(p.lex.bytes[t.start+len(LeftDelim) : t.end])
		p.handleAction(action)
		break
	case TokenEOF:
		// TODO
		break
	}
}
