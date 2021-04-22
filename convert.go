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
				d.Entity = entityBytesToMap(el)

				var s []Entity
				for k, v := range d.Entity {
					s = append(s, Entity{k, v})
				}

				db.CreateInBatches(s, len(s))
			}
		}
	}
}

var reEntity = regexp.MustCompile(`<!ENTITY\s+(\S+)\s+"([^"]+)">`)

// entityBytesToMap finds DTD entities from a byte slice and returns a map of
// those entities. This also filters out duplicates.
func entityBytesToMap(b []byte) map[string]string {
	m := make(map[string]string)
	for _, v := range reEntity.FindAllSubmatch(b, -1) {
		m[string(v[1])] = string(v[2])
	}
	return m
}
