package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"dasa.cc/damsel/dmsl"
	"dasa.cc/damsel/dmsl/parse"
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

		parse.DocParse(dmsl.Open("tests/bigtable2.dmsl", ""))
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
		r, err := parse.DocParse(dmsl.Open(*filename, ""))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)
	} else {
		t, err := dmsl.ParseFile(*filename)

		var d interface{}
		if err = json.Unmarshal([]byte(*data), &d); err != nil {
			fmt.Println(err.Error())
		}

		result, err := t.Execute(d.([]interface{}))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(result)
		}
	}
}
