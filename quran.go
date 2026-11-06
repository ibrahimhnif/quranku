package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// The Quran text, Tanzil's Uthmani text (data/quran-uthmani.txt), with
// Tanzil's metadata of surahs, juz and Mushaf pages (data/quran-data.js),
// the tajweed rules of every verse (data/tajweed.json, built by
// scripts/build_tajweed.py), and Pickthall's English translation
// (data/en.pickthall.txt). Other translations are downloaded.
var (
	//go:embed data/quran-uthmani.txt
	quranText []byte
	//go:embed data/quran-data.js
	quranDataJS []byte
	//go:embed data/tajweed.json
	tajweedJSON []byte
	//go:embed data/en.pickthall.txt
	pickthallText []byte
)

// Segment is a run of text with at most one tajweed rule.
type Segment struct {
	Text string
	Rule *Rule // nil for plain text
}

// Verse is one ayah.
type Verse struct {
	Key       string // "2:255"
	Index     int    // in the whole Quran, from 0
	Surah     int
	Number    int
	Segments  []Segment
	Words     []Word // its words, as the Mushaf sets them
	Pickthall string // its English translation, built in
	// Where it is in the Mushaf.
	Page, Juz, Hizb int
	EndLine         int // the line of its ayah mark
}

// Surah is one chapter.
type Surah struct {
	ID          int
	Name        string // "Al-Fatihah"
	Arabic      string // "الفاتحة"
	Meaning     T      // "The Opener", "Pembukaan"
	Place       *T     // Meccan or Medinan
	VerseCount  int
	Bismillah   bool // shows the basmalah before the first verse
	Page        int  // the Mushaf page of its header
	Verses      []Verse
	RuleCounts  map[*Rule]int
	searchIndex string
}

// Quran holds every surah.
type Quran struct {
	Surahs   []*Surah
	Basmalah []Segment // 1:1, shown above surahs that start with it
	Pages    []*Page   // the Mushaf's pages, from 1
	All      []*Verse  // every verse, in order
}

// VerseCount is the number of verses in the Quran.
const VerseCount = 6236

// Verse returns verse n (from 1) of surah s (from 1), or nil.
func (q *Quran) Verse(s, n int) *Verse {
	if s < 1 || s > len(q.Surahs) || n < 1 || n > len(q.Surahs[s-1].Verses) {
		return nil
	}
	return &q.Surahs[s-1].Verses[n-1]
}

var (
	footnoteRe = regexp.MustCompile(`<sup[^>]*>.*?</sup>`)
	anyTagRe   = regexp.MustCompile(`<[^>]+>`)
)

// tajweedRange is a rule over text[Start:End], in code points.
type tajweedRange struct {
	Start, End int
	Rule       *Rule
}

// isPauseMark reports whether r is a pause mark (waqf) set above the end
// of a word.
func isPauseMark(r rune) bool { return r >= 0x06D6 && r <= 0x06DB }

// verseSegments splits a verse's text into runs of one tajweed rule. A rule
// starting inside another takes the rest of it. Tanzil sets pause marks
// apart, between spaces; the Mushaf prints them above the word before, so
// they join it here, uncolored.
func verseSegments(text []rune, ranges []tajweedRange) []Segment {
	rules := make([]*Rule, len(text))
	for _, r := range ranges {
		for i := max(r.Start, 0); i < min(r.End, len(text)); i++ {
			rules[i] = r.Rule
		}
	}
	var segs []Segment
	for i := 0; i < len(text); i++ {
		c, rule := text[i], rules[i]
		if c == ' ' && i+1 < len(text) && isPauseMark(text[i+1]) {
			continue
		}
		if isPauseMark(c) {
			rule = nil
		}
		if n := len(segs); n > 0 && segs[n-1].Rule == rule {
			segs[n-1].Text += string(c)
			continue
		}
		segs = append(segs, Segment{Text: string(c), Rule: rule})
	}
	return segs
}

func cleanTranslation(s string) string {
	s = footnoteRe.ReplaceAllString(s, "")
	s = anyTagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(html.UnescapeString(s))
}

// arabicDigits writes n in Arabic-Indic digits.
func arabicDigits(n int) string {
	const digits = "٠١٢٣٤٥٦٧٨٩"
	d := []rune(digits)
	var b []rune
	for _, c := range fmt.Sprint(n) {
		b = append(b, d[c-'0'])
	}
	return string(b)
}

