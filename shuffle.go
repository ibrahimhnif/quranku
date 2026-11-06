package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/egoist/mygo/ui"
)

// Where a shuffle picks from.
const (
	ScopeQuran = iota
	ScopeJuz
	ScopeSurah
)

// How much a shuffle picks.
const (
	LengthOne = iota
	LengthThree
	LengthFive
	LengthPage
)

// Passage is a run of verses a shuffle picked: count verses from First
// (a verse index), or a whole Mushaf page.
type Passage struct {
	First, Count int
	Page         int // the page, when the shuffle picked one
}

// shuffle is the state of the shuffle section.
type shuffle struct {
	scope, length int
	juz           int // 1 to 30
	surah         int // 1 to 114
	current       *Passage
	history       []Passage
	hide          bool // memorization: the text hides until revealed
	rng           *rand.Rand
}

// verses returns the verses of the current passage.
func (s *shuffle) verses(q *Quran) []*Verse {
	if s.current == nil {
		return nil
	}
	return q.All[s.current.First : s.current.First+s.current.Count]
}

// pick chooses a passage within the scope: verses that follow each other in
// one surah, or a page.
func (s *shuffle) pick(q *Quran) Passage {
	in := func(v *Verse) bool {
		switch s.scope {
		case ScopeJuz:
			return v.Juz == s.juz
		case ScopeSurah:
			return v.Surah == s.surah
		}
		return true
	}
	if s.length == LengthPage {
		var pages []int
		for _, p := range q.Pages[1:] {
			if v := p.FirstVerse(); v != nil && in(v) {
				pages = append(pages, p.Number)
			}
		}
		if len(pages) > 0 {
			n := pages[s.rng.IntN(len(pages))]
			var first, last *Verse
			for _, l := range q.Pages[n].Lines {
				for _, w := range l.Words {
					if first == nil {
						first = w.Verse
					}
					last = w.Verse
				}
			}
			return Passage{First: first.Index, Count: last.Index - first.Index + 1, Page: n}
		}
	}
	want := map[int]int{LengthOne: 1, LengthThree: 3, LengthFive: 5}[s.length]
	var candidates []int
	for _, v := range q.All {
		if in(v) {
			candidates = append(candidates, v.Index)
		}
	}
	first := candidates[s.rng.IntN(len(candidates))]
	// Keep the passage in the verse's surah, and the scope.
	v := q.All[first]
	surah := q.Surahs[v.Surah-1]
	count := min(want, len(surah.Verses))
	start := min(first, surah.Verses[len(surah.Verses)-1].Index-count+1)
	for start+count-1 >= 0 && !in(q.All[start+count-1]) && count > 1 {
		count--
	}
	return Passage{First: start, Count: count}
}

// next picks a new passage and keeps the last ones.
func (a *app) shuffleNext() {
	p := a.shuf.pick(a.quran)
	if a.shuf.current != nil {
		a.shuf.history = append([]Passage{*a.shuf.current}, a.shuf.history...)
		if len(a.shuf.history) > 8 {
			a.shuf.history = a.shuf.history[:8]
		}
	}
	a.shuf.current = &p
	a.revealed = false
	a.shuffleScroll.Y = 0
}

// passageName names a passage: "Al-Baqarah 255–257".
func (a *app) passageName(p Passage) string {
	v := a.quran.All[p.First]
	s := a.quran.Surahs[v.Surah-1]
	last := a.quran.All[p.First+p.Count-1]
	if p.Count == 1 {
		return fmt.Sprintf("%s %d:%d", s.Name, s.ID, v.Number)
	}
	if last.Surah != v.Surah {
		return fmt.Sprintf("%s %d:%d – %d:%d", s.Name, v.Surah, v.Number, last.Surah, last.Number)
	}
	return fmt.Sprintf("%s %d:%d–%d", s.Name, s.ID, v.Number, last.Number)
}

// Texts of the shuffle.
var (
	txtShuffle      = T{"Shuffle", "Acak"}
	txtShuffleIntro = T{"Pick a random passage to read, recite or test your memorization.", "Pilih bacaan acak untuk dibaca, didengarkan, atau menguji hafalan."}
	txtShuffleBtn   = T{"Shuffle now", "Acak sekarang"}
	txtAgain        = T{"Again", "Acak lagi"}
	txtScope        = T{"From", "Dari"}
	txtWholeQuran   = T{"Whole Quran", "Seluruh Al-Qur'an"}
	txtThisSurah    = T{"Surah", "Surah"}
	txtLength       = T{"Length", "Panjang"}
	txtOneVerse     = T{"1 verse", "1 ayat"}
	txtThreeVerses  = T{"3 verses", "3 ayat"}
	txtFiveVerses   = T{"5 verses", "5 ayat"}
	txtOnePage      = T{"1 page", "1 halaman"}
	txtHideText     = T{"Memorization test: hide the text", "Tes hafalan: sembunyikan teks"}
	txtReveal       = T{"Show the text", "Tampilkan teks"}
	txtHiddenHint   = T{"Recite it from memory, then show the text to check.", "Lafalkan dari hafalan, lalu tampilkan teks untuk memeriksa."}
	txtOpenVerses   = T{"Open in reading", "Buka di bacaan"}
	txtOpenMushaf   = T{"Open in Mushaf", "Buka di mushaf"}
	txtPlay         = T{"Play", "Putar"}
	txtRecent       = T{"Recent", "Terakhir"}
	txtPageN        = T{"page", "halaman"}
	txtStartHint    = T{"Press Shuffle now to begin.", "Tekan Acak sekarang untuk mulai."}
)

