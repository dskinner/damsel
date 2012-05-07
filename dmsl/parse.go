package dmsl

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func Open(filename string, dir string) []byte {
	b, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		log.Fatal(err)
	}
	return b
}

const eof = -1

var DefaultTag []byte = []byte("div")
var AttrId []byte = []byte("id")
var AttrClass []byte = []byte("class")

func LexerParse(bytes []byte, tmplDir string) *Elem {
	lexer := new(Lexer)
	lexer.bytes = bytes
	lexer.tmplDir = tmplDir
	lexer.Start()

	// combine #ids
	for _, elems := range lexer.ids {
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
	return lexer.root.children[0]
}

type stateFn func(*Lexer) stateFn

type Lexer struct {
	bytes []byte
	state stateFn
	pos   int
	start int
	action []byte

	tmplDir string
	textWs  float64
	curWs   float64
	prevWs  float64
	root    *Elem
	curElem *Elem
	ids     map[string][]*Elem
	cache   map[float64]*Elem
}

func (l *Lexer) Start() {
	l.state = lexWhiteSpace
	l.root = new(Elem)
	l.root.tag = []byte("root")
	l.ids = make(map[string][]*Elem)
	l.cache = make(map[float64]*Elem)

	for l.state != nil {
		l.state = l.state(l)
	}
}

func (l *Lexer) next() {
	l.pos++
}

func (l *Lexer) reset() {
	l.start = l.pos
}

func (l *Lexer) rune() rune {
	if l.pos >= len(l.bytes) {
		return eof
	}
	return rune(l.bytes[l.pos])
}

func (l *Lexer) discard() {
	l.pos++
	l.start = l.pos
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.bytes) {
		return eof
	}
	return rune(l.bytes[l.pos+1])
}

func (l *Lexer) NewElem() {
	if l.curWs == 0 {
		l.curElem = l.root.SubElement()
	} else if l.curWs > l.prevWs {
		l.curElem = l.cache[l.prevWs].SubElement()
	} else if l.curWs == l.prevWs {
		l.curElem = l.cache[l.prevWs].parent.SubElement()
	} else if l.curWs < l.prevWs {
		l.curElem = l.cache[l.curWs].parent.SubElement()
		for ws := range l.cache {
			if ws > l.curWs {
				delete(l.cache, ws)
			}
		}
	}
	l.cache[l.curWs] = l.curElem
}

func (l *Lexer) getBytes() []byte {
	return l.bytes[l.start:l.pos]
}

func (l *Lexer) AppendText() {
	if l.textWs == 0 || l.textWs > l.prevWs {
		l.curElem.text = append(l.curElem.text, l.getBytes()...)
	} else if l.textWs == l.prevWs {
		l.curElem.tail = append(l.curElem.tail, l.getBytes()...)
	} else if l.textWs < l.prevWs {
		l.cache[l.textWs].tail = append(l.cache[l.textWs].tail, l.getBytes()...)
	}
}

func (l *Lexer) AppendAttrKey() {
	l.curElem.attr = append(l.curElem.attr, [][][]byte{[][]byte{l.bytes[l.start:l.pos], nil}}...)
}

// These are the lexer states that will execute on each iteration based on which is set to Lexer.state

func lexWhiteSpace(l *Lexer) stateFn {
	switch l.rune() {
	case ' ', '\t':
		l.next()
		break
	case '\n':
		l.discard()
		break
	case '[': // continued attr
		l.discard()
		return lexAttributeKey
	case '{':
		l.textWs = float64(l.pos - l.start)
		l.reset()
		return lexText
	case '\\':
		l.textWs = float64(l.pos - l.start)
		l.discard()
		return lexText
	case eof:
		return nil
	default:
		l.curWs = float64(l.pos - l.start)
		l.start = l.pos
		l.NewElem()
		return lexHash
	}
	return lexWhiteSpace
}

func lexHashTag(l *Lexer) stateFn {
	switch l.rune() {
	case '#', '.', '[', ' ', '\t', '\n':
		l.curElem.tag = l.bytes[l.start:l.pos]
		return lexHash
	case eof:
		l.curElem.tag = l.bytes[l.start:l.pos]
		return nil
	default:
		l.next()
	}
	return lexHashTag
}

func lexHashId(l *Lexer) stateFn {
	switch l.rune() {
	case '#', '.', '[', ' ', '\t', '\n':
		l.curElem.id = l.bytes[l.start:l.pos]
		l.ids[string(l.curElem.id)] = append(l.ids[string(l.curElem.id)], l.curElem)
		return lexHash
	case eof:
		l.curElem.id = l.bytes[l.start:l.pos]
		return nil
	default:
		l.next()
	}
	return lexHashId
}

func lexHashClass(l *Lexer) stateFn {
	switch l.rune() {
	case '#', '.', '[', ' ', '\t', '\n':
		l.curElem.class = append(l.curElem.class, l.bytes[l.start:l.pos])
		return lexHash
	case eof:
		l.curElem.class = append(l.curElem.class, l.bytes[l.start:l.pos])
		return nil
	default:
		l.next()
	}
	return lexHashClass
}

