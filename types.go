package main

import "database/sql/driver"

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
	EntrySeq uint            `gorm:"notNull"`
	Element  string          `xml:"keb" gorm:"notNull"`
	Info     []KanjiInfo     `xml:"ke_inf"`
	Priority []KanjiPriority `xml:"ke_pri"`
}

type KanjiInfo struct {
	ID      uint
	KanjiID uint   `gorm:"notNull"`
	Code    string `xml:",chardata" gorm:"notNull"`
}

type KanjiPriority struct {
	ID      uint
	KanjiID uint   `gorm:"notNull"`
	Code    string `xml:",chardata" gorm:"notNull"`
}

type Reading struct {
	ID          uint
	EntrySeq    uint                 `gorm:"notNull"`
	Element     string               `xml:"reb" gorm:"notNull"`
	NoKanji     *BoolTag             `xml:"re_nokanji" gorm:"notNull"`
	Restriction []ReadingRestriction `xml:"re_restr"`
	Info        []ReadingInfo        `xml:"re_inf"`
	Priority    []ReadingPriority    `xml:"re_pri"`
}

type ReadingRestriction struct {
	ID           uint
	ReadingID    uint   `gorm:"notNull"`
	KanjiElement string `xml:",chardata" gorm:"notNull"`
}

type ReadingInfo struct {
	ID        uint
	ReadingID uint   `gorm:"notNull"`
	Code      string `xml:",chardata" gorm:"notNull"`
}

type ReadingPriority struct {
	ID        uint
	ReadingID uint   `gorm:"notNull"`
	Code      string `xml:",chardata" gorm:"notNull"`
}

type Sense struct {
	ID                 uint
	EntrySeq           uint                      `gorm:"notNull"`
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
	Info               []SenseInfo               `xml:"s_inf"`
}

type SenseKanjiRestriction struct {
	ID           uint
	SenseID      uint   `gorm:"notNull"`
	KanjiElement string `xml:",chardata" gorm:"notNull"`
}

type SenseReadingRestriction struct {
	ID             uint
	SenseID        uint   `gorm:"notNull"`
	ReadingElement string `xml:",chardata" gorm:"notNull"`
}

type SenseCrossReference struct {
	ID      uint
	SenseID uint   `gorm:"notNull"`
	Content string `xml:",chardata" gorm:"notNull"`
}

type SenseAntonym struct {
	ID      uint
	SenseID uint   `gorm:"notNull"`
	Content string `xml:",chardata" gorm:"notNull"`
}

type SensePartOfSpeech struct {
	ID      uint
	SenseID uint   `gorm:"notNull"`
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseFieldOfApplication struct {
	ID      uint
	SenseID uint   `gorm:"notNull"`
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseMiscInfo struct {
	ID      uint
	SenseID uint   `gorm:"notNull"`
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
	SenseID uint   `gorm:"notNull"`
	Code    string `xml:",chardata" gorm:"notNull"`
}

type SenseGloss struct {
	ID      uint
	SenseID uint    `gorm:"notNull"`
	Type    *string `xml:"g_type,attr"`
	Text    string  `xml:",chardata"`
}

type SenseInfo struct {
	ID      uint
	SenseID uint   `gorm:"notNull"`
	Text    string `xml:",chardata" gorm:"notNull"`
}

type BoolTag struct{}

func (t *BoolTag) Value() (driver.Value, error) {
	return t != nil, nil
}

type BoolAttr string

func (a *BoolAttr) Value() (driver.Value, error) {
	return a != nil, nil
}

type NullableString string

func (s NullableString) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}
