package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

var rEntity = regexp.MustCompile(`<!ENTITY\s+(\S+)\s+"(.+)">`)

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
				for i, v := range rEntity.FindAllSubmatch(el, -1) {
					m[string(v[1])] = string(v[1])
					fmt.Printf("%d %s %q\n", i+1, string(v[1]), string(v[2]))
				}
				d.Entity = m
			}
		}
	}
}
