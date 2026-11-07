package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// reader shows the surah verse by verse.
func (a *app) reader(c *ui.Context) {
	s := a.surah
	a.verses.Key = func(i int) any {
		if i == 0 {
			return fmt.Sprint("head", s.ID)
		}
		return s.Verses[i-1].Key
	}
	ui.List(c, &a.verses, len(s.Verses)+1, func(i int) {
		ui.Row(c).Padding(0, 32).Justify(ui.Center).Children(func() {
			ui.Column(c).Grow(1).MaxWidth(880).Children(func() {
				if i == 0 {
					a.surahHeader(c, s)
				} else {
					a.verse(c, &s.Verses[i-1])
				}
			})
		})
	}).Grow(1).Padding(0, 0, 40, 0)
}

func (a *app) surahHeader(c *ui.Context, s *Surah) {
	t := c.Theme()
	ui.Column(c).Padding(32, 0, 16, 0).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "سورة "+s.Arabic).Font(quranFont).FontSize(40).TextColor(t.Accent)
		ui.Textf(c, "%s — %s", s.Name, a.tx(s.Meaning)).FontSize(14).TextColor(t.TextMuted)
		if s.Bismillah {
			ui.Box(c).Height(12)
			a.arabic(c, a.quran.Basmalah, 0).TextAlign(ui.Center).AlignSelf(ui.Stretch)
		}
	})
}

func (a *app) verse(c *ui.Context, v *Verse) {
	t := c.Theme()
	playing := a.playing(v)
	col := ui.Column(c).Padding(20, 16).Gap(12)
	if playing {
		// The verse recited: a frame in the accent, with what plays.
		col.Margin(6, 0).Radius(14).Background(t.Accent.Alpha(0.06)).Border(1, t.Accent.Alpha(0.45))
	} else {
		col.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
	}
	col.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, v.Key).FontSize(12).Bold().TextColor(t.TextMuted).FontFeatures("tnum").Padding(3, 4)
			if playing {
				pb := a.rec.play
				svg, tip := iconPause, a.tx(txtPause)
				if pb.paused {
					svg, tip = iconPlay, a.tx(txtResume)
				}
				if iconButton(c, svg, tip).Padding(2, 8).Clicked() {
					a.togglePause()
				}
				status := a.tx(txtNowReciting)
				switch {
				case pb.loading:
					status = a.tx(txtLoading)
				case pb.paused:
					status = a.tx(txtPaused)
				}
				ui.Row(c).Gap(6).AlignItems(ui.Center).Padding(3, 10).Radius(999).Background(t.Accent).Children(func() {
					ui.Icon(c, iconSound).FontSize(12).TextColor(t.AccentText)
					ui.Text(c, status).FontSize(11).Bold().TextColor(t.AccentText)
				})
				return
			}
			if iconButton(c, iconPlay, a.tx(txtPlayFromHere)).Padding(2, 8).Clicked() {
				a.playFrom(v.Surah, v.Number, 0)
			}
			if iconButton(c, iconShare, a.tx(txtShareVerse)).Padding(2, 8).Clicked() {
				a.openShare(ShareCard{Kind: CardVerse, Verse: v})
			}
			if a.reading.isRead(v.Index) {
				ui.Icon(c, iconCheck).FontSize(13).TextColor(t.TextMuted).Tooltip(a.tx(txtVersesRead))
			}
		})
		a.arabic(c, v.Segments, v.Number)
		if a.set.Translation {
			ui.Text(c, a.translationText(v)).FontSize(float32(a.set.TransSize)).LineHeight(1.6).TextColor(t.TextMuted)
		}
	})
}

// arabic shows a verse's text in one paragraph. A verse number above zero
// ends it with an ayah mark.
func (a *app) arabic(c *ui.Context, segs []Segment, number int) ui.Element {
	f := quranFonts[a.set.Font]
	size := float32(a.set.Size) * f.Scale
	return ui.RichText(c).Font(f.Family).FontSize(size).LineHeight(float32(a.set.LineSpacing)).TextAlign(ui.Start).Children(func() {
		a.segments(c, segs)
		if number > 0 {
			ui.Text(c, " ")
			a.ayahMark(c, number)
		}
	})
}

// segments builds a run of text inside a paragraph, each tajweed segment
// colored, with its rule on hover, and studied when clicked.
func (a *app) segments(c *ui.Context, segs []Segment) {
	a.segmentsQuiet(c, segs, false)
}

// quietAlpha is the opacity of the verses around the one recited.
const quietAlpha = 0.38

// segmentsQuiet builds segments as segments does; quiet ones fade, as the
// verses around the one recited do.
func (a *app) segmentsQuiet(c *ui.Context, segs []Segment, quiet bool) {
	t := c.Theme()
	alpha := float32(1)
	if quiet {
		alpha = quietAlpha
	}
	for _, seg := range segs {
		el := ui.Text(c, seg.Text)
		r := seg.Rule
		if r == nil || !a.set.Colors || (a.focus != nil && a.focus != r) {
			if alpha < 1 {
				el.TextColor(t.Text.Alpha(alpha))
			}
			continue
		}
		col := r.Color(t)
		el.Cursor(ui.CursorPointer).TextColor(col.Alpha(alpha))
		if el.Clicked() {
			a.focus = r
			a.set.Guide = true
		}
		ui.TooltipBase(c, el, func(tip ui.Element) {
			a.ruleCard(c, tip, r)
		})
	}
}

// ayahMark builds the ornament with a verse's number.
func (a *app) ayahMark(c *ui.Context, n int) ui.Element {
	return a.ayahMarkQuiet(c, n, false)
}

func (a *app) ayahMarkQuiet(c *ui.Context, n int, quiet bool) ui.Element {
	t := c.Theme()
	col := t.Accent
	if quiet {
		col = col.Alpha(quietAlpha)
	}
	return ui.Text(c, ayahMark(n)).TextColor(col)
}

// ruleCard styles the hover card explaining a rule.
func (a *app) ruleCard(c *ui.Context, tip ui.Element, r *Rule) {
	t := c.Theme()
	col := r.Color(t)
	tip.Margin(0, 0, 6, 0).Padding(12, 14).Radius(10).Width(280).Gap(6).
		Background(t.Background).Border(1, t.Border).TextColor(t.Text).
		Shadow(0, 8, 24, 0, ui.RGBA(0, 0, 0, 0.18))
	tip.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(10, 10).Radius(5).Background(col)
			ui.Text(c, a.tx(r.Name)).Bold().FontSize(14).Grow(1)
			ui.Text(c, r.Arabic).Font(quranFont).FontSize(18).TextColor(col)
		})
		ui.Text(c, a.tx(r.Summary)).FontSize(13)
		ui.Text(c, a.tx(txtLearnMore)).FontSize(11).TextColor(t.TextMuted)
	})
}
