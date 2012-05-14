package dmsl

type ActionType int

const (
	ActionStart      ActionType = iota
	ActionEnd
	ActionIgnore
)

type actionFn func(string) ActionType

func defActionHandler(s string) ActionType {
	return ActionIgnore
}

var ActionHandler actionFn = defActionHandler

type Filter struct {
	start     int
	name      []byte
	args      []byte
	content   [][]byte
	ws        int
	contentWs int
}

type FuncMap map[string]filterFn

type Parser struct {
	bytes []byte
	root    *Elem
	curElem *Elem
	prevWs int
	curWs  int
	textWs int
	ids   map[string][]*Elem
	cache []*Elem
	filter *Filter
	action []byte
	funcMap FuncMap
}

func ParserParse(bytes []byte) string {
	p := new(Parser)
	p.funcMap = FuncMap{
		"js":      js,
		"css":     css,
		"extends": extends,
	}
	p.bytes = bytes
	p.root = new(Elem)
	p.root.tag = []byte("root")
	p.ids = make(map[string][]*Elem)

	l := new(Lexer)
	l.bytes = bytes
	l.state = lexWhiteSpace
	l.parser = p
	//l.tokens = make(chan Token, 100)

	
	for l.state != nil {
		l.state = l.state(l)
	}
	
	//go l.Run()
	//p.receiveTokens(l)

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

	return p.root.children[0].String()
}

func CountWhitespace(t Token) int {
	return (t.end - t.start) * 2
}

func (p *Parser) getBytes(t Token) []byte {
	return p.bytes[t.start:t.end]
}

func (p *Parser) extendCache(el *Elem) {
	for i := el.ws-len(p.cache)+1; i > 0; i-- {
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
		p.cache = p.cache[:p.curWs+1]
	}
	
	p.curElem.ws = p.curWs
	p.extendCache(p.curElem)
}

func (p *Parser) AppendAttrKey(t Token) {
	p.curElem.attr = append(p.curElem.attr, [][][]byte{[][]byte{p.bytes[t.start:t.end], nil}}...)
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

func (p *Parser) handleFilterContentDefault(t Token, l *Lexer) {
	if p.filter.contentWs == 0 {
		p.filter.contentWs = CountWhitespace(t)
	}

	if CountWhitespace(t) <= p.filter.ws {
		var result string
		// TODO need better filter scheme so no special casing here
		if string(p.filter.name) == "include" {
			result = include(p.filter.args, p.filter.ws)
		} else {
			// TODO check that p.filter.name is actually in funcMap
			result = p.funcMap[string(p.filter.name)](p.filter.args, p.filter.content...)
		}
		b := []byte(result)
		// need to evaluate filterFn result against normal lexing
		p.bytes = append(p.bytes[:p.filter.start], append(b, p.bytes[t.start-1:]...)...)
		l.bytes = p.bytes
		// TODO reset pos to delete/insert point for lexer
		l.pos = p.filter.start
		// TODO reset start point to beginning of line
		// TODO have to divide by 2 since CountWhitespace returns * 2 due to inlines and ++ ws for them. abstract this maybe?
		l.start = l.pos - (p.filter.ws / 2)
		l.filterDone = true
	}

	//runtime.Gosched()
}

func (p *Parser) receiveTokens(t Token, l *Lexer) {
//LOOP:
	//for {
		//t := <- l.tokens
		switch t.typ {
		case TokenElement:
			if t.start == 0 || rune(p.bytes[t.start-1]) == '\n' {
				p.prevWs = p.curWs
				p.curWs = CountWhitespace(t)
			} else { // handle inline
				p.prevWs = p.curWs
				p.curWs++
			}
			p.NewElem()
			break
		case TokenHashTag:
			p.curElem.tag = p.bytes[t.start:t.end]
			break
		case TokenHashId:
			p.curElem.id = p.bytes[t.start:t.end]
			p.ids[string(p.curElem.id)] = append(p.ids[string(p.curElem.id)], p.curElem)
			break
		case TokenHashClass:
			p.curElem.class = append(p.curElem.class, p.bytes[t.start:t.end])
			break
		case TokenAttrKey:
			p.AppendAttrKey(t)
			break
		case TokenAttrValue:
			switch p.bytes[t.start] {
			case '\'', '"':
				t.start++
				t.end--
				break
			}
			p.curElem.attr[len(p.curElem.attr)-1][1] = p.bytes[t.start:t.end]
			break
		case TokenText:
			p.AppendText(t)
			break
		case TokenTextWs:
			if t.start != 0 && rune(p.bytes[t.start-1]) != '\n' {
				p.textWs = 0
			} else { // multiline text
				p.textWs = CountWhitespace(t)
			}
			break
		case TokenComment:
			p.curElem.isComment = true
			break
		case TokenFilter:
			p.filter = &Filter{start: t.start, ws: CountWhitespace(t)}
			break
		case TokenFilterName:
			p.filter.name = p.getBytes(t)
			break
		case TokenFilterArgs:
			p.filter.args = p.getBytes(t)
			break
		case TokenFilterContent:
			p.filter.content = append(p.filter.content, p.bytes[(t.start+p.filter.contentWs):t.end])
			break
		case TokenFilterContentDefault:
			p.handleFilterContentDefault(t, l)
		case TokenAction:
			action := string(p.bytes[t.start+len(LeftDelim):t.end])
			p.handleAction(action)
			break
		case TokenEOF:
			//break LOOP
			break
		}
	//}
}
