package main

import (
	"strings"
	"unicode"
)

// Kashida (tatweel, ـ) lengthens the joins between letters, as the printed
// Mushaf fills its lines rather than widening the spaces between words.

const tatweel = 'ـ'

// joinsNext reports whether letter r connects to the letter after it
// (Unicode's dual-joining Arabic letters, and the tatweel).
func joinsNext(r rune) bool {
	return strings.ContainsRune("بتثجحخسشصضطظعغفقكلمنهيىئـٮڡٯ", r)
}

// isBase reports whether r starts a letter's cluster, which marks follow.
func isBase(r rune) bool {
	return unicode.Is(unicode.Lo, r) || r == tatweel
}

// alifs are the letters lam forms a ligature with; a kashida between them
// would break it.
const alifs = "اأإآٱٲ"

// kashidaSpot is where a kashida may go: after rune index At of a word's
// runes, at the end of its letter's marks.
type kashidaSpot struct{ word, at int }

// kashidaSpots returns the places in words a kashida may lengthen: after a
// letter joining the next, and its marks, right before the next letter.
// Words with the ligature of the name of Allah are left whole.
func kashidaSpots(words [][]rune) []kashidaSpot {
	var spots []kashidaSpot
	for wi, w := range words {
		if strings.Contains(skeleton(w), "لله") {
			continue
		}
		for i := 0; i < len(w); i++ {
			if !isBase(w[i]) || !joinsNext(w[i]) {
				continue
			}
			j := i + 1
			for j < len(w) && unicode.Is(unicode.Mn, w[j]) {
				j++ // the letter's marks
			}
			if j >= len(w) || !unicode.Is(unicode.Lo, w[j]) || w[j] == 'ء' {
				continue // no letter after it to join
			}
			if w[i] == 'ل' && strings.ContainsRune(alifs, w[j]) {
				continue // lam-alif
			}
			spots = append(spots, kashidaSpot{wi, j})
		}
	}
	return spots
}

// skeleton is a word's letters without their marks.
func skeleton(w []rune) string {
	var b strings.Builder
	for _, r := range w {
		if unicode.Is(unicode.Lo, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// addKashidas lengthens words with n kashidas, spread evenly over the spots
// they have, at most three in a spot. It returns the new words and how
// many it added.
func addKashidas(words [][]Segment, n int) ([][]Segment, int) {
	runes := make([][]rune, len(words))
	for i, w := range words {
		for _, s := range w {
			runes[i] = append(runes[i], []rune(s.Text)...)
		}
	}
	spots := kashidaSpots(runes)
	if n <= 0 || len(spots) == 0 {
		return words, 0
	}
	n = min(n, 3*len(spots))
	// count[k] kashidas go in spot k: every spot gets n/len, and the rest
	// go to spots spread across the line.
	count := make([]int, len(spots))
	for k := range count {
		count[k] = n / len(spots)
	}
	if rest := n % len(spots); rest > 0 {
		for r := 0; r < rest; r++ {
			count[(2*r+1)*len(spots)/(2*rest)]++
		}
	}
	// Insert them, keeping each kashida in the segment of the letter it
	// lengthens.
	extra := make([]map[int]int, len(words)) // word -> rune index -> kashidas
	for k, s := range spots {
		if count[k] == 0 {
			continue
		}
		if extra[s.word] == nil {
			extra[s.word] = map[int]int{}
		}
		extra[s.word][s.at] += count[k]
	}
	out := make([][]Segment, len(words))
	for wi, w := range words {
		if extra[wi] == nil {
			out[wi] = w
			continue
		}
		at := 0
		for _, s := range w {
			r := []rune(s.Text)
			var b strings.Builder
			for i, ch := range r {
				if k := extra[wi][at+i]; k > 0 && i > 0 {
					b.WriteString(strings.Repeat(string(tatweel), k))
				}
				b.WriteRune(ch)
			}
			// A spot at the start of the next segment ends this one.
			if k := extra[wi][at+len(r)]; k > 0 {
				b.WriteString(strings.Repeat(string(tatweel), k))
			}
			at += len(r)
			out[wi] = append(out[wi], Segment{Text: b.String(), Rule: s.Rule})
		}
	}
	return out, n
}
