package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// Texts of the home page.
var (
	txtHome          = T{"Today", "Hari ini"}
	txtGreeting      = T{"Assalamu alaikum", "Assalamu'alaikum"}
	txtContinue      = T{"Continue", "Lanjutkan"}
	txtStartFatihah  = T{"Begin with Al-Fatihah", "Mulai dari Al-Fatihah"}
	txtVerseN        = T{"%s, verse %d", "%s, ayat %d"}
	txtOfToday       = T{"%d of %d verses today", "%d dari %d ayat hari ini"}
	txtGoalReached   = T{"Today's goal reached: %d verses", "Target hari ini tercapai: %d ayat"}
	txtDaysInRow     = T{"%d days in a row", "%d hari berturut-turut"}
	txtStartStreak   = T{"Read today to start a streak", "Baca hari ini untuk memulai runtutan"}
	txtVersesSoFar   = T{"%s verses read", "%s ayat terbaca"}
	txtShuffleVerse  = T{"Shuffle a passage", "Acak ayat"}
	txtListenHere    = T{"Listen from here", "Dengarkan dari sini"}
	txtOpenMushafAlt = T{"Open in the Mushaf", "Buka di mushaf"}
	txtOpenVersesAlt = T{"Open verse by verse", "Buka per ayat"}
	txtAllSurahs     = T{"The Quran", "Al-Qur'an"}
	txtBySurah       = T{"Surah", "Surah"}
	txtByJuzTab      = T{"Juz", "Juz"}
	txtJuzStarts     = T{"from %s %d:%d", "mulai %s %d:%d"}
)

var (
	idDays   = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	idMonths = []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
)

// longDate writes a date as people say it: "Jumat, 6 November 2026".
func (a *app) longDate(t time.Time) string {
	if a.set.Lang == Indonesian {
		return fmt.Sprintf("%s, %d %s %d", idDays[t.Weekday()], t.Day(), idMonths[t.Month()-1], t.Year())
	}
	return t.Format("Monday, 2 January 2006")
}

// thousands writes n with the language's thousands separator.
func (a *app) thousands(n int) string {
	s := fmt.Sprint(n)
	sep := ","
	if a.set.Lang == Indonesian {
		sep = "."
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + sep + s[i:]
	}
	return s
}

// lastVerse is where the reader stopped.
func (a *app) lastVerse() *Verse { return a.quran.All[a.set.LastVerse] }

// continueReading opens the reader where it stopped.
func (a *app) continueReading(mode int) {
	a.openVerse(a.lastVerse(), mode)
}

// homePage is the first screen: where reading stopped, today's goal, and
// the Quran's surahs and juz.
func (a *app) homePage(c *ui.Context) {
	t := c.Theme()
	a.indexWidth = a.contentWidth(c, 1280)
	ui.Scroll(c).Grow(1).TrackScroll(&a.homeScroll).Children(func() {
		pageColumn(c, 1280, func() {
			{
				ui.Column(c).Gap(4).Children(func() {
					ui.Text(c, a.tx(txtGreeting)).FontSize(26).Bold()
					ui.Text(c, a.longDate(a.now())).FontSize(14).TextColor(t.TextMuted)
				})
				a.continuePanel(c)
				a.homeStrip(c)
				a.quranIndex(c)
			}
		})
	})
}

