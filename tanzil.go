package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// tanzilVerse is a line of a Tanzil text file: "2|255|text".
type tanzilVerse struct {
	Surah, Number int
	Text          string
}

// parseTanzilText reads a Tanzil text or translation file, "surah|ayah|text"
// a line, skipping its notice.
func parseTanzilText(data []byte) ([]tanzilVerse, error) {
	var out []tanzilVerse
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] < '0' || line[0] > '9' {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %q", line)
		}
		s, err1 := strconv.Atoi(parts[0])
		n, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("line %q", line)
		}
		out = append(out, tanzilVerse{s, n, parts[2]})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) != VerseCount {
		return nil, fmt.Errorf("%d verses, want %d", len(out), VerseCount)
	}
	return out, nil
}

// quranMeta is what Quranku takes from Tanzil's quran-data.js.
type quranMeta struct {
	Suras []suraMeta
	// Where each juz, quarter of a hizb, and page of the Madinah Mushaf
	// starts.
	Juz, Quarters, Pages []verseRef
}

type suraMeta struct {
	Verses        int
	Arabic        string
	English       string // its meaning
	Meccan        bool
	Transliterate string
}

type verseRef struct{ Surah, Number int }

func (r verseRef) before(s, n int) bool {
	return r.Surah < s || r.Surah == s && r.Number <= n
}

// parseQuranMeta reads the arrays of quran-data.js.
func parseQuranMeta(js []byte) (*quranMeta, error) {
	src := string(js)
	array := func(name string) ([][]string, error) {
		at := strings.Index(src, "QuranData."+name+" = [")
		if at < 0 {
			return nil, fmt.Errorf("no %s", name)
		}
		body := src[at+len("QuranData."+name+" = ["):]
		end := strings.Index(body, "];")
		if end < 0 {
			return nil, fmt.Errorf("%s: no end", name)
		}
		// Rows are [...], one or several a line, after comments.
		var rows [][]string
		for _, line := range strings.Split(body[:end], "\n") {
			if c := strings.Index(line, "//"); c >= 0 {
				line = line[:c]
			}
			for _, m := range jsRow.FindAllStringSubmatch(line, -1) {
				if row := jsFields(m[1]); len(row) > 0 {
					rows = append(rows, row)
				}
			}
		}
		return rows, nil
	}
	refs := func(name string, want int) ([]verseRef, error) {
		rows, err := array(name)
		if err != nil {
			return nil, err
		}
		var out []verseRef
		for _, r := range rows {
			if len(r) < 2 {
				return nil, fmt.Errorf("%s: row %v", name, r)
			}
			s, err1 := strconv.Atoi(r[0])
			n, err2 := strconv.Atoi(r[1])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("%s: row %v", name, r)
			}
			out = append(out, verseRef{s, n})
		}
		// Tanzil ends some lists with the end of the Quran, [115, 1].
		if len(out) == want+1 && out[want].Surah == 115 {
			out = out[:want]
		}
		if len(out) != want {
			return nil, fmt.Errorf("%s: %d entries, want %d", name, len(out), want)
		}
		return out, nil
	}
	m := &quranMeta{}
	suras, err := array("Sura")
	if err != nil {
		return nil, err
	}
	for _, r := range suras {
		if len(r) != 8 {
			continue // the end marker
		}
		n, err := strconv.Atoi(r[1])
		if err != nil {
			return nil, fmt.Errorf("sura %v", r)
		}
		m.Suras = append(m.Suras, suraMeta{Verses: n, Arabic: r[4], Transliterate: r[5], English: r[6], Meccan: r[7] == "Meccan"})
	}
	if len(m.Suras) != 114 {
		return nil, fmt.Errorf("%d suras", len(m.Suras))
	}
	if m.Juz, err = refs("Juz", 30); err != nil {
		return nil, err
	}
	if m.Quarters, err = refs("HizbQaurter", 240); err != nil {
		return nil, err
	}
	if m.Pages, err = refs("Page", PageCount); err != nil {
		return nil, err
	}
	return m, nil
}

var jsRow = regexp.MustCompile(`\[([^\[\]]*)\]`)

// jsFields splits the fields of a JavaScript array literal's row: numbers,
// and strings in single or double quotes.
func jsFields(s string) []string {
	var out []string
	for {
		s = strings.TrimLeft(s, " \t,")
		if s == "" || strings.HasPrefix(s, "//") {
			return out
		}
		if q := s[0]; q == '\'' || q == '"' {
			end := strings.IndexByte(s[1:], q)
			if end < 0 {
				return append(out, s[1:])
			}
			out = append(out, s[1:1+end])
			s = s[end+2:]
			continue
		}
		end := strings.IndexByte(s, ',')
		if end < 0 {
			return append(out, strings.TrimSpace(s))
		}
		out = append(out, strings.TrimSpace(s[:end]))
		s = s[end+1:]
	}
}
