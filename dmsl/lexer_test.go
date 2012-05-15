package dmsl

import (
	"testing"
	"fmt"
)

type TokenPrinter struct {
	l *lexer
}

func (t *TokenPrinter) ReceiveToken(tkn Token) {
	fmt.Println(TokenString[tkn.typ])
	if tkn.typ == TokenFilterContentWs {
		fmt.Println(tkn.end-tkn.start)
	} else {
		fmt.Println(string(t.l.bytes[tkn.start:tkn.end]))
	}
}

func Test_lexer1(t *testing.T) {
	s := `
%html
	%head
		:css /css/
			main.css
			extra.css
			
	%body
		%h1 Hello, World
`
	tknPrinter := new(TokenPrinter)
	l := NewLexer(tknPrinter)
	l.bytes = []byte(s)
	tknPrinter.l = l
	l.Run()
}