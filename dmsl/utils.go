package dmsl

import (
	"io/ioutil"
	"log"
	"path/filepath"
)

func Open(filename string, dir string) []byte {
	b, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		log.Fatal(err)
	}
	return b
}

// CountWs is only called for appropriate emitted tokens that are known to be
// whitespace. *2 is to account for inlined tags that occupy ws+1
func CountWs(t Token) int {
	return (t.end - t.start) * 2
}
