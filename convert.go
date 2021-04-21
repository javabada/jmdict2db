package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"os"
	"regexp"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Entity struct {
	Name  string `gorm:"primaryKey;notNull"`
	Value string `gorm:"notNull"`
}

func main() {
	f, err := os.Open("sample.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Entity{})

	d := xml.NewDecoder(f)

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

				test := entityMapToSlice(d.Entity)

				db.CreateInBatches(test, len(test))
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

func entityMapToSlice(m map[string]string) []Entity {
	s := make([]Entity, 0)
	for k, v := range m {
		s = append(s, Entity{Name: k, Value: v})
	}
	return s
}
