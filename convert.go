package main

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const url = "http://ftp.edrdg.org/pub/Nihongo/JMdict_e.gz"
const filePath = "JMdict_e.gz"
const dbPath = "jmdict.db"
const batchSize = 500

func main() {
	timer := time.Now()

	if err := dlFile(url, filePath); err != nil {
		log.Fatal(err)
	}

	log.Printf("Download took: %s", time.Since(timer))
	timer = time.Now()

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	if err := os.Remove(dbPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
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

	var batch []Entry
	var done chan bool

	for {
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
				// Also create an entity slice for GORM to store in DB.
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
				// JMdict entry
				// Unmarshal entries and batch insert them into the DB.
				// Handle writes in a new goroutine. Wait for the current batch to
				// finish before handling the next batch to avoid concurrent writes.
				var e Entry
				dec.DecodeElement(&e, &t)
				batch = append(batch, e)

				if len(batch) == batchSize {
					if done != nil {
						<-done
					}
					done = make(chan bool)
					go insertBatch(db, batch, done)
					batch = nil
				}
			}
		}
	}

	// Handle final batch
	<-done
	insertBatch(db, batch, nil)

	log.Printf("Conversion took: %s", time.Since(timer))
}

func dlFile(url, out string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func insertBatch(db *gorm.DB, b []Entry, done chan bool) {
	db.CreateInBatches(b, len(b))
	if done != nil {
		done <- true
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
