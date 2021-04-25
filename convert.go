package main

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	start := time.Now()

	f, err := os.Open("JMdict_e.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	if err := os.Remove("test.db"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	}

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Entity{}, &Entry{}, &Kanji{}, &KanjiInfo{}, &KanjiPriority{},
		&Reading{}, &ReadingRestriction{}, &ReadingInfo{}, &ReadingPriority{},
		&Sense{}, &SenseKanjiRestriction{}, &SenseReadingRestriction{},
		&SenseCrossReference{}, &SenseAntonym{}, &SensePartOfSpeech{},
		&SenseFieldOfApplication{}, &SenseMiscInfo{}, &SenseSourceLanguage{},
		&SenseDialect{}, &SenseGloss{}, &SenseInfo{})

	dec := xml.NewDecoder(r)

	curr := 0 // TODO: used while dev, remove later

	var batch []Entry

	for curr < 20000 {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		switch t := tok.(type) {
		case xml.Directive:
			if bytes.HasPrefix(t, []byte("DOCTYPE")) {
				// Document type definition (DTD)
				// Create a dummy entity map for the decoder, which maps the entity name
				// to the name itself, e.g. {"key1": "key1", "key2": "key2"}.
				// This is so the decoder can interpret those entities, and store them
				// by name when unmarshaling.
				// Also create an entity slice for GORM to store in DB
				m := make(map[string]string)
				var s []Entity
				for k, v := range entityBytesToMap(t) {
					m[k] = k
					s = append(s, Entity{k, v})
				}

				dec.Entity = m
				db.CreateInBatches(s, len(s))
			}
		case xml.StartElement:
			if t.Name.Local == "entry" {
				curr++
				var e Entry
				dec.DecodeElement(&e, &t)

				batch = append(batch, e)

				if len(batch) == 500 {
					db.CreateInBatches(batch, 500)
					batch = nil
					// TODO: the last batch (remaining)
				}
			}
		}
	}

	log.Printf("Done. Took %s", time.Since(start))
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
