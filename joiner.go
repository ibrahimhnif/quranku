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
// them. Where the system shapes each span alone, a letter whose word goes
// on in the next span would take its unjoined form: zero width joiners on
// both sides of that boundary keep the two letters joined.
func spanTexts(segs []Segment) []string {
	out := make([]string, len(segs))
	for i, s := range segs {
		out[i] = s.Text
	}
	if !shapesBySpan {
		return out
	}
	for i := 0; i+1 < len(segs); i++ {
		if joinsAcross(segs[i].Text, segs[i+1].Text) {
			out[i] += string(zwj)
			out[i+1] = string(zwj) + out[i+1]
		}
	}
	return out
}

// joinsAcross reports whether the last letter of a joins the first of b:
// a letter that joins the one after it, then one that joins the one before
// it, with only marks between them.
func joinsAcross(a, b string) bool {
	var last rune
	for _, r := range a {
		if !unicode.Is(unicode.Mn, r) {
			last = r
		}
	}
	var first rune
	for _, r := range b {
		if !unicode.Is(unicode.Mn, r) {
			first = r
			break
		}
	}
	return joinsLeft(last) && joinsRight(first)
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
