package dmsl

const eof = -1

type TokenType int

const (
	TokenElement TokenType = iota
	TokenHashTag
	TokenHashId
	TokenHashClass
	TokenAttrKey
	TokenAttrValue
	TokenText
	TokenTextWs
	TokenComment
	TokenFilterStart
	TokenFilterName
	TokenFilterArgs
	TokenFilterContent
	TokenFilterContentWs
	TokenFilterEnd
	TokenAction
	TokenEOF
)

type Token struct {
	typ   TokenType
	start int
	end   int
}

var TokenString = map[TokenType]string{
	TokenElement:         "Element",
	TokenHashTag:         "HashTag",
	TokenHashId:          "HashId",
	TokenHashClass:       "HashClass",
	TokenAttrKey:         "AttrKey",
	TokenAttrValue:       "AttrValue",
	TokenText:            "Text",
	TokenTextWs:          "TextWs",
	TokenComment:         "Comment",
	TokenFilterStart:     "FilterStart",
	TokenFilterName:      "FilterName",
	TokenFilterArgs:      "FilterArgs",
	TokenFilterContent:   "FilterContent",
	TokenFilterContentWs: "FilterContentWs",
	TokenFilterEnd:       "FilterEnd",
	TokenAction:          "Action",
	TokenEOF:             "EOF",
}

type TokenReceiver interface {
	ReceiveToken(Token)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	bytes []byte
	state stateFn
	pos   int
	start int

	// In same cases, we want to save the current start position and restore it later.
	// For simple cases, this can be easier then spinning off a second lexer to handle
	// a specific set of input. See lexer.save() and lexer.restore()
	saveStart int

	ident    int
	receiver TokenReceiver
}

func NewLexer(receiver TokenReceiver) *lexer {
	l := new(lexer)
	l.receiver = receiver
	l.state = lexWhiteSpace
	return l
}

func (l *lexer) Run() {
	for l.state != nil {
		l.state = l.state(l)
	}
}

func (l *lexer) emit(t TokenType) {
	l.receiver.ReceiveToken(Token{typ: t, start: l.start, end: l.pos})
}

func (l *lexer) next() {
	l.pos++
}

func (l *lexer) reset() {
	l.start = l.pos
}

func (l *lexer) save() {
	l.saveStart = l.start
}

func (l *lexer) restore() {
	l.start = l.saveStart
}

func (l *lexer) rune() rune {
	if l.pos >= len(l.bytes) {
		return eof
	}
	return rune(l.bytes[l.pos])
}

func (l *lexer) discard() {
	l.pos++
	l.start = l.pos
}

func (l *lexer) saveIdent() {
	// TODO error if l.ident != -1
	l.ident = (l.pos - l.start) * 2
}

func (l *lexer) discardIdent() {
	l.ident = -1
}

// These are the lexer states that will execute on each iteration based on which lexer.state is set to.

func lexWhiteSpace(l *lexer) stateFn {
	for {
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
			l.saveIdent()
			l.emit(TokenFilterStart)
			l.discard()
			return lexFilter
		case '%', '#', '.', '!':
			l.emit(TokenElement)
			return lexHash
		case eof:
			return nil
		default:
			l.emit(TokenTextWs)
			if l.rune() == '\\' {
				l.discard()
			}
			if l.rune() == '`' {
				l.discard()
				l.reset()
				return lexTextEscape
			}

			l.reset()
			return lexText
		}
	}

	panic("unreachable")
}

func lexComment(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '\n':
			l.discard()
			return lexWhiteSpace
		case eof:
			return nil
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexHash(l *lexer) stateFn {
	for {
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
			l.emit(TokenComment)
			return lexWhiteSpace
		case eof:
			return nil
		default:
			return lexWhiteSpace
		}
	}

	panic("unreachable")
}

func lexHashTag(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '#', '.', '[', ' ', '\t', '\n', eof:
			l.emit(TokenHashTag)
			return lexHash
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexHashId(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '#', '.', '[', ' ', '\t', '\n', eof: // dup id will throw error later for strict rule enforcement
			l.emit(TokenHashId)
			return lexHash
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexHashClass(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '#', '.', '[', ' ', '\t', '\n', eof:
			l.emit(TokenHashClass)
			return lexHash
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexAttributeKey(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '=':
			l.emit(TokenAttrKey)
			l.discard()
			return lexAttributeValue
		case ']':
			l.emit(TokenAttrKey)
			l.discard()
			return lexWhiteSpace
		case eof:
			return nil // TODO emit error
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexAttributeValue(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '\\':
			l.next()
			// skip next input
			l.next()
			break
		case ']':
			l.emit(TokenAttrValue)
			l.discard()
			return lexWhiteSpace
		case eof:
			return nil // TODO handle error
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexTextEscape(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '`':
			l.emit(TokenText)
			l.discard()
			return lexWhiteSpace
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexText(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '\n':
			l.emit(TokenText)
			l.discard()
			return lexWhiteSpace
		case '{': // TODO use correct delimiters!!!!
			l.save()
			l.reset()
			return lexAction
		case eof:
			l.emit(TokenText)
			return nil
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexAction(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '}':
			l.emit(TokenAction)
			l.restore()
			return lexText
		case eof:
			return nil
		default:
			l.next()
		}
	}

	panic("unreachable")
}

// lexFilter stands alone for parsing, not mingling with lexWhiteSpace until
// it's completely finished.
func lexFilter(l *lexer) stateFn {
	for {
		switch l.rune() {
		case ' ', '\t':
			l.emit(TokenFilterName)
			l.discard()
			return lexFilterArgs
		case '\n':
			l.emit(TokenFilterName)
			l.discard()
			return lexFilterWhiteSpace
		case eof:
			return nil
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexFilterArgs(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '\n':
			l.emit(TokenFilterArgs)
			l.discard()
			return lexFilterWhiteSpace
		case eof:
			return nil
		default:
			l.next()
		}
	}

	panic("unreachable")
}

func lexFilterWhiteSpace(l *lexer) stateFn {
	for {
		switch l.rune() {
		case ' ', '\t':
			l.next()
			break
		case '\n':
			l.emit(TokenFilterContent)
			l.discard()
			break
		default:
			if (l.pos-l.start)*2 <= l.ident {
				l.emit(TokenFilterEnd)
				l.discardIdent() // discards previously saved ident level
				// dont reset position so lexWhiteSpace is aware of current ident level
				return lexWhiteSpace
			}
			l.emit(TokenFilterContentWs)
			return lexFilterContent
		}
	}

	panic("unreachable")
}

func lexFilterContent(l *lexer) stateFn {
	for {
		switch l.rune() {
		case '\n':
			l.emit(TokenFilterContent)
			l.discard()
			return lexFilterWhiteSpace
		default:
			l.next()
			break
		}
	}

	panic("unreachable")
}
