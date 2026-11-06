package main

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// The Mushaf follows the pages of the Madinah Mushaf, 604 of them: each
// page starts on the verse the printed one does (from Tanzil's metadata).
// Its 15 lines are Quranku's own: a page's words are spread over them as
// evenly as they go, and the view justifies each line with kashidas.

const (
	PageCount    = 604
	LinesPerPage = 15
)

// LineKind is what a line of a page shows.
type LineKind int

const (
	LineEmpty LineKind = iota
	LineText
	LineHeader   // the surah's name, in a frame
	LineBasmalah // بسم الله الرحمن الرحيم before the surah
)

// PageWord is a word of a verse on a line, or the verse's end mark.
type PageWord struct {
	Verse *Verse
	Word  int // index in Verse.Words, or EndMark
}

// EndMark is PageWord.Word for the ayah number ending a verse.
const EndMark = -1

// Line is one of a page's 15 lines.
type Line struct {
	Kind  LineKind
	Surah *Surah // for headers and basmalahs
	Words []PageWord
}

// Page is one page of the Mushaf.
type Page struct {
	Number int
	Juz    int
	Hizb   int
	Lines  [LinesPerPage]Line
	Surahs []*Surah // the surahs with text on the page, in order
	// RuleCounts counts the segments of each rule on the page.
	RuleCounts map[*Rule]int
}

// Word is a word of a verse, as segments with their tajweed rules.
type Word struct {
	Segments []Segment
	Line     int
}

// splitWords splits a verse's segments into words at its spaces. A rule may
// cover a space (an iqlab from one word to the next); each word then gets
// its part.
func splitWords(segs []Segment) [][]Segment {
	words := [][]Segment{nil}
	for _, s := range segs {
		for k, part := range strings.Split(s.Text, " ") {
			if k > 0 {
				words = append(words, nil)
			}
			if part != "" {
				words[len(words)-1] = append(words[len(words)-1], Segment{Text: part, Rule: s.Rule})
			}
		}
	}
	return words
}

// layoutMushaf places the verses on the pages of the Madinah Mushaf, and
// breaks each page into its lines: a surah's header and basmalah on lines
// of their own, the words on the rest. The first two pages, set apart in
// the print, hold eight lines.
func (q *Quran) layoutMushaf(meta *quranMeta) error {
	q.Pages = make([]*Page, PageCount+1) // 1-based
	for i := range q.Pages[1:] {
		q.Pages[i+1] = &Page{Number: i + 1, RuleCounts: map[*Rule]int{}}
	}
	var byPage [PageCount + 1][]*Verse
	pg, juz, quarter := 0, 0, 0
	for _, v := range q.All {
		for pg < PageCount && meta.Pages[pg].before(v.Surah, v.Number) {
			pg++
		}
		for juz < 30 && meta.Juz[juz].before(v.Surah, v.Number) {
			juz++
		}
		for quarter < 240 && meta.Quarters[quarter].before(v.Surah, v.Number) {
			quarter++
		}
		v.Page, v.Juz, v.Hizb = pg, juz, (quarter+3)/4
		if v.Page < 1 {
			return fmt.Errorf("mushaf: %s is before the first page", v.Key)
		}
		byPage[pg] = append(byPage[pg], v)
	}
	for n := 1; n <= PageCount; n++ {
		if len(byPage[n]) == 0 {
			return fmt.Errorf("mushaf: page %d is empty", n)
		}
		if err := q.Pages[n].fill(q, byPage[n]); err != nil {
			return err
		}
	}
	for _, s := range q.Surahs {
		s.Page = s.Verses[0].Page
	}
	return nil
}

// block is what a page holds in reading order: a surah's header or
// basmalah, a line each, or a run of words.
type block struct {
	kind  LineKind
	surah *Surah
	words []PageWord
}

