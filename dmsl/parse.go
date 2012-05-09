package dmsl

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

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
	bytes := Open(string(filename), "tests")
	return string(bytes)
}

type filterFn func([]byte, ...[]byte) string
type FuncMap map[string]filterFn

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

type Filter struct {
	start     int
	name      []byte
	args      []byte
	content   [][]byte
	ws        float64
	contentWs float64
}

type stateFn func(*Lexer) stateFn

type Lexer struct {
	bytes  []byte
	state  stateFn
	pos    int
	start  int
	action []byte

	tmplDir string
	textWs  float64
	curWs   float64
	prevWs  float64
	root    *Elem
	curElem *Elem
	ids     map[string][]*Elem
	cache   map[float64]*Elem
	
	filter *Filter
	funcMap FuncMap
}

func (l *Lexer) Start() {
	l.funcMap = FuncMap {
		"js": js,
		"css": css,
		"extends": extends,
	}
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
	if l.textWs == 0 || l.textWs > l.curWs {
		l.curElem.text = append(l.curElem.text, l.getBytes()...)
	} else if l.textWs == l.curWs {
		l.curElem.tail = append(l.curElem.tail, l.getBytes()...)
	} else if l.textWs < l.curWs {
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
	case '[':
		l.discard()
		return lexAttributeKey
	case '/':
		l.discard()
		return lexComment
	case ':':
		l.filter = &Filter{start: l.pos, ws: float64(l.pos-l.start)}
		l.discard()
		return lexFilter
	case '%', '#', '.', '!':

		if l.start == 0 || rune(l.bytes[l.start-1]) == '\n' {
			l.prevWs = l.curWs
			l.curWs = float64(l.pos - l.start)
		} else { // handle inline
			l.prevWs = l.curWs
			l.curWs += 0.5 // inlines get ws++ to nest under prev elem
		}

		l.NewElem()
		return lexHash

	case eof:
		return nil
	default:
		// handle text
		if l.start != 0 && rune(l.bytes[l.start-1]) != '\n' {
			l.textWs = 0
		} else { // multiline text
			l.textWs = float64(l.pos - l.start)
		}

		if l.rune() == '\\' {
			l.discard()
		}

		l.reset()
		return lexText
	}

	return lexWhiteSpace
}

// lexFilter stands alone for parsing, not mingling with lexWhiteSpace until
// it's completely finished.
func lexFilter(l *Lexer) stateFn {
	switch l.rune() {
	case ' ', '\t':
		l.filter.name = l.getBytes()
		l.discard()
		return lexFilterArgs
	case '\n':
		l.filter.name = l.getBytes()
		l.discard()
		return lexFilterContent
	case eof:
		return nil
	default:
		l.next()
	}
	return lexFilter
}

func lexFilterArgs(l *Lexer) stateFn {
	switch l.rune() {
	case '\n':
		l.filter.args = l.getBytes()
		l.discard()
		return lexFilterContent
	case eof:
		return nil
	default:
		l.next()
	}
	return lexFilterArgs
}

func lexFilterWhiteSpace(l *Lexer) stateFn {
	switch l.rune() {
		case ' ', '\t':
			l.next()
		case '\n':
			break
	}
	return lexFilterWhiteSpace
}

func lexFilterContent(l *Lexer) stateFn {
	switch l.rune() {
	case ' ', '\t':
		l.next()
		break
	case '\n':
		l.filter.content = append(l.filter.content, l.bytes[l.start+int(l.filter.contentWs):l.pos])
		l.discard()
		break
	default:
		// TODO use lexFilterWhiteSpace so we aren't checking this all the time
		if l.filter.contentWs == 0 {
			l.filter.contentWs = float64(l.pos - l.start)
		}
		
		if float64(l.pos - l.start) <= l.filter.ws {
			
			// TODO check that l.filter.name is actually in funcMap
			result := l.funcMap[string(l.filter.name)](l.filter.args, l.filter.content...)
			b := []byte(result)
			
			// need to evaluate filterFn result against normal lexing
			l.bytes = append(l.bytes[:l.filter.start], append(b, l.bytes[l.start-1:]...)...)
			// reset pos to delete/insert point for lexer
			l.pos = l.filter.start
			// reset start point to beginning of line
			l.start = l.pos-int(l.filter.ws)
			
			return lexWhiteSpace
		}
		
		l.next()
	}
	return lexFilterContent
}

func lexComment(l *Lexer) stateFn {
	switch l.rune() {
	case '\n':
		l.discard()
		return lexWhiteSpace
	case eof:
		return nil
	default:
		l.next()
	}

	return lexComment
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
		return lexWhiteSpace
	default:
		return lexWhiteSpace
	}
	return lexHash
}

func lexHashTag(l *Lexer) stateFn {
	switch l.rune() {
	case '#', '.', '[', ' ', '\t', '\n', eof:
		l.curElem.tag = l.bytes[l.start:l.pos]
		return lexHash
	default:
		l.next()
	}
	return lexHashTag
}

func lexHashId(l *Lexer) stateFn {
	switch l.rune() {
	case '#', '.', '[', ' ', '\t', '\n', eof: // TODO technically there should never be multiple ids, and damsel should be strict regarding this due to extends
		l.curElem.id = l.bytes[l.start:l.pos]
		l.ids[string(l.curElem.id)] = append(l.ids[string(l.curElem.id)], l.curElem)
		return lexHash
	default:
		l.next()
	}
	return lexHashId
}

func lexHashClass(l *Lexer) stateFn {
	switch l.rune() {
	case '#', '.', '[', ' ', '\t', '\n', eof:
		l.curElem.class = append(l.curElem.class, l.bytes[l.start:l.pos])
		return lexHash
	default:
		l.next()
	}
	return lexHashClass
}

func lexAttributeKey(l *Lexer) stateFn {
	switch l.rune() {
	case '=':
		l.AppendAttrKey()
		l.discard()

		// check for literal value
		switch l.rune() {
		case '\'', '"':
			l.next()
			return lexAttributeValueLiteral
		}

		return lexAttributeValue
	case ']':
		l.AppendAttrKey()
		l.discard()
		return lexWhiteSpace
	default:
		l.next()
	}
	return lexAttributeKey
}

func lexAttributeValueLiteral(l *Lexer) stateFn {
	switch l.rune() {
	case rune(l.bytes[l.start]):
		if l.bytes[l.pos-1] == '\\' { // check for escaped quote
			l.next()
			break
		} else {
			return lexAttributeValue
		}
	default:
		l.next()
	}
	return lexAttributeValueLiteral
}

func lexAttributeValue(l *Lexer) stateFn {
	switch l.rune() {
	case ']':
		// was this a string literal?
		switch rune(l.bytes[l.start]) {
		case '\'', '"':
			l.curElem.attr[len(l.curElem.attr)-1][1] = l.bytes[l.start+1 : l.pos-1]
		default:
			l.curElem.attr[len(l.curElem.attr)-1][1] = l.bytes[l.start:l.pos]
		}
		l.discard()

		return lexWhiteSpace
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
		return nil
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
		return nil
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
		bytes := append([]byte{'\n'}, Open(s2, l.tmplDir)...)
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
