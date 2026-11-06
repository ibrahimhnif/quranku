package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

// lineKey names a line of the Mushaf in a font, to cache its width.
type lineKey struct{ font, page, line int }

// mushafSize is the font size of the Mushaf's text: the size at which the
// page, or the spread, fits the area last measured, times the Mushaf's zoom.
func (a *app) mushafSize(c *ui.Context) float32 {
	f := quranFonts[a.set.Font]
	zoom := float32(a.set.MushafZoom)
	r := a.mushafArea.Bounds(c)
	if r.W <= 0 || r.H <= 0 {
		return float32(a.set.Size) * 0.8 * f.Scale * zoom // before the first frame
	}
	// A page is about k ems high (15 lines, header, footer, margins) and
	// W+1.6 wide.
	k := 15*float32(a.set.LineSpacing+0.05) + 3.8
	n := float32(len(a.shownPages()))
	fitH := (r.H - 40) / k
	fitW := (r.W - 48 - 20*(n-1)) / (n * (f.Width + 1.8))
	// The page fits the window at a zoom of 1; the zoom scales it from
	// there, larger pages scrolling down, never wider than the window.
	fit := min(fitH, fitW)
	a.mushafMaxZoom = float64(fitW / fit)
	return max(min(fit*zoom, fitW), 10)
}

// shownPages returns the page shown, and the one after it in a spread.
func (a *app) shownPages() []*Page {
	pages := []*Page{a.quran.Pages[a.set.Page]}
	if a.set.Spread && a.set.Page < PageCount {
		pages = append(pages, a.quran.Pages[a.set.Page+1])
	}
	return pages
}

// mushaf shows the Mushaf's page, or two side by side as an open book.
func (a *app) mushaf(c *ui.Context) {
	ui.Column(c).Grow(1).Background(a.pal.Desk).Children(func() {
		a.pageNav(c)
		ui.Scroll(c).Grow(1).TrackScroll(&a.pageScroll).Bind(&a.mushafArea).Children(func() {
			// The first page of a spread is on the right.
			ui.Row(c).Direction(ui.RightToLeft).Justify(ui.Center).AlignItems(ui.Start).
				Padding(8, 24, 32, 24).Gap(20).Children(func() {
				for _, p := range a.shownPages() {
					a.mushafPage(c, p)
				}
			})
		})
	})
}

// pageNav turns the pages. As in a book read right to left, the next page
// is on the left.
func (a *app) pageNav(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Padding(10, 20).Gap(12).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		next := labeledButton(c, iconBack, a.tx(txtNextShort), false).Disabled(a.set.Page+a.pageStep() > PageCount).Tooltip(a.tx(txtNextPage) + " (←)")
		if next.Clicked() {
			a.goPage(a.set.Page + a.pageStep())
		}
		pages := fmt.Sprint(a.set.Page)
		if a.set.Spread && a.set.Page < PageCount {
			pages = fmt.Sprintf("%d–%d", a.set.Page, a.set.Page+1)
		}
		ui.Textf(c, "%s %s / %d", a.tx(txtPage), pages, PageCount).FontSize(13).TextColor(t.TextMuted).FontFeatures("tnum").MinWidth(150).TextAlign(ui.Center)
		prev := ui.Button(c.Key("prev"), "").Label(a.tx(txtPrevShort)).Disabled(a.set.Page <= 1).Tooltip(a.tx(txtPrevPage) + " (→)").Children(func() {
			ui.Text(c, a.tx(txtPrevShort)).SingleLine()
			ui.Icon(c, iconForward)
		})
		if prev.Clicked() {
			a.goPage(a.set.Page - a.pageStep())
		}
	})
}

