package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"dasa.cc/damsel"
	"dasa.cc/damsel/parse"
)

var (
	filename   = flag.String("f", "", "file to parse")
	debug      = flag.Bool("d", false, "print parser debug info")
	pprint     = flag.Bool("pprint", false, "pretty print output")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	data       = flag.String("data", "", "json string to decode as data for template")
	html       = flag.Bool("html", false, "parses template with html/template pkg; if unset, will be true if data is set")
)

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

		if *html || len(*data) > 0 {
			r, err := damsel.NewHtmlTemplate(t).Execute(d)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(r)
			}
		} else {
			result, err := t.Result()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(result)
			}
		}
	}
}
