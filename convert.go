package main

import (
	"bytes"
	"compress/gzip"
	"database/sql/driver"
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
	NoKanji     *BoolTag             `xml:"re_nokanji" gorm:"notNull"`
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
	CrossReference     []SenseCrossReference     `xml:"xref"`
	Antonym            []SenseAntonym            `xml:"ant"`
	PartOfSpeech       []SensePartOfSpeech       `xml:"pos"`
	FieldOfApplication []SenseFieldOfApplication `xml:"field"`
	MiscInfo           []SenseMiscInfo           `xml:"misc"`
	SourceLanguage     []SenseSourceLanguage     `xml:"lsource"`
	Dialect            []SenseDialect            `xml:"dial"`
	Gloss              []SenseGloss              `xml:"gloss"`
	MoreInfo           []SenseMoreInfo           `xml:"s_inf"`
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

type SenseCrossReference struct {
	ID      uint
	SenseID uint
	Content string `xml:",chardata" gorm:"notNull"`
}

type SenseAntonym struct {
	ID      uint
	SenseID uint
	Element string `xml:",chardata" gorm:"notNull"`
}

type SensePartOfSpeech struct {
	ID      uint
	SenseID uint
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseFieldOfApplication struct {
	ID      uint
	SenseID uint
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseMiscInfo struct {
	ID      uint
	SenseID uint
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseSourceLanguage struct {
	ID       uint
	SenseID  uint           `gorm:"notNull"`
	Language string         `xml:"lang,attr" gorm:"notNull;default:eng"`
	Partial  *BoolAttr      `xml:"ls_type,attr" gorm:"notNull"`
	Wasei    *BoolAttr      `xml:"ls_wasei,attr" gorm:"notNull"`
	Text     NullableString `xml:",chardata"`
}

type SenseDialect struct {
	ID      uint
	SenseID uint
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseGloss struct {
	ID      uint
	SenseID uint
	Type    *string `xml:"g_type,attr"`
	Text    string  `xml:",chardata"`
}

type SenseMoreInfo struct {
	ID         uint
	SenseID    uint
	EntityCode string `xml:",chardata" gorm:"notNull"`
}

type BoolTag struct{}

func (s *BoolTag) Value() (driver.Value, error) {
	return s != nil, nil
}

type BoolAttr string

func (s *BoolAttr) Value() (driver.Value, error) {
	return s != nil, nil
}

type NullableString string

func (s NullableString) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}

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
	db.AutoMigrate(&SenseCrossReference{})
	db.AutoMigrate(&SenseAntonym{})
	db.AutoMigrate(&SensePartOfSpeech{})
	db.AutoMigrate(&SenseFieldOfApplication{})
	db.AutoMigrate(&SenseMiscInfo{})
	db.AutoMigrate(&SenseSourceLanguage{})
	db.AutoMigrate(&SenseDialect{})
	db.AutoMigrate(&SenseGloss{})
	db.AutoMigrate(&SenseMoreInfo{})

	dec := xml.NewDecoder(r)

	curr := 0

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

	elapsed := time.Since(start)
	log.Printf("Took %s", elapsed)
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