// mushafPage draws a page: its surah and juz above, 15 lines in a frame,
// its number below.
func (a *app) mushafPage(c *ui.Context, p *Page) {
	t := c.Theme()
	f := quranFonts[a.set.Font]
	size := a.mushafSize(c)
	width := size * f.Width
	lineH := size * float32(a.set.LineSpacing+0.05)
	pad := size * 0.8
	scope := fmt.Sprint("page", p.Number)
	small := size * 0.55

	ui.Column(c.Key(p.Number)).Padding(pad*0.6, pad, pad*0.6, pad).Gap(pad*0.4).Radius(6).
		Background(a.pal.Page).Border(1, t.Border).Shadow(0, 2, 14, 0, ui.RGBA(0, 0, 0, 0.10)).
		Children(func() {
			// The surah on the right, the juz on the left.
			ui.Row(c).Direction(ui.RightToLeft).Width(width + 2*pad*0.5).AlignSelf(ui.Center).Children(func() {
				name := ""
				if len(p.Surahs) > 0 {
					name = "سورة " + p.Surahs[0].Arabic
				}
				ui.Text(c, name).Font(quranFont).FontSize(small).TextColor(t.TextMuted).Grow(1)
				ui.Text(c, "الجزء "+arabicDigits(p.Juz)).Font(quranFont).FontSize(small).TextColor(t.TextMuted)
			})
			ui.Column(c).Padding(pad*0.2, pad*0.5).Radius(4).
				Border(1.5, a.pal.Frame.Alpha(0.55)).
				InsetShadow(0, 0, 0, 3, a.pal.Page).
				Children(func() {
					for i := range p.Lines {
						a.mushafLine(c, p, i, size, width, lineH, scope)
					}
				})
			ui.Text(c, arabicDigits(p.Number)).Font(quranFont).FontSize(small).TextColor(t.TextMuted).AlignSelf(ui.Center)
		})
}

// mushafLine draws line i of a page: its words spread from edge to edge,
// as the printed Mushaf sets them.
func (a *app) mushafLine(c *ui.Context, p *Page, i int, size, width, lineH float32, scope string) {
	t := c.Theme()
	f := quranFonts[a.set.Font]
	l := &p.Lines[i]
	row := ui.Row(c.Key(i)).Direction(ui.RightToLeft).Width(width).Height(lineH).AlignItems(ui.Center)
	switch l.Kind {
	case LineEmpty:
		return
	case LineHeader:
		row.Justify(ui.Center).Children(func() {
			ui.Row(c).Width(width*0.86).Height(lineH*0.78).Justify(ui.Center).AlignItems(ui.Center).
				Radius(lineH*0.39).Border(1.5, a.pal.Frame.Alpha(0.7)).Background(a.pal.Frame.Alpha(0.08)).
				Children(func() {
					ui.Text(c, "سُورَةُ "+l.Surah.Arabic).Font(quranFont).FontSize(size * 0.85).TextColor(t.Accent)
				})
		})
		return
	case LineBasmalah:
		row.Justify(ui.Center).Children(func() {
			ui.RichText(c).Font(f.Family).FontSize(size).NoWrap().Children(func() {
				a.segmentsQuiet(c, a.quran.Basmalah, scope, a.recitedOn(p))
			})
		})
		return
	}

	// Fit the line: wider than the page, its text shrinks; narrower, its
	// words lengthen with kashidas, and spread apart for what is left. The
	// first pages, and lines much shorter than the page, stay centered.
	j := a.justify(p, i)
	// Words measured one by one take a little more than the line measured
	// whole: keep a margin, so no line runs past the frame.
	// A line too long shrinks, keeping a gap between its words wider than
	// the font's space, so words ending and starting in an alif don't run
	// together.
	fs := size
	if gaps := float32(len(l.Words) - 1); j.width+gaps*minWordGap > f.Width*0.97 {
		fs = size * f.Width * 0.97 / (j.width + gaps*minWordGap)
	}
	if j.centered {
		row.Justify(ui.Center).Gap(size * 0.35)
	} else {
		row.Justify(ui.SpaceBetween)
	}
	recited := a.recitedOn(p)
	row.Children(func() {
		for k, w := range l.Words {
			if w.Word == EndMark {
				a.ayahMarkQuiet(c, w.Verse.Number, scope, recited && !a.playing(w.Verse)).Font(f.Family).FontSize(fs).NoWrap()
				continue
			}
			// While a verse of the page is recited, the others fade.
			ui.RichText(c).Font(f.Family).FontSize(fs).NoWrap().Children(func() {
				a.segmentsQuiet(c, j.words[k], scope, recited && !a.playing(w.Verse))
			})
		}
	})
}