// loadQuran parses the embedded data.
func loadQuran() (*Quran, error) {
	meta, err := parseQuranMeta(quranDataJS)
	if err != nil {
		return nil, fmt.Errorf("quran-data.js: %w", err)
	}
	text, err := parseTanzilText(quranText)
	if err != nil {
		return nil, fmt.Errorf("quran text: %w", err)
	}
	pickthall, err := parseTanzilText(pickthallText)
	if err != nil {
		return nil, fmt.Errorf("pickthall: %w", err)
	}
	var tajweed struct {
		Verses [][][3]any `json:"verses"`
	}
	if err := json.Unmarshal(tajweedJSON, &tajweed); err != nil {
		return nil, fmt.Errorf("tajweed: %w", err)
	}
	if len(tajweed.Verses) != VerseCount {
		return nil, fmt.Errorf("tajweed has %d verses, want %d", len(tajweed.Verses), VerseCount)
	}

	q := &Quran{}
	for i, m := range meta.Suras {
		place := &txtMeccan
		if !m.Meccan {
			place = &txtMedinan
		}
		s := &Surah{
			ID: i + 1, Name: surahNames[i].Name, Arabic: m.Arabic,
			Meaning: T{m.English, surahNames[i].Indonesian}, Place: place,
			VerseCount: m.Verses, Bismillah: i+1 != 1 && i+1 != 9,
			RuleCounts: map[*Rule]int{},
		}
		s.searchIndex = strings.ToLower(fmt.Sprintf("%d %s %s %s %s %s", s.ID, s.Name, m.Transliterate, s.Meaning[0], s.Meaning[1], s.Arabic))
		q.Surahs = append(q.Surahs, s)
	}
	// Tanzil opens verse 1 of every surah but Al-Fatihah and At-Tawbah
	// with the basmalah, which the view sets above the surah.
	basmalah := []rune(text[0].Text + " ")
	for i, v := range text {
		if v.Surah < 1 || v.Surah > len(q.Surahs) || pickthall[i].Surah != v.Surah || pickthall[i].Number != v.Number {
			return nil, fmt.Errorf("verse %d:%d is out of order", v.Surah, v.Number)
		}
		s := q.Surahs[v.Surah-1]
		var ranges []tajweedRange
		for _, r := range tajweed.Verses[i] {
			start, ok1 := r[0].(float64)
			end, ok2 := r[1].(float64)
			key, _ := r[2].(string)
			rule := rulesByKey[key]
			if !ok1 || !ok2 || rule == nil {
				return nil, fmt.Errorf("verse %d:%d: tajweed range %v", v.Surah, v.Number, r)
			}
			ranges = append(ranges, tajweedRange{int(start), int(end), rule})
		}
		runes := []rune(v.Text)
		if v.Number == 1 && s.Bismillah && strings.HasPrefix(v.Text, string(basmalah)) {
			runes = runes[len(basmalah):]
			for k := range ranges {
				ranges[k].Start -= len(basmalah)
				ranges[k].End -= len(basmalah)
			}
		}
		verse := Verse{
			Key: fmt.Sprintf("%d:%d", v.Surah, v.Number), Surah: v.Surah, Number: v.Number,
			Segments:  verseSegments(runes, ranges),
			Pickthall: strings.TrimSpace(pickthall[i].Text),
		}
		for _, seg := range verse.Segments {
			if seg.Rule != nil {
				s.RuleCounts[seg.Rule]++
			}
		}
		s.Verses = append(s.Verses, verse)
	}
	for _, s := range q.Surahs {
		if len(s.Verses) != s.VerseCount {
			return nil, fmt.Errorf("surah %d has %d verses, want %d", s.ID, len(s.Verses), s.VerseCount)
		}
	}
	q.Basmalah = q.Surahs[0].Verses[0].Segments
	for _, s := range q.Surahs {
		for i := range s.Verses {
			s.Verses[i].Index = len(q.All)
			q.All = append(q.All, &s.Verses[i])
		}
	}
	if len(q.All) != VerseCount {
		return nil, fmt.Errorf("%d verses, want %d", len(q.All), VerseCount)
	}
	if err := q.layoutMushaf(meta); err != nil {
		return nil, err
	}
	return q, nil
}

// Search returns the surahs matching the query, by number or name.
func (q *Quran) Search(query string) []*Surah {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return q.Surahs
	}
	var out []*Surah
	for _, s := range q.Surahs {
		if strings.Contains(s.searchIndex, query) || strings.Contains(strings.ReplaceAll(s.searchIndex, "-", ""), query) {
			out = append(out, s)
		}
	}
	return out
}
