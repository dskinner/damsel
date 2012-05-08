package main

import (
	"flag"
	"fmt"
	"github.com/dskinner/damsel/dmsl"
)

func main() {
	var filename string
	var debug bool
	flag.StringVar(&filename, "f", "", "file to parse")
	flag.BoolVar(&debug, "d", false, "print parser debug info")
	flag.Parse()

	if filename == "" {
		flag.Usage()
		return
	}

	if debug {
		/*
			f := strings.Split(parse.Open(filename, ""), "\n")
			s := parse.Pre(f, "")
			for _, l := range s {
				fmt.Println(l)
			}
		*/
		fmt.Println(dmsl.LexerParse(dmsl.Open(filename, ""), "").String())
	} else {
		t := dmsl.ParseFile(filename)
		result, err := t.Execute(nil)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(result)
		}
	}
}
