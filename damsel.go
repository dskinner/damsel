package main

import (
	"flag"
	"fmt"
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
		fmt.Println("debug does nothing atm")
	} else {
		t := ParseFile(filename)
		fmt.Println(t.Execute(nil))
	}
}
