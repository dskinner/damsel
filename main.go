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
		fmt.Println(dmsl.ParserParse(dmsl.Open(filename, "")))
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