// fill sets a page's verses on its lines.
func (p *Page) fill(q *Quran, verses []*Verse) error {
	first := verses[0]
	p.Juz, p.Hizb = first.Juz, first.Hizb
	var blocks []block
	text := func() *block {
		if n := len(blocks); n > 0 && blocks[n-1].kind == LineText {
			return &blocks[n-1]
		}
		blocks = append(blocks, block{kind: LineText})
		return &blocks[len(blocks)-1]
	}
	for _, v := range verses {
		s := q.Surahs[v.Surah-1]
		if n := len(p.Surahs); n == 0 || p.Surahs[n-1] != s {
			p.Surahs = append(p.Surahs, s)
		}
		if v.Number == 1 {
			blocks = append(blocks, block{kind: LineHeader, surah: s})
			if s.Bismillah {
				blocks = append(blocks, block{kind: LineBasmalah, surah: s})
			}
		}
		for _, seg := range v.Segments {
			if seg.Rule != nil {
				p.RuleCounts[seg.Rule]++
			}
		}
		v.Words = nil
		b := text()
		for wi, w := range splitWords(v.Segments) {
			v.Words = append(v.Words, Word{Segments: w})
			b.words = append(b.words, PageWord{v, wi})
		}
		b.words = append(b.words, PageWord{v, EndMark})
	}

	lines := LinesPerPage
	if p.Number <= 2 {
		lines = 8
	}
	// Share the lines left by the headers among the runs of words, by
	// their widths.
	free := lines
	var runs []int // indexes of the text blocks
	var total float64
	for i, b := range blocks {
		if b.kind != LineText {
			free--
			continue
		}
		runs = append(runs, i)
		total += runWidth(b.words)
	}
	if free < len(runs) {
		return fmt.Errorf("mushaf: page %d has no room for its words", p.Number)
	}
	share := make([]int, len(blocks))
	left := free
	for k, i := range runs {
		w := runWidth(blocks[i].words)
		n := int(math.Round(float64(free) * w / total))
		if k == len(runs)-1 {
			n = left
		}
		// Each run takes a line at least, and leaves one to each run after.
		n = max(1, min(n, left-(len(runs)-1-k), len(blocks[i].words)))
		share[i] = n
		left -= n
	}

	line := 1
	for i, b := range blocks {
		if b.kind != LineText {
			p.add(line, b.kind, b.surah)
			line++
			continue
		}
		for _, words := range breakLines(b.words, share[i]) {
			for _, w := range words {
				if w.Word == EndMark {
					w.Verse.EndLine = line
				} else {
					w.Verse.Words[w.Word].Line = line
				}
			}
			p.add(line, LineText, nil, words...)
			line++
		}
	}
	return nil
}

// wordWidth estimates how wide a word is, in letters: the marks above and
// below take no room. An ayah's end mark is about two.
func wordWidth(w PageWord) float64 {
	if w.Word == EndMark {
		return 2
	}
	n := 0.0
	for _, s := range w.Verse.Words[w.Word].Segments {
		for _, r := range s.Text {
			if !unicode.Is(unicode.Mn, r) {
				n++
			}
		}
	}
	return n
}

// wordGap is the room between words, in letters.
const wordGap = 1

func runWidth(words []PageWord) float64 {
	w := 0.0
	for i, x := range words {
		if i > 0 {
			w += wordGap
		}
		w += wordWidth(x)
	}
	return w
}

// breakLines breaks words into n lines as even as they go: it minimizes
// the squares of the lines' differences from the mean.
func breakLines(words []PageWord, n int) [][]PageWord {
	k := len(words)
	n = max(1, min(n, k))
	pre := make([]float64, k+1)
	for i, w := range words {
		pre[i+1] = pre[i] + wordWidth(w)
	}
	width := func(i, j int) float64 { return pre[j] - pre[i] + float64(j-i-1)*wordGap }
	target := width(0, k) / float64(n)
	// cost[l][j]: the best breaking of words[:j] into l lines.
	inf := math.Inf(1)
	cost := make([][]float64, n+1)
	from := make([][]int, n+1)
	for l := range cost {
		cost[l] = make([]float64, k+1)
		from[l] = make([]int, k+1)
		for j := range cost[l] {
			cost[l][j] = inf
		}
	}
	cost[0][0] = 0
	for l := 1; l <= n; l++ {
		for j := l; j <= k-(n-l); j++ {
			for i := j - 1; i >= l-1; i-- {
				if cost[l-1][i] == inf {
					continue
				}
				d := width(i, j) - target
				if d > target*1.5 && i < j-1 {
					break // longer lines only cost more
				}
				if c := cost[l-1][i] + d*d; c < cost[l][j] {
					cost[l][j], from[l][j] = c, i
				}
			}
		}
	}
	out := make([][]PageWord, n)
	j := k
	for l := n; l >= 1; l-- {
		i := from[l][j]
		out[l-1] = words[i:j]
		j = i
	}
	return out
}

func (p *Page) add(line int, kind LineKind, s *Surah, words ...PageWord) {
	ln := &p.Lines[line-1]
	ln.Kind, ln.Surah = kind, s
	ln.Words = append(ln.Words, words...)
}

// Text is the line's words, as one string, to measure it.
func (l *Line) Text() string {
	var b strings.Builder
	for i, w := range l.Words {
		if i > 0 {
			b.WriteByte(' ')
		}
		if w.Word == EndMark {
			b.WriteString(ayahMark(w.Verse.Number))
			continue
		}
		for _, s := range w.Verse.Words[w.Word].Segments {
			b.WriteString(s.Text)
		}
	}
	return b.String()
}

// FirstVerse returns the first verse with text on the page.
func (p *Page) FirstVerse() *Verse {
	for _, l := range p.Lines {
		if len(l.Words) > 0 {
			return l.Words[0].Verse
		}
	}
	return nil
}