// shufflePage is the shuffle section.
func (a *app) shufflePage(c *ui.Context) {
	t := c.Theme()
	s := &a.shuf
	ui.Scroll(c).Grow(1).TrackScroll(&a.shuffleScroll).Children(func() {
		pageColumn(c, 960, func() {
			{
				sectionTitle(c, a.tx(txtShuffle))
				ui.Text(c, a.tx(txtShuffleIntro)).TextColor(t.TextMuted)
				// Options.
				ui.Column(c).Padding(16).Gap(12).Radius(12).Background(t.Surface).Border(1, t.Border).Children(func() {
					ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
						ui.Text(c, a.tx(txtScope)).FontSize(13).Width(70)
						ui.Segmented(c, &s.scope, a.tx(txtWholeQuran), a.tx(txtJuz), a.tx(txtThisSurah)).Label(a.tx(txtScope))
						switch s.scope {
						case ScopeJuz:
							juz := float64(s.juz)
							if ui.Stepper(c, &juz, 1, 30, 1).Label(a.tx(txtJuz)).Width(110).Changed() || int(juz) != s.juz {
								s.juz = int(juz)
							}
						case ScopeSurah:
							ui.Textf(c, "%d. %s", s.surah, a.quran.Surahs[s.surah-1].Name).FontSize(13).TextColor(t.Accent)
						}
					})
					ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
						ui.Text(c, a.tx(txtLength)).FontSize(13).Width(70)
						ui.Segmented(c, &s.length, a.tx(txtOneVerse), a.tx(txtThreeVerses), a.tx(txtFiveVerses), a.tx(txtOnePage)).Label(a.tx(txtLength))
					})
					a.settingRow(c, "hide", &s.hide, a.tx(txtHideText))
				})
				label := a.tx(txtShuffleBtn)
				if s.current != nil {
					label = a.tx(txtAgain)
				}
				if labeledButton(c, iconShuffle, label, true).Padding(12, 28).AlignSelf(ui.Start).Clicked() {
					a.shuffleNext()
				}
				if s.current == nil {
					ui.Text(c, a.tx(txtStartHint)).TextColor(t.TextMuted)
					return
				}
				a.passageCard(c, *s.current)
				if len(s.history) > 0 {
					ui.Text(c, a.tx(txtRecent)).FontSize(13).Bold().TextColor(t.TextMuted)
					ui.Row(c).Gap(8).Wrap().Children(func() {
						for i, p := range s.history {
							if ui.Button(c.Key(i), a.passageName(p)).Clicked() {
								cur := *s.current
								s.history[i] = cur
								s.current = &p
								a.revealed = false
							}
						}
					})
				}
			}
		})
	})
}

// passageCard shows the passage picked, with what to do with it.
func (a *app) passageCard(c *ui.Context, p Passage) {
	t := c.Theme()
	first := a.quran.All[p.First]
	ui.Column(c).Padding(20).Gap(10).Radius(12).Background(t.Background).Border(1, t.Border).Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Grow(1).Gap(2).Children(func() {
				ui.Text(c, a.passageName(p)).FontSize(17).Bold()
				ui.Textf(c, "%s %d · %s %d", a.tx(txtJuz), first.Juz, a.tx(txtPageN), first.Page).FontSize(12).TextColor(t.TextMuted)
			})
			if labeledButton(c, iconPlay, a.tx(txtPlay), false).Clicked() {
				a.playFrom(first.Surah, first.Number, a.quran.All[p.First+p.Count-1].Number)
			}
			if ui.Button(c, a.tx(txtOpenVerses)).Clicked() {
				a.openVerse(first, ModeVerses)
			}
			if ui.Button(c, a.tx(txtOpenMushaf)).Clicked() {
				a.openVerse(first, ModeMushaf)
			}
		})
		if a.shuf.hide && !a.revealed {
			ui.Column(c).Padding(30).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.tx(txtHiddenHint)).TextColor(t.TextMuted)
				if ui.PrimaryButton(c, a.tx(txtReveal)).Clicked() {
					a.revealed = true
				}
			})
			return
		}
		for _, v := range a.quran.All[p.First : p.First+p.Count] {
			a.verse(c.Key(v.Key), v)
		}
	})
}

// openVerse shows a verse in the reader, in a mode.
func (a *app) openVerse(v *Verse, mode int) {
	a.section = SectionRead
	a.set.Mode = mode
	a.shown.Mode = mode // no switch to follow
	a.surah = a.quran.Surahs[v.Surah-1]
	if mode == ModeMushaf {
		a.goPage(v.Page)
		a.surah = a.quran.Surahs[v.Surah-1]
		return
	}
	a.verses.ScrollTo(v.Number, ui.Start)
}
