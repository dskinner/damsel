package dmsl

import "bytes"

const (
	LeftCarrot   = '<'
	Slash        = '/'
	RightCarrot  = '>'
	Space        = ' '
	Equal        = '='
	Quote        = '"'
	Exclamation  = '!'
	Hyphen       = '-'
	LeftBracket  = '['
	RightBracket = ']'
)

var DefaultTag []byte = []byte("div")
var AttrId []byte = []byte("id")
var AttrClass []byte = []byte("class")

type Elem struct {
	parent     *Elem
	children   []*Elem
	ws int
	tag        []byte
	id         []byte
	class      [][]byte
	attr       [][][]byte
	text       []byte
	tail       []byte
	actionEnds int
	isComment bool
}

func (el *Elem) SubElement() *Elem {
	newElem := new(Elem)
	newElem.tag = DefaultTag
	newElem.parent = el
	el.children = append(el.children, newElem)
	return newElem
}

func (el *Elem) String() string {
	buf := new(bytes.Buffer)
	el.ToString(buf)
	return buf.String()
}

func contains(container [][]byte, item []byte) bool {
	for _, x := range container {
		if bytes.Equal(x, item) {
			return true
		}
	}
	return false
}

func (el *Elem) ToString(buf *bytes.Buffer) {

	// TODO get this `if` out of here
	if el.isComment {
		buf.WriteRune(LeftCarrot)
		buf.WriteRune(Exclamation)
		buf.WriteRune(Hyphen)
		buf.WriteRune(Hyphen)

		isCond := len(el.attr) == 1

		if isCond {
			buf.WriteRune(LeftBracket)
			buf.Write(el.attr[0][0])
			buf.WriteRune(RightBracket)
			buf.WriteRune(RightCarrot)
		} else {
			buf.Write(el.text)
		}

		for _, child := range el.children {
			child.ToString(buf)
		}

		if isCond {
			buf.WriteString("<![endif]-->")
		} else {
			buf.WriteRune(Hyphen)
			buf.WriteRune(Hyphen)
			buf.WriteRune(RightCarrot)
		}
		return
	}
	
	keys := [][]byte{}

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
		if contains(keys, v[0]) {
			continue
		}

		buf.WriteRune(Space)
		buf.Write(v[0])
		buf.WriteRune(Equal)
		buf.WriteRune(Quote)
		buf.Write(v[1])
		buf.WriteRune(Quote)

		keys = append(keys, v[0])
	}

	buf.WriteRune(RightCarrot)
	buf.Write(el.text)

	for _, child := range el.children {
		child.ToString(buf)
	}

	for i := 0; i < el.actionEnds; i++ {
		buf.WriteString("{end}")
	}
	buf.WriteRune(LeftCarrot)
	buf.WriteRune(Slash)
	buf.Write(el.tag)
	buf.WriteRune(RightCarrot)
	buf.Write(el.tail)
}
