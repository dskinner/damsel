package main

import (
	"bytes"
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

type EOF struct{}

func (eof EOF) Error() string {
	return "EOF"
}

var DefaultTag []byte = []byte("div")
var AttrId []byte = []byte("id")
var AttrClass []byte = []byte("class")

const (
	LeftCarrot  = '<'
	Slash       = '/'
	RightCarrot = '>'
	Space       = ' '
	Equal       = '='
	Quote       = '"'
)

type Elem struct {
	parent   *Elem
	children []*Elem
	tag      []byte
	id       []byte
	class    [][]byte
	attr     [][][]byte
	text     []byte
	tail     []byte
}

func (el *Elem) SubElement() *Elem {
	newElem := new(Elem)
	newElem.tag = DefaultTag
	newElem.parent = el
	el.children = append(el.children, newElem)
	//fmt.Println("elem:", newElem)
	return newElem
}

func (el *Elem) String() string {
	buf := new(bytes.Buffer)
	el.ToString(buf)
	return buf.String()
}

func (el *Elem) ToString(buf *bytes.Buffer) {

	// TODO probably no need to make a new map for each subelement all the time
	keys := make(map[string]bool)

	buf.WriteRune(LeftCarrot)
	buf.Write(el.tag)

	if el.id != nil {
		buf.WriteRune(Space)
		buf.Write(AttrId)
		buf.WriteRune(Equal)
		buf.WriteRune(Quote)
		buf.Write(el.id)
		buf.WriteRune(Quote)
	}

	if el.class != nil {
		buf.WriteRune(Space)
		buf.Write(AttrClass)
		buf.WriteRune(Equal)
		buf.WriteRune(Quote)
		for i, bytes := range el.class {
			if i != 0 { // no space for first attr value
				buf.WriteRune(Space)
			}
			buf.Write(bytes)
		}
		buf.WriteRune(Quote)
	}

	for _, v := range el.attr {
		if keys[string(v[0])] {
			continue
		}

		buf.WriteRune(Space)
		buf.Write(v[0])
		buf.WriteRune(Equal)
		buf.WriteRune(Quote)
		buf.Write(v[1])
		buf.WriteRune(Quote)

		keys[string(v[0])] = true
	}

	buf.WriteRune(RightCarrot)
	buf.Write(el.text)

	for _, child := range el.children {
		child.ToString(buf)
	}

	buf.WriteRune(LeftCarrot)
	buf.WriteRune(Slash)
	buf.Write(el.tag)
	buf.WriteRune(RightCarrot)
	buf.Write(el.tail)
}

func LexerParse(bytes []byte, tmplDir string) *Elem {
	lexer := new(Lexer)
	lexer.bytes = bytes
	lexer.state = lexWhiteSpace
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

	tmplDir string
	textWs  int
	curWs   int
	prevWs  int
	root    *Elem
	curElem *Elem
	ids     map[string][]*Elem
	cache   map[int]*Elem
}

func (l *Lexer) Start() {
	l.root = new(Elem)
	l.root.tag = []byte("root")
	l.ids = make(map[string][]*Elem)
	l.cache = make(map[int]*Elem)

	for l.pos < len(l.bytes) {
		l.state = l.state(l)
	}
}

func (l *Lexer) next() {
	l.pos++
}

func (l *Lexer) discard() {
	l.pos++
	l.start = l.pos
}

func (l *Lexer) eof() bool {
	if l.pos >= len(l.bytes) {
		return true
	}
	return false
}

func (l *Lexer) peek() (byte, error) {
	if l.pos+1 >= len(l.bytes) {
		return '0', EOF{}
	}
	return l.bytes[l.pos+1], nil
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

func lexAction(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case '}':
		l.next()
		s := string(l.bytes[l.start:l.pos])
		if strings.HasPrefix(s, "extends") {
			s2 := strings.Split(s, "\"")[1]
			bytes := Open(s2, l.tmplDir) // TODO use the configured template dir!!!
			l.bytes = append(l.bytes[:l.pos], append(bytes, l.bytes[l.pos:]...)...)
		}
		// TODO implement include, will probably be slow due to needing to loop through and insert l.curWs after every line break, before insert into l.bytes
		l.start = l.pos
		return lexWhiteSpace
	default:
		l.next()
	}
	return lexAction
}

func lexWhiteSpace(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case ' ', '\t':
		l.next()
	case '\n':
		l.discard()
	case '[': // continued attr
		l.discard()
		return lexAttributeKey
	case '\\':
		l.textWs = l.pos - l.start
		l.discard()
		return lexText
	case '{': // TODO use the configured delimiter!!!
		l.discard()
		return lexAction
	default:
		l.curWs = l.pos - l.start
		//fmt.Println("  ws:", l.curWs)
		l.start = l.pos
		l.NewElem()
		return lexHash
	}
	return lexWhiteSpace
}

