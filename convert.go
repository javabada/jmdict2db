package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"os"
	"regexp"
)

func main() {
	f, err := os.Open("sample.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := xml.NewDecoder(f)

	m := make(map[string]string)
	m["err"] = "0"

	for {
		t, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		switch el := t.(type) {
		case xml.Directive:
			if bytes.HasPrefix(el, []byte("DOCTYPE")) {
				d.Entity = entitiesToMap(el)
			}
		}
	}
}

var reEntity = regexp.MustCompile(`<!ENTITY\s+(\S+)\s+"(.+)">`)

// entitiesToMap finds DTD entities from a byte slice and puts them in a map
func entitiesToMap(e []byte) map[string]string {
	m := make(map[string]string)
	for _, v := range reEntity.FindAllSubmatch(e, -1) {
		m[string(v[1])] = string(v[2])
	}
	return m
}