// continuePanel shows the verse where reading stopped, and today's goal.
func (a *app) continuePanel(c *ui.Context) {
	t := c.Theme()
	v := a.lastVerse()
	s := a.quran.Surahs[v.Surah-1]
	today := a.reading.day(a.now()).Read
	goal := int(a.set.DailyGoal)
	fresh := a.set.LastVerse == 0 && a.reading.uniqueRead(nil) == 0
	ui.Row(c).Padding(28, 28, 24, 28).Gap(32).Radius(18).Background(t.Surface).Border(1, t.Border).
		Shadow(0, 6, 24, 0, ui.RGBA(0, 0, 0, 0.05)).Children(func() {
		ui.Column(c).Grow(1).Basis(0).Gap(14).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				title := fmt.Sprintf(a.tx(txtVerseN), s.Name, v.Number)
				if fresh {
					title = a.tx(txtStartFatihah)
				}
				ui.Text(c, title).FontSize(22).Bold()
				ui.Textf(c, "%s · %s %d · %s %d", a.tx(s.Meaning), a.tx(txtJuz), v.Juz, a.tx(txtPage), v.Page).FontSize(13).TextColor(t.TextMuted)
			})
			f := quranFonts[a.set.Font]
			ui.RichText(c).Font(f.Family).FontSize(30).LineHeight(1.9).MaxLines(2).TextAlign(ui.Start).Children(func() {
				a.segments(c, v.Segments, "continue")
			})
			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Progress(c, min(float64(today)/float64(max(goal, 1)), 1)).Width(220).Label(a.tx(txtGoal))
				ui.Text(c, a.todayLine(today, goal)).FontSize(13).TextColor(t.TextMuted)
			})
		})
		ui.Column(c).Gap(10).Justify(ui.End).AlignItems(ui.Stretch).Width(210).Children(func() {
			if labeledButton(c, iconRead, a.tx(txtContinue), true).Padding(12, 20).Clicked() {
				a.continueReading(a.set.Mode)
			}
			other, label := ModeMushaf, txtOpenMushafAlt
			if a.set.Mode == ModeMushaf {
				other, label = ModeVerses, txtOpenVersesAlt
			}
			if ui.Button(c, a.tx(label)).Clicked() {
				a.continueReading(other)
			}
		})
	})
}

// homeStrip is a quiet line of the habit and the other ways in.
func (a *app) homeStrip(c *ui.Context) {
	t := c.Theme()
	streak := a.reading.streak(a.now())
	ui.Row(c).Gap(28).AlignItems(ui.Center).Wrap().Padding(0, 4).Children(func() {
		note := fmt.Sprintf(a.tx(txtDaysInRow), streak)
		if streak == 0 {
			note = a.tx(txtStartStreak)
		}
		a.stripItem(c, iconStreak, note, nil)
		a.stripItem(c, iconBook, fmt.Sprintf(a.tx(txtVersesSoFar), a.thousands(a.reading.uniqueRead(nil))), func() { a.section = SectionStats })
		ui.Spacer(c)
		a.stripItem(c, iconShuffle, a.tx(txtShuffleVerse), func() { a.section = SectionShuffle })
		a.stripItem(c, iconListen, a.tx(txtListenHere), func() {
			v := a.lastVerse()
			a.playFrom(v.Surah, v.Number, 0)
		})
	})
	_ = t
}

// stripItem is an icon and a line; with an action, it is a link to it.
func (a *app) stripItem(c *ui.Context, svg *ui.SVG, label string, action func()) {
	t := c.Theme()
	row := ui.Row(c.Key(label)).Gap(8).AlignItems(ui.Center).Padding(6, 8).Radius(8)
	if action != nil {
		row.Cursor(ui.CursorPointer).Label(label)
		if row.Hovered() {
			row.Background(t.SurfaceHover)
		}
		if row.Clicked() {
			action()
		}
	}
	row.Children(func() {
		col := t.TextMuted
		if action != nil {
			col = t.Accent
		}
		ui.Icon(c, svg).FontSize(16).TextColor(col)
		ui.Text(c, label).FontSize(14).TextColor(t.Text)
	})
}

// quranIndex lists the surahs, or the juz, to open one.
func (a *app) quranIndex(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Gap(14).Children(func() {
		ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.tx(txtAllSurahs)).FontSize(18).Bold().Grow(1)
			ui.Segmented(c, &a.indexTab, a.tx(txtBySurah), a.tx(txtByJuzTab)).Label(a.tx(txtAllSurahs))
			if a.indexTab == 0 {
				ui.SearchField(c, &a.query).Placeholder(a.tx(txtSearch)).Width(260)
			}
		})
		if a.indexTab == 1 {
			a.juzIndex(c)
			return
		}
		results := a.quran.Search(a.query)
		if len(results) == 0 {
			ui.Text(c, a.tx(txtNoMatch)).TextColor(t.TextMuted).Padding(12, 4)
			return
		}
		ui.Grid(c).Columns(a.indexColumns()).Gap(4).Children(func() {
			for _, s := range results {
				a.surahTile(c, s)
			}
		})
	})
}

func (a *app) surahTile(c *ui.Context, s *Surah) {
	t := c.Theme()
	current := s.ID == a.lastVerse().Surah
	row := ui.Row(c.Key(s.ID)).Padding(10, 12).Gap(12).AlignItems(ui.Center).Radius(10).Cursor(ui.CursorPointer)
	if row.Hovered() {
		row.Background(t.SurfaceHover)
	}
	if row.Clicked() {
		a.open(s)
	}
	row.Children(func() {
		num := ui.Text(c, fmt.Sprint(s.ID)).FontSize(12).Bold().Width(30).Height(30).
			TextAlign(ui.Center).FixedLineHeight(30).Radius(15).FontFeatures("tnum")
		if current {
			num.Background(t.Accent).TextColor(t.AccentText)
		} else {
			num.Background(t.Surface).TextColor(t.TextMuted)
		}
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, s.Name).FontSize(14).SingleLine()
			ui.Textf(c, "%s · %d %s", a.tx(s.Meaning), s.VerseCount, a.tx(txtVerses)).FontSize(12).TextColor(t.TextMuted).SingleLine()
		})
		ui.Text(c, s.Arabic).Font(quranFont).FontSize(19)
	})
}

// juzIndex lists the 30 juz with where each starts and how much was read.
func (a *app) juzIndex(c *ui.Context) {
	t := c.Theme()
	first := [31]*Verse{}
	total, read := [31]int{}, [31]int{}
	for _, v := range a.quran.All {
		if first[v.Juz] == nil {
			first[v.Juz] = v
		}
		total[v.Juz]++
		if a.reading.isRead(v.Index) {
			read[v.Juz]++
		}
	}
	ui.Grid(c).Columns(a.indexColumns()).Gap(4).Children(func() {
		for j := 1; j <= 30; j++ {
			v := first[j]
			s := a.quran.Surahs[v.Surah-1]
			row := ui.Row(c.Key(j)).Padding(10, 12).Gap(12).AlignItems(ui.Center).Radius(10).Cursor(ui.CursorPointer)
			if row.Hovered() {
				row.Background(t.SurfaceHover)
			}
			if row.Clicked() {
				a.openVerse(v, a.set.Mode)
			}
			row.Children(func() {
				ui.Text(c, fmt.Sprint(j)).FontSize(12).Bold().Width(30).Height(30).TextAlign(ui.Center).
					FixedLineHeight(30).Radius(15).Background(t.Surface).TextColor(t.TextMuted)
				ui.Column(c).Grow(1).Gap(4).Children(func() {
					ui.Textf(c, "%s %d", a.tx(txtJuz), j).FontSize(14)
					ui.Text(c, fmt.Sprintf(a.tx(txtJuzStarts), s.Name, v.Surah, v.Number)).FontSize(12).TextColor(t.TextMuted).SingleLine()
				})
				ui.Progress(c, float64(read[j])/float64(total[j])).Width(48).Label(fmt.Sprintf("%s %d", a.tx(txtJuz), j))
			})
		}
	})
}

// pageColumn lays a section's content out in the middle of the window,
// maxW wide at most.
func pageColumn(c *ui.Context, maxW float32, fn func()) {
	ui.Row(c).Padding(36, 40, 48, 40).Justify(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MaxWidth(maxW).Gap(28).Children(fn)
	})
}

// contentWidth is the width of the pages' area, as last laid out, less its
// margins; 0 before the first frame.
func (a *app) contentWidth(c *ui.Context, maxW float32) float32 {
	w := a.contentArea.Bounds(c).W - 80
	return max(min(w, maxW), 0)
}

// indexColumns is how many surahs a row of the index holds: a row each
// 320 wide, two to four.
func (a *app) indexColumns() int {
	w := a.indexWidth
	if w <= 0 {
		return 3
	}
	return min(max(int(w/320), 2), 4)
}

// sectionTitle is the title of a section page.
func sectionTitle(c *ui.Context, s string) {
	ui.Text(c, strings.TrimSpace(s)).FontSize(24).Bold()
}