func lexHash(l *Lexer) stateFn {
	switch l.rune() {
	case '%':
		l.discard()
		return lexHashTag(l)
	case '#':
		l.discard()
		return lexHashId(l)
	case '.':
		l.discard()
		return lexHashClass(l)
	case '!':
		l.discard()
		l.curElem.isComment = true
	case '[':
		l.discard()
		return lexAttributeKey
	case ' ', '\t':
		// check for inlined tag
		l.discard()
		
		switch l.rune() {
		case '.', '#', '%':
			l.prevWs = l.curWs
			l.curWs += 0.5 // inlines get ws++ to nest under prev elem
			l.NewElem()
			return lexHash
		case eof:
			return nil
		}
		
		l.textWs = 0 // this isn't multiline text, so 0
		return lexText
	case '\n':
		l.discard()
		l.prevWs = l.curWs
		return lexWhiteSpace
	default:
		l.next()
	}
	return lexHash
}

func lexAttributeKey(l *Lexer) stateFn {
	switch l.rune() {
	case '=':
		l.AppendAttrKey()
		l.discard()
		return lexAttributeValue
	case ']':
		l.AppendAttrKey()
		l.discard()
		
		// the order of the following two switch statements matter
		
		switch rune(l.bytes[l.pos]) {
		case eof:
			return lexWhiteSpace // TODO maybe should return an eof stateFn
		case '[':
			l.discard()
			return lexAttributeKey
		case '\n':
			l.discard()
			return lexWhiteSpace
		}
		
		// check if inline tag
		switch l.peek() {
		case '%', '#', '.':
			l.discard()
			l.NewElem()
			return lexHash
		}
		
		// should be text TODO should-be?!
		l.discard()
		l.textWs = 0 // not multiline text, so 0
		return lexText
	default:
		l.next()
	}
	return lexAttributeKey
}

func lexAttributeValue(l *Lexer) stateFn {
	// TODO string literals
	switch l.rune() {
	case ']':
		l.curElem.attr[len(l.curElem.attr)-1][1] = l.bytes[l.start:l.pos]
		l.discard()
		
		switch l.rune() {
		case '[': // check for another attr
			l.discard()
			return lexAttributeKey
		case eof:
			return nil
		}

		// TODO maybe quicker to just skip lexText and jump straight to lexWhiteSpace
		// check for end of line
		//if l.bytes[l.pos] == '\n' {
		//	l.discard()
		//	return lexWhiteSpace
		//}

		// check if inline tag
		switch l.peek() {
		case '%', '#', '.':
			l.discard()
			l.NewElem()
			return lexHash
		case eof:
			return nil
		}

		// should be text or newline
		if l.rune() != '\n' {
			l.discard()
		}
		l.textWs = 0 // not multiline text, so 0
		return lexText
	default:
		l.next()
	}
	return lexAttributeValue
}

func lexText(l *Lexer) stateFn {
	switch l.rune() {
	case '\n':
		l.AppendText()
		l.discard()
		l.prevWs = l.curWs
		return lexWhiteSpace
	case '{':
		return lexActionText
	case eof:
		l.AppendText()
		return nil
	default:
		l.next()
	}
	
	return lexText
}

func lexActionText(l *Lexer) stateFn {
	l.action = append(l.action, l.bytes[l.pos])
	switch l.rune() {
	case '}':
		f := handleAction(l)
		l.action = []byte{}
		return f
	case eof:
		return nil // TODO handle eof during unfinished action
	default:
		l.next()
	}
	return lexActionText
}

func lexAction(l *Lexer) stateFn {
	l.action = append(l.action, l.bytes[l.pos])
	switch l.rune() {
	case '}':
		f := handleAction(l)
		l.action = []byte{}
		return f
	case eof:
		return nil // TODO handle eof during unfinished action
	default:
		l.next()
	}
	return lexAction
}

// TODO implement include, will probably be slow due to needing to loop through and insert l.curWs after every line break, before insert into l.bytes
func handleAction(l *Lexer) stateFn {
	s := string(l.action)
	switch {
	case strings.HasPrefix(s, "{extends ") || strings.HasPrefix(s, "extends "):
		l.next()
		s2 := strings.Split(s, "\"")[1]
		bytes := Open(s2, l.tmplDir) // TODO use the configured template dir!!!
		l.bytes = append(l.bytes[:l.pos], append(bytes, l.bytes[l.pos:]...)...)
		l.start = l.pos
		return lexWhiteSpace
	case strings.HasPrefix(s, "{range "):
		l.curElem.actionEnds++
		return lexText
	case strings.HasPrefix(s, "{if "):
		l.curElem.actionEnds++
		return lexText
	case s == "{end}":
		if l.textWs > l.prevWs {
			l.cache[l.prevWs].actionEnds--
		} else if l.textWs == l.prevWs {
			l.cache[l.prevWs].parent.actionEnds--
		} else if l.textWs < l.prevWs { // TODO this may be wrong, refer to other areas in code for how this is done
			l.cache[l.textWs].parent.actionEnds--
		}
		return lexText
	}
	// 
	return lexText
}

