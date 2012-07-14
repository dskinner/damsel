package main

import (
	"fmt"
	"github.com/dskinner/damsel/dmsl"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

var TestsDir = "tests"

func get_html(t *testing.T, s string) string {
	b, err := ioutil.ReadFile(filepath.Join(TestsDir, s+".html"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func test(t *testing.T, s string, data interface{}) {
	fmt.Println("testing:", s)
	dmsl.TemplateDir = TestsDir
	html := get_html(t, s)

	tmpl, _ := dmsl.ParseFile(s + ".dmsl")
	r, err := tmpl.Execute(data)
	r = strings.TrimSpace(r)
	if r != html {
		fmt.Println("\nExpected\n========\n", html, "\nReceived\n========\n", r, "\n\n")
		t.Fatal("parse failed:", s, "\n", err)
	}
}

func Test_html(t *testing.T) {
	test(t, "html", nil)
}

func Test_indent(t *testing.T) {
	test(t, "indent", nil)
}

func Test_variable_indent(t *testing.T) {
	test(t, "variable_indent", nil)
}

func Test_inline(t *testing.T) {
	data := []string{"a", "b", "c", "d"}
	test(t, "inline", data)
}

func Test_multiline_text(t *testing.T) {
	test(t, "multiline_text", nil)
}

func Test_ul_range_header(t *testing.T) {
	data := []string{"a", "b", "c", "d"}
	test(t, "ul_range_header", data)
}

func Test_tabs(t *testing.T) {
	test(t, "tabs", nil)
}

func Test_tag_hashes(t *testing.T) {
	test(t, "tag_hashes", nil)
}

func Test_html_comment(t *testing.T) {
	test(t, "hcomment", nil)
}

func Test_extends(t *testing.T) {
	test(t, "extends", nil)
}

func Test_extends_super(t *testing.T) {
	test(t, "extends_super", nil)
}

func Test_big_table(t *testing.T) {
	table := [2][10]int{}
	test(t, "bigtable", table)
}

func Benchmark_parser(b *testing.B) {
	b.StopTimer()
	bytes, err := ioutil.ReadFile(filepath.Join(TestsDir, "bigtable2.dmsl"))
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		dmsl.DocParse(bytes)
	}
}

func Benchmark_bigtable_go_stdlib(b *testing.B) {
	b.StopTimer()
	bytes, err := ioutil.ReadFile(filepath.Join(TestsDir, "bigtable.dmsl"))
	if err != nil {
		b.Fatal(err)
	}
	table := [1000][10]int{}
	tmpl, _ := dmsl.Parse(bytes)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tmpl.Execute(table)
	}
}

func Benchmark_bigtable(b *testing.B) {
	b.StopTimer()
	bytes, err := ioutil.ReadFile(filepath.Join(TestsDir, "bigtable.dmsl"))
	if err != nil {
		b.Fatal(err)
	}
	table := [1000][10]int{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tpl, _ := dmsl.Parse(bytes)
		tpl.Execute(table)
	}
}

func Benchmark_bigtable2(b *testing.B) {
	b.StopTimer()
	bytes, err := ioutil.ReadFile(filepath.Join(TestsDir, "bigtable2.dmsl"))
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tpl, _ := dmsl.Parse(bytes)
		tpl.Execute(nil)
	}
}