// justified is a line of the Mushaf set in a font.
type justified struct {
	words    [][]Segment // by the line's words; nil for ayah marks
	width    float32     // its natural width, in ems
	centered bool
}

// kashidaShare is how much of a line's room kashidas fill; the spaces
// between its words take the rest.
const kashidaShare = 0.8

// minWordGap is the room, in ems, added between the words of a line beyond
// the font's own space.
const minWordGap = 0.12

// justify sets line i of a page in the chosen font: the words of a line
// narrower than the page get kashidas to fill most of the room.
func (a *app) justify(p *Page, i int) *justified {
	k := lineKey{a.set.Font, p.Number, i}
	if j, ok := a.lines[k]; ok {
		return j
	}
	f := quranFonts[a.set.Font]
	l := &p.Lines[i]
	j := &justified{width: a.lineWidth(p, i)}
	j.centered = p.Number <= 2 || j.width < f.Width*0.6
	var words [][]Segment
	for _, w := range l.Words {
		if w.Word != EndMark {
			words = append(words, w.Verse.Words[w.Word].Segments)
		}
	}
	set := words
	// The room kashidas may fill: the line's, less a gap between words.
	room := f.Width - float32(len(l.Words)-1)*minWordGap
	if !j.centered && j.width < room {
		tw := a.textWidth(string(tatweel))
		n := int((room - j.width) * kashidaShare / tw)
		// Kashidas join with their letters, a little wider or narrower
		// than alone: measure, and take some off while it overflows.
		for try := 0; n > 0 && try < 4; try++ {
			var added int
			set, added = addKashidas(words, n)
			w := a.setWidth(l, set)
			if w <= room*0.99 {
				j.width = w
				break
			}
			set = words
			n = min(added, n) - int((w-room*0.97)/tw) - 1
		}
	}
	j.words = make([][]Segment, len(l.Words))
	at := 0
	for wi, w := range l.Words {
		if w.Word != EndMark {
			j.words[wi] = set[at]
			at++
		}
	}
	a.lines[k] = j
	return j
}

// setWidth is the width of a line's words as set, drawn one by one as the
// view draws them, with a space between each.
func (a *app) setWidth(l *Line, words [][]Segment) float32 {
	var w float32
	at := 0
	for i, pw := range l.Words {
		if i > 0 {
			w += a.textWidth(" ")
		}
		if pw.Word == EndMark {
			w += a.textWidth(ayahMark(pw.Verse.Number))
			continue
		}
		var b strings.Builder
		for _, s := range words[at] {
			b.WriteString(s.Text)
		}
		w += a.textWidth(b.String())
		at++
	}
	return w
}

// textWidth measures text in the chosen font, in ems.
func (a *app) textWidth(s string) float32 {
	const size = 100
	w, _ := ui.ShapeText(s, ui.Font{Family: quranFonts[a.set.Font].Family, Size: size}, 0).Size()
	return w / size
}

// lineWidth returns the natural width of a line of the page in the chosen
// font, in ems.
func (a *app) lineWidth(p *Page, i int) float32 {
	k := lineKey{a.set.Font, p.Number, i}
	if w, ok := a.widths[k]; ok {
		return w
	}
	l := &p.Lines[i]
	var words [][]Segment
	for _, w := range l.Words {
		if w.Word != EndMark {
			words = append(words, w.Verse.Words[w.Word].Segments)
		}
	}
	a.widths[k] = a.setWidth(l, words)
	return a.widths[k]
}
