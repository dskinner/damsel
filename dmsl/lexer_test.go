package dmsl

import (
	"fmt"
	"io/ioutil"
	"testing"
)

type TokenPrinter struct {
	l *lexer
}

func (t *TokenPrinter) ReceiveToken(tkn Token) {
	fmt.Println(TokenString[tkn.typ])
	if tkn.typ == TokenFilterContentWs {
		fmt.Println(tkn.end - tkn.start)
	} else {
		fmt.Println(string(t.l.bytes[tkn.start:tkn.end]))
	}
}

type TokenNil struct{}

func (t *TokenNil) ReceiveToken(tkn Token) {}

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

func Benchmark_bigtable2(b *testing.B) {
	b.StopTimer()
	bytes, err := ioutil.ReadFile("../tests/bigtable2.dmsl")
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tknNil := new(TokenNil)
		l := NewLexer(tknNil)
		l.bytes = bytes
		l.Run()
	}
}
