package dmsl

import "bytes"

const (
	LeftCarrot  = '<'
	Slash       = '/'
	RightCarrot = '>'
	Space       = ' '
	Equal       = '='
	Quote       = '"'
	Exclamation = '!'
	Hyphen      = '-'
)

type Elem struct {
	parent     *Elem
	children   []*Elem
	tag        []byte
	id         []byte
	class      [][]byte
	attr       [][][]byte
	text       []byte
	tail       []byte
	actionEnds int
	// TODO use Comment struct
	isComment bool
}

type Comment struct {
	Elem
}

func (el *Comment) ToString(buf *bytes.Buffer) {
	buf.WriteRune(LeftCarrot)
	buf.WriteRune(Exclamation)
	buf.WriteRune(Hyphen)
	buf.WriteRune(Hyphen)

	for _, child := range el.children {
		child.ToString(buf)
	}

	buf.WriteRune(Hyphen)
	buf.WriteRune(Hyphen)
	buf.WriteRune(RightCarrot)
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

	// TODO use Comment struct in parser
	if el.isComment {
		buf.WriteRune(LeftCarrot)
		buf.WriteRune(Exclamation)
		buf.WriteRune(Hyphen)
		buf.WriteRune(Hyphen)

		buf.Write(el.text)

		for _, child := range el.children {
			child.ToString(buf)
		}

		buf.WriteRune(Hyphen)
		buf.WriteRune(Hyphen)
		buf.WriteRune(RightCarrot)
		return
	}
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

	for i := 0; i < el.actionEnds; i++ {
		buf.WriteString("{end}")
	}
	buf.WriteRune(LeftCarrot)
	buf.WriteRune(Slash)
	buf.Write(el.tag)
	buf.WriteRune(RightCarrot)
	buf.Write(el.tail)
}
