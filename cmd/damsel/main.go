package main

import (
	"dasa.cc/damsel"
	"dasa.cc/damsel/parse"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

var filename = flag.String("f", "", "file to parse")
var debug = flag.Bool("d", false, "print parser debug info")
var pprint = flag.Bool("pprint", false, "pretty print output")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var data = flag.String("data", "", "json string to decode as data for template")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		parse.DocParse(damsel.Open("tests/bigtable2.dmsl", ""))
		return
	}

	if *filename == "" {
		flag.Usage()
		return
	}

	if *pprint {
		parse.Pprint = true
	}

	if *debug {
		r, err := parse.DocParse(damsel.Open(*filename, ""))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)
	} else {
		t, err := damsel.ParseFile(*filename)

		var d interface{}
		if err = json.Unmarshal([]byte(*data), &d); err != nil {
			if len(*data) > 0 {
				fmt.Println(err.Error())
			}
		}

		result, err := t.Execute(d)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(result)
		}
	}
}
