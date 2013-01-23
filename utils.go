package damsel

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
