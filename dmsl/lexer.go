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
	TokenFilter
	TokenFilterName
	TokenFilterArgs
	TokenFilterContent
	TokenFilterContentDefault
	TokenAction
	TokenEOF
)

type Token struct {
	typ   TokenType
	start int
	end   int
}

var TokenString = map[TokenType]string{
	TokenElement:              "Element",
	TokenHashTag:              "HashTag",
	TokenHashId:               "HashId",
	TokenHashClass:            "HashClass",
	TokenAttrKey:              "AttrKey",
	TokenAttrValue:            "AttrValue",
	TokenText:                 "Text",
	TokenTextWs:               "TextWs",
	TokenComment:              "Comment",
	TokenFilter:               "Filter",
	TokenFilterName:           "FilterName",
	TokenFilterArgs:           "FilterArgs",
	TokenFilterContent:        "FilterContent",
	TokenFilterContentDefault: "FilterContentDefault",
	TokenAction:               "Action",
	TokenEOF:                  "EOF",
}

type stateFn func(*Lexer) stateFn

type Lexer struct {
	bytes      []byte
	state      stateFn
	pos        int
	start      int
	filterDone bool // TODO ditch the need for this
	tokens     chan Token
	parser     *Parser
}

func (l *Lexer) Run() {
	for l.state != nil {
		l.state = l.state(l)
	}
	l.tokens <- Token{typ: TokenEOF, start: 0, end: 0}
}

func (l *Lexer) emit(t TokenType) {
	//l.tokens <- Token{typ: t, start: l.start, end: l.pos}
	l.parser.receiveTokens(Token{typ: t, start: l.start, end: l.pos}, l)
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

// These are the lexer states that will execute on each iteration based on which is set to Lexer.state

func lexWhiteSpace(l *Lexer) stateFn {
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
			l.emit(TokenFilter)
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
			l.reset()
			return lexText
		}
	}
	return lexWhiteSpace
}

func lexComment(l *Lexer) stateFn {
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

	return lexComment
}

func lexHash(l *Lexer) stateFn {
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
		default:
			return lexWhiteSpace
		}
	}
	return lexHash
}

func lexHashTag(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '#', '.', '[', ' ', '\t', '\n', eof:
			l.emit(TokenHashTag)
			return lexHash
		default:
			l.next()
		}
	}
	return lexHashTag
}

func lexHashId(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '#', '.', '[', ' ', '\t', '\n', eof: // TODO technically there should never be multiple ids, and damsel should be strict regarding this due to extends
			l.emit(TokenHashId)
			return lexHash
		default:
			l.next()
		}
	}
	return lexHashId
}

func lexHashClass(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '#', '.', '[', ' ', '\t', '\n', eof:
			l.emit(TokenHashClass)
			return lexHash
		default:
			l.next()
		}
	}
	return lexHashClass
}

func lexAttributeKey(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '=':
			l.emit(TokenAttrKey)
			l.discard()

			// check for literal value
			switch l.rune() {
			case '\'', '"':
				l.next()
				return lexAttributeValueLiteral
			}

			return lexAttributeValue
		case ']':
			l.emit(TokenAttrKey)
			l.discard()
			return lexWhiteSpace
		default:
			l.next()
		}
	}
	return lexAttributeKey
}

func lexAttributeValueLiteral(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case rune(l.bytes[l.start]):
			if l.bytes[l.pos-1] == '\\' { // check for escaped quote
				l.next()
				break
			} else {
				return lexAttributeValue
			}
		case eof:
			return nil // TODO handle error
		default:
			l.next()
		}
	}
	return lexAttributeValueLiteral
}

func lexAttributeValue(l *Lexer) stateFn {
	for {
		switch l.rune() {
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
	return lexAttributeValue
}

func lexText(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '\n':
			l.emit(TokenText)
			l.discard()
			return lexWhiteSpace
		case '{': // TODO use correct delimiters!!!!
			l.reset() // TODO check on this, originally lexAction just appended each byte along the way
			return lexAction
		case eof:
			l.emit(TokenText)
			return nil
		default:
			l.next()
		}
	}
	return lexText
}

func lexAction(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '}':
			l.emit(TokenAction)
			return lexText
		case eof:
			return nil
		default:
			l.next()
		}
	}
	return lexAction
}

// lexFilter stands alone for parsing, not mingling with lexWhiteSpace until
// it's completely finished.
func lexFilter(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case ' ', '\t':
			l.emit(TokenFilterName)
			l.discard()
			return lexFilterArgs
		case '\n':
			l.emit(TokenFilterName)
			l.discard()
			return lexFilterContent
		case eof:
			return nil
		default:
			l.next()
		}
	}
	return lexFilter
}

func lexFilterArgs(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case '\n':
			l.emit(TokenFilterArgs)
			l.discard()
			return lexFilterContent
		case eof:
			return nil
		default:
			l.next()
		}
	}
	return lexFilterArgs
}

func lexFilterWhiteSpace(l *Lexer) stateFn {
	for {
		switch l.rune() {
		case ' ', '\t':
			l.next()
		case '\n':
			break
		}
	}
	return lexFilterWhiteSpace
}

func lexFilterContent(l *Lexer) stateFn {
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
			// TODO use lexFilterWhiteSpace so we aren't checking p.filter.contentWs ==0 all the time, refer to Parser
			l.emit(TokenFilterContentDefault)

			// since a filter can cause additional content to be inserted into what we are reading from, wait to hear from Parser that it's finished
			//runtime.Gosched()

			if l.filterDone {
				l.filterDone = false
				return lexWhiteSpace
			}

			l.next()
		}
	}
	
	return lexFilterContent
}