func lexHashTag(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case '#', '.', '[', ' ', '\t', '\n':
		l.curElem.tag = l.bytes[l.start:l.pos]
		return lexHash
	default:
		l.next()
	}
	if l.eof() {
		l.curElem.tag = l.bytes[l.start:l.pos]
	}
	return lexHashTag
}

func lexHashId(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case '#', '.', '[', ' ', '\t', '\n':
		l.curElem.id = l.bytes[l.start:l.pos]
		l.ids[string(l.curElem.id)] = append(l.ids[string(l.curElem.id)], l.curElem)
		return lexHash
	default:
		l.next()
	}
	if l.eof() {
		l.curElem.id = l.bytes[l.start:l.pos]
	}
	return lexHashId
}

func lexHashClass(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case '#', '.', '[', ' ', '\t', '\n':
		l.curElem.class = append(l.curElem.class, l.bytes[l.start:l.pos])
		return lexHash
	default:
		l.next()
	}
	if l.eof() {
		l.curElem.class = append(l.curElem.class, l.bytes[l.start:l.pos])
	}
	return lexHashClass
}

func lexHash(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case '%':
		l.discard()
		return lexHashTag(l)
	case '#':
		l.discard()
		return lexHashId(l)
	case '.':
		l.discard()
		return lexHashClass(l)
	case '[':
		l.discard()
		return lexAttributeKey
	case ' ', '\t':
		// check for inlined tag
		l.discard()

		//if b, err := l.peek(); err == nil {
		//fmt.Println(string(l.bytes[l.pos:l.pos+5]))
		switch l.bytes[l.pos] {
		case '.', '#', '%':
			l.prevWs = l.curWs
			l.curWs++ // inlines get ws++ to nest under prev elem
			l.NewElem()
			return lexHash
		}
		l.textWs = 0 // this isn't multiline text, so 0
		return lexText
		//}
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
	switch l.bytes[l.pos] {
	case '=':
		l.curElem.attr = append(l.curElem.attr, [][][]byte{[][]byte{l.bytes[l.start:l.pos], nil}}...)
		l.discard()
		return lexAttributeValue
	case ']':
		l.curElem.attr = append(l.curElem.attr, [][][]byte{[][]byte{l.bytes[l.start:l.pos], nil}}...)
		l.discard()
		
		if l.eof() {
			return lexWhiteSpace // TODO maybe should return an EOF stateFn
		}

		// check for another attr
		if l.bytes[l.pos] == '[' {
			l.discard()
			return lexAttributeKey
		}
		
		// check for end of line
		if l.bytes[l.pos] == '\n' {
			l.discard()
			return lexWhiteSpace
		}

		// check if inline tag
		if b, err := l.peek(); err == nil {
			switch b {
			case '%', '#', '.':
				l.discard()
				l.NewElem()
				return lexHash
			}
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
	switch l.bytes[l.pos] {
	case ']':
		l.curElem.attr[len(l.curElem.attr)-1][1] = l.bytes[l.start:l.pos]
		l.discard()

		if l.eof() {
			return lexWhiteSpace // TODO maybe should return an EOF stateFn
		}
		
		// check for another attr
		if l.bytes[l.pos] == '[' {
			l.discard()
			return lexAttributeKey
		}

		// TODO maybe quicker to just skip lexText and jump straight to lexWhiteSpace
		// check for end of line
		//if l.bytes[l.pos] == '\n' {
		//	l.discard()
		//	return lexWhiteSpace
		//}

		// check if inline tag
		if b, err := l.peek(); err == nil {
			switch b {
			case '%', '#', '.':
				l.discard()
				l.NewElem()
				return lexHash
			}
		}

		// should be text or newline
		if l.bytes[l.pos] != '\n' {
			l.discard()
		}
		l.textWs = 0 // not multiline text, so 0
		return lexText
	default:
		l.next()
	}
	return lexAttributeValue
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

func lexText(l *Lexer) stateFn {
	switch l.bytes[l.pos] {
	case '\n':
		l.AppendText()
		l.discard()
		l.prevWs = l.curWs
		return lexWhiteSpace
	default:
		l.next()
	}
	if l.eof() {
		l.AppendText()
	}
	return lexText
}
