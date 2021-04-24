package main

import (
	"bytes"
	"database/sql/driver"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"os"
	"regexp"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Entity struct {
	Name  string `gorm:"primaryKey"`
	Value string `gorm:"notNull"`
}

type Entry struct {
	Seq     uint      `xml:"ent_seq" gorm:"primaryKey"`
	Kanji   []Kanji   `xml:"k_ele"`
	Reading []Reading `xml:"r_ele"`
	Sense   []Sense   `xml:"sense"`
}

type Kanji struct {
	ID       uint
	EntrySeq uint
	Element  string          `xml:"keb" gorm:"notNull"`
	Info     []KanjiInfo     `xml:"ke_inf"`
	Priority []KanjiPriority `xml:"ke_pri"`
}

type KanjiInfo struct {
	ID      uint
	KanjiID uint
	Code    string `xml:",chardata" gorm:"notNull"`
}

type KanjiPriority struct {
	ID      uint
	KanjiID uint
	Code    string `xml:",chardata" gorm:"notNull"`
}

type Reading struct {
	ID          uint
	EntrySeq    uint
	Element     string               `xml:"reb" gorm:"notNull"`
	NoKanji     *Exist               `xml:"re_nokanji" gorm:"notNull"`
	Restriction []ReadingRestriction `xml:"re_restr"`
	Info        []ReadingInfo        `xml:"re_inf"`
	Priority    []ReadingPriority    `xml:"re_pri"`
}

type ReadingRestriction struct {
	ID           uint
	ReadingID    uint
	KanjiElement string `xml:",chardata" gorm:"notNull"`
}

type ReadingInfo struct {
	ID        uint
	ReadingID uint
	Code      string `xml:",chardata" gorm:"notNull"`
}

type ReadingPriority struct {
	ID        uint
	ReadingID uint
	Code      string `xml:",chardata" gorm:"notNull"`
}

type Sense struct {
	ID                 uint
	EntrySeq           uint
	KanjiRestriction   []SenseKanjiRestriction   `xml:"stagk"`
	ReadingRestriction []SenseReadingRestriction `xml:"stagr"`
}

type SenseKanjiRestriction struct {
	ID           uint
	SenseID      uint
	KanjiElement string `xml:",chardata" gorm:"notNull"`
}

type SenseReadingRestriction struct {
	ID             uint
	SenseID        uint
	ReadingElement string `xml:",chardata" gorm:"notNull"`
}

type Exist struct{}

func (e *Exist) Value() (driver.Value, error) {
	return e != nil, nil
}

func main() {
	f, err := os.Open("sample.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := os.Remove("test.db"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	}

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Entity{})
	db.AutoMigrate(&Entry{})
	db.AutoMigrate(&Kanji{})
	db.AutoMigrate(&KanjiInfo{})
	db.AutoMigrate(&KanjiPriority{})
	db.AutoMigrate(&Reading{})
	db.AutoMigrate(&ReadingRestriction{})
	db.AutoMigrate(&ReadingInfo{})
	db.AutoMigrate(&ReadingPriority{})
	db.AutoMigrate(&Sense{})
	db.AutoMigrate(&SenseKanjiRestriction{})
	db.AutoMigrate(&SenseReadingRestriction{})

	dec := xml.NewDecoder(f)

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
				var e Entry
				dec.DecodeElement(&e, &t)
				db.Create(e)
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
