package main

import (
	"runtime"
	"unicode"
)

// zwj, the zero width joiner, asks the letters on its sides for their
// joined forms.
const zwj = '‍'

// shapesBySpan reports whether the system's text shapes each span of a
// paragraph alone: Android's (MyGo draws its text with Android's Paint, a
// span at a time). Elsewhere a word is shaped whole, then colored.
var shapesBySpan = runtime.GOOS == "android"

// spanTexts returns the texts of a run of segments, as the paragraph draws
// them. Where the system shapes each span alone (cutting the text between
// grapheme clusters, a letter with its marks), two things keep a word
// whole:
//
//   - A span that starts with marks would cut their letter from the span
//     before: the marks go with their letter, uncolored.
//   - A letter that joins the next one, ending a span the word goes on
//     after, gets a zero width joiner: it takes its joined form. (One at
//     the start of the next span would belong to the cluster before it.)
//     The next span's first letter takes its unjoined form, which looks
//     much as its joined one beside it.
func spanTexts(segs []Segment) []string {
	out := make([]string, len(segs))
	for i, s := range segs {
		out[i] = s.Text
	}
	if !shapesBySpan {
		return out
	}
	for i := 1; i < len(out); i++ {
		n := 0
		for _, r := range out[i] {
			if !unicode.Is(unicode.Mn, r) {
				break
			}
			n += len(string(r))
		}
		if n > 0 {
			out[i-1] += out[i][:n]
			out[i] = out[i][n:]
		}
	}
	for i := range out {
		last := lastLetter(out[i])
		if last == 0 || !joinsLeft(last) {
			continue
		}
		for _, next := range out[i+1:] {
			if next == "" {
				continue
			}
			if joinsRight([]rune(next)[0]) {
				out[i] += string(zwj)
			}
			break
		}
	}
	return out
}

// lastLetter is the last rune of s that is not a mark, or 0.
func lastLetter(s string) rune {
	var last rune
	for _, r := range s {
		if !unicode.Is(unicode.Mn, r) {
			last = r
		}
	}
	return last
}

// joinsRight reports whether an Arabic letter joins the letter before it
// (in reading order): every letter but hamza.
func joinsRight(r rune) bool {
	return arabicLetter(r) && r != 'ء'
}

// joinsLeft reports whether an Arabic letter joins the letter after it:
// the dual-joining letters, and the tatweel. The alefs, dal, thal, ra,
// zay, waw and ta marbuta join only the one before them.
func joinsLeft(r rune) bool {
	if r == 'ـ' {
		return true
	}
	if !arabicLetter(r) {
		return false
	}
	switch r {
	case 'ء', 'آ', 'أ', 'ؤ', 'إ', 'ا', 'ة', 'د', 'ذ', 'ر', 'ز', 'و', 'ٱ', 'ٲ', 'ٳ', 'ۀ', 'ۥ', 'ۦ':
		return false
	}
	return true
}

func arabicLetter(r rune) bool {
	return unicode.Is(unicode.Arabic, r) && unicode.IsLetter(r)
}
