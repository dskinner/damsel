package main

import (
	"flag"
	"fmt"
	"github.com/dskinner/damsel/dmsl"

// 	"os"
// 	"log"
// 	"runtime/pprof"
)

var filename = flag.String("f", "", "file to parse")
var debug = flag.Bool("d", false, "print parser debug info")
var pprint = flag.Bool("pprint", false, "pretty print output")

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()

	/*
		if *cpuprofile != "" {
			f, err := os.Create(*cpuprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		dmsl.ParserParse(dmsl.Open("tests/bigtable2.dmsl", ""))
	*/

	if *filename == "" {
		flag.Usage()
		return
	}

	if *pprint {
		dmsl.Pprint = true
	}

	if *debug {
		fmt.Println(dmsl.ParserParse(dmsl.Open(*filename, "")))
	} else {
		t := dmsl.ParseFile(*filename)
		result, err := t.Execute(nil)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(result)
		}
	}
}
