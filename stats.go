package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

// DayStats is a day's reading.
type DayStats struct {
	Read     int     `json:"read"`     // verses read, each once a day
	Listened int     `json:"listened"` // verses heard recited
	Seconds  float64 `json:"seconds"`  // time reading
	// Seen marks the verses read that day, a bit each, in hex: today's
	// only, so a verse read again after a restart counts once.
	Seen string `json:"seen,omitempty"`
	// Surahs counts the verses read of each surah, Hours those read in
	// each hour of the day, and Reciters the verses heard of each reciter:
	// what the recaps tell.
	Surahs   map[int]int `json:"surahs,omitempty"`
	Hours    [24]int     `json:"hours"`
	Reciters map[int]int `json:"reciters,omitempty"`
}

// Stats is the reading history, as stats.json keeps it.
type Stats struct {
	Days map[string]*DayStats `json:"days"` // by date, "2026-11-06"
	// Read marks every verse ever read, a bit each, in hex.
	Read string `json:"read"`
	// Badges are the badges earned, with the day each was.
	Badges map[string]string `json:"badges,omitempty"`
}

// reading tracks what is read as the app runs.
type reading struct {
	stats Stats
	bits  []byte // Stats.Read, decoded
	file  *jsonFile
	dirty bool
	last  time.Time // the last tick
	// progress is the time each verse has been read this session.
	progress map[int]float64
	// today and seen are the day being counted and its verses read.
	today string
	seen  []byte
}

func newReading(file *jsonFile) *reading {
	r := &reading{file: file, progress: map[int]float64{}}
	if err := file.load(&r.stats); err != nil {
		r.stats = Stats{}
	}
	if r.stats.Days == nil {
		r.stats.Days = map[string]*DayStats{}
	}
	r.bits, _ = hex.DecodeString(r.stats.Read)
	if len(r.bits) != (VerseCount+7)/8 {
		r.bits = make([]byte, (VerseCount+7)/8)
	}
	return r
}

func dayKey(t time.Time) string { return t.Format("2006-01-02") }

func (r *reading) day(t time.Time) *DayStats {
	k := dayKey(t)
	d := r.stats.Days[k]
	if d == nil {
		d = &DayStats{}
		r.stats.Days[k] = d
	}
	return d
}

func (r *reading) isRead(i int) bool { return r.bits[i/8]&(1<<(i%8)) != 0 }

// seenToday returns the bits of the verses read on now's day.
func (r *reading) seenToday(now time.Time) []byte {
	if k := dayKey(now); k != r.today {
		r.today = k
		r.seen, _ = hex.DecodeString(r.day(now).Seen)
		if len(r.seen) != (VerseCount+7)/8 {
			r.seen = make([]byte, (VerseCount+7)/8)
		}
		r.progress = map[int]float64{}
	}
	return r.seen
}

// readOn reports whether verse i was read on now's day.
func (r *reading) readOn(i int, now time.Time) bool {
	return r.seenToday(now)[i/8]&(1<<(i%8)) != 0
}

// markRead counts verse v read now, once a day.
func (r *reading) markRead(v *Verse, now time.Time) {
	i := v.Index
	if r.readOn(i, now) {
		return
	}
	r.seen[i/8] |= 1 << (i % 8)
	d := r.day(now)
	d.Read++
	if d.Surahs == nil {
		d.Surahs = map[int]int{}
	}
	d.Surahs[v.Surah]++
	d.Hours[now.Hour()]++
	r.bits[i/8] |= 1 << (i % 8)
	r.dirty = true
}

// markListened counts a verse of a reciter heard to its end.
func (r *reading) markListened(reciter int, now time.Time) {
	d := r.day(now)
	d.Listened++
	if d.Reciters == nil {
		d.Reciters = map[int]int{}
	}
	d.Reciters[reciter]++
	r.dirty = true
}

// readNeed is how long verse v takes to read: about 0.45 s a word, at least
// 2 s and at most 20.
func readNeed(v *Verse) float64 {
	return min(max(0.45*float64(len(v.Words)), 2), 20)
}

// credit gives dt seconds of reading to the visible verses in their order:
// the first not yet counted takes it, and what it does not need goes on to
// the next, as a reader reads one verse after another.
func (r *reading) credit(visible []*Verse, dt float64, now time.Time) {
	if len(visible) == 0 || dt <= 0 {
		return
	}
	r.day(now).Seconds += dt
	r.dirty = true
	for _, v := range visible {
		if dt <= 0 {
			break
		}
		if r.readOn(v.Index, now) {
			continue
		}
		need := readNeed(v) - r.progress[v.Index]
		if dt < need {
			r.progress[v.Index] += dt
			return
		}
		dt -= need
		r.markRead(v, now)
	}
}

// save writes the stats when they changed.
func (r *reading) save() {
	if !r.dirty {
		return
	}
	r.dirty = false
	r.stats.Read = hex.EncodeToString(r.bits)
	// Only today keeps its verses; past days keep their counts.
	for k, d := range r.stats.Days {
		if k == r.today {
			d.Seen = hex.EncodeToString(r.seen)
		} else {
			d.Seen = ""
		}
	}
	r.file.save(r.stats)
}

// uniqueRead counts the verses ever read, of those in in, or all.
func (r *reading) uniqueRead(in func(i int) bool) (n int) {
	for i := range VerseCount {
		if r.isRead(i) && (in == nil || in(i)) {
			n++
		}
	}
	return n
}

// streak counts the days in a row with a verse read, ending today, or
// yesterday when today has none yet.
func (r *reading) streak(now time.Time) int {
	d := now
	if s := r.stats.Days[dayKey(d)]; s == nil || s.Read == 0 {
		d = d.AddDate(0, 0, -1)
	}
	n := 0
	for {
		s := r.stats.Days[dayKey(d)]
		if s == nil || s.Read == 0 {
			return n
		}
		n++
		d = d.AddDate(0, 0, -1)
	}
}

// visibleVerses returns the verses the reader has in front of them, in
// order.
func (a *app) visibleVerses() []*Verse {
	switch a.section {
	case SectionRead:
		if a.set.Mode == ModeMushaf {
			var out []*Verse
			seen := map[*Verse]bool{}
			for _, p := range a.shownPages() {
				for _, l := range p.Lines {
					for _, w := range l.Words {
						if !seen[w.Verse] {
							seen[w.Verse] = true
							out = append(out, w.Verse)
						}
					}
				}
			}
			return out
		}
		first, last := a.verses.Visible()
		var out []*Verse
		for row := max(first, 1); row <= last && row <= len(a.surah.Verses); row++ {
			out = append(out, &a.surah.Verses[row-1])
		}
		return out
	case SectionShuffle:
		return a.shuf.verses(a.quran)
	}
	return nil
}

// tick credits the time since the last tick to what is on screen, while
// the window has the focus.
func (a *app) tick(now time.Time, focused bool) {
	r := a.reading
	dt := now.Sub(r.last).Seconds()
	r.last = now
	if !focused || dt <= 0 {
		r.save()
		return
	}
	visible := a.visibleVerses()
	r.credit(visible, min(dt, 2), now)
	if r.dirty {
		a.checkBadges(now)
	}
	r.save()
	// Where reading stopped: the first verse in view.
	if a.section == SectionRead && len(visible) > 0 {
		a.set.LastVerse = visible[0].Index
	}
}

// Texts of the statistics.
var (
	txtStats        = T{"Statistics", "Statistik"}
	txtToday        = T{"Today", "Hari ini"}
	txtVersesRead   = T{"verses read", "ayat dibaca"}
	txtGoal         = T{"Daily goal", "Target harian"}
	txtStreak       = T{"Streak", "Runtutan"}
	txtDays         = T{"days", "hari"}
	txtTotalRead    = T{"Read so far", "Total terbaca"}
	txtOfQuran      = T{"of the Quran's 6,236 verses", "dari 6.236 ayat Al-Qur'an"}
	txtListened     = T{"Listened today", "Didengar hari ini"}
	txtMinutes      = T{"min reading", "menit membaca"}
	txtLast14       = T{"The last 14 days", "14 hari terakhir"}
	txtByJuz        = T{"Progress by juz", "Kemajuan per juz"}
	txtHowCounted   = T{"A verse counts as read after it has been on screen, in reading order, for about half a second a word, while Quranku is the active window. Verses heard to their end count as listened.", "Ayat dihitung terbaca setelah tampil di layar, sesuai urutan baca, kira-kira setengah detik per kata, selama Quranku menjadi jendela aktif. Ayat yang diputar sampai selesai dihitung didengar."}
	txtResetStats   = T{"Reset statistics", "Hapus statistik"}
	txtResetConfirm = T{"Delete all reading statistics? This cannot be undone.", "Hapus semua statistik baca? Tindakan ini tidak bisa dibatalkan."}
	txtCancel       = T{"Cancel", "Batal"}
	txtDelete       = T{"Delete", "Hapus"}
)

// statsPage shows the reading statistics.
func (a *app) statsPage(c *ui.Context) {
	t := c.Theme()
	r := a.reading
	now := a.now()
	today := r.day(now)
	wide := a.contentWidth(c, 1280) >= 1100
	ui.Scroll(c).Grow(1).TrackScroll(&a.statsScroll).Children(func() {
		pageColumn(c, 1280, func() {
			ui.Column(c).Gap(16).Children(func() {
				sectionTitle(c, a.tx(txtStats))
				if ui.Tabs(c, &a.statsTab, a.tx(txtSummary), a.tx(txtRecap), a.tx(txtBadges)).Changed() {
					a.statsScroll.Y = 0
				}
			})
			switch a.statsTab {
			case 1:
				a.recapTab(c)
				return
			case 2:
				a.badgesTab(c)
				return
			}
			// The cards.
			ui.Row(c).Gap(14).AlignItems(ui.Stretch).Children(func() {
				a.statCard(c, a.tx(txtToday), fmt.Sprint(today.Read), a.tx(txtVersesRead), func() {
					goal := max(a.set.DailyGoal, 1)
					ui.Progress(c, min(float64(today.Read)/goal, 1)).Label(a.tx(txtGoal))
					ui.Textf(c, "%s: %d", a.tx(txtGoal), int(goal)).FontSize(11).TextColor(t.TextMuted)
				})
				a.statCard(c, a.tx(txtStreak), fmt.Sprint(r.streak(now)), a.tx(txtDays), nil)
				unique := r.uniqueRead(nil)
				a.statCard(c, a.tx(txtTotalRead), fmt.Sprint(unique), a.tx(txtOfQuran), func() {
					ui.Progress(c, float64(unique)/VerseCount).Label(a.tx(txtTotalRead))
					ui.Textf(c, "%.1f%%", 100*float64(unique)/VerseCount).FontSize(11).TextColor(t.TextMuted)
				})
				a.statCard(c, a.tx(txtListened), fmt.Sprint(today.Listened), fmt.Sprintf("%d %s", int(today.Seconds/60), a.tx(txtMinutes)), nil)
			})
			a.goalStepper(c)
			// Side by side where there is room.
			group := ui.Column(c).Gap(14)
			if wide {
				group = ui.Row(c).Gap(14).AlignItems(ui.Stretch)
			}
			group.Children(func() {
				part := func(fn func()) {
					col := ui.Column(c).Grow(1)
					if wide {
						col.Basis(0)
					}
					col.Children(fn)
				}
				part(func() { a.statsChart(c, now) })
				part(func() { a.juzProgress(c) })
			})
			ui.Text(c, a.tx(txtHowCounted)).FontSize(12).TextColor(t.TextMuted).MaxWidth(640)
			if ui.Button(c, a.tx(txtResetStats)).AlignSelf(ui.Start).Clicked() {
				a.confirmReset = true
			}
		})
	})
	ui.Modal(c, &a.confirmReset, func() {
		ui.Column(c).Gap(14).Width(360).Children(func() {
			ui.Text(c, a.tx(txtResetConfirm))
			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, a.tx(txtCancel)).Clicked() {
					a.confirmReset = false
				}
				if ui.PrimaryButton(c, a.tx(txtDelete)).Clicked() {
					a.resetStats()
					a.confirmReset = false
				}
			})
		})
	})
}

func (a *app) resetStats() {
	r := a.reading
	r.stats = Stats{Days: map[string]*DayStats{}}
	r.bits = make([]byte, (VerseCount+7)/8)
	r.progress, r.today, r.seen = map[int]float64{}, "", nil
	r.dirty = true
	r.save()
}

// Texts of the statistics' tabs.
var (
	txtSummary    = T{"Overview", "Ringkasan"}
	txtRecap      = T{"Recap", "Rekap"}
	txtBadges     = T{"Badges", "Lencana"}
	txtThisWeek   = T{"This week", "Minggu ini"}
	txtThisMonth  = T{"This month", "Bulan ini"}
	txtThisYear   = T{"This year", "Tahun ini"}
	txtRecapEmpty = T{"Nothing read yet in this period. Your recap fills in as you read.", "Belum ada bacaan di periode ini. Rekapmu terisi seiring kamu membaca."}
	txtRecapNote  = T{"Pick a card to share as an image. Shared cards invite others to read; your numbers stay on this computer unless you share them.", "Pilih kartu untuk dibagikan sebagai gambar. Kartu mengajak orang lain ikut membaca; angkamu tetap di komputer ini kecuali kamu membagikannya."}
	txtBadgesOf   = T{"%d of %d badges", "%d dari %d lencana"}
	txtNotYet     = T{"Not yet", "Belum diraih"}
)

// recapTab shows the period's reading as cards to share.
func (a *app) recapTab(c *ui.Context) {
	t := c.Theme()
	ui.Segmented(c, &a.recapPeriod, a.tx(txtThisWeek), a.tx(txtThisMonth), a.tx(txtThisYear)).Label(a.tx(txtRecap)).AlignSelf(ui.Start)
	r := a.recap(a.recapPeriod, a.now())
	if r.Read == 0 && r.Listened == 0 {
		ui.Column(c).Gap(14).Padding(28).Radius(14).Background(t.Surface).Border(1, t.Border).Children(func() {
			ui.Text(c, a.tx(txtRecapEmpty)).FontSize(14).TextColor(t.TextMuted)
			if labeledButton(c, iconRead, a.tx(txtContinue), true).AlignSelf(ui.Start).Clicked() {
				a.continueReading(a.set.Mode)
			}
		})
		return
	}
	ui.Text(c, a.tx(txtRecapNote)).FontSize(13).TextColor(t.TextMuted).MaxWidth(720)
	cards := []ShareCard{{Kind: CardSummary, Recap: r}}
	if len(r.TopSurahs) > 0 {
		cards = append(cards, ShareCard{Kind: CardSurahs, Recap: r})
		cards = append(cards, ShareCard{Kind: CardTime, Recap: r})
	}
	if r.LongestStreak > 0 || r.Reciter != nil {
		cards = append(cards, ShareCard{Kind: CardHabit, Recap: r})
	}
	style := a.cardStyles()[0]
	ui.Row(c).Gap(20).Wrap().Children(func() {
		for i, card := range cards {
			ui.Column(c.Key(i)).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Radius(8).Clip().Shadow(0, 4, 16, 0, ui.RGBA(0, 0, 0, 0.12)).Children(func() {
					a.cardView(c, card, style, FormatStory, 0.42)
				})
				if labeledButton(c, iconShare, a.tx(txtShare), false).Clicked() {
					a.openShare(card)
				}
			})
		}
	})
}

// badgesTab shows the badges, earned or not.
func (a *app) badgesTab(c *ui.Context) {
	t := c.Theme()
	earned := a.reading.stats.Badges
	ui.Textf(c, a.tx(txtBadgesOf), len(earned), len(badges)).FontSize(14).TextColor(t.TextMuted)
	ui.Grid(c).Columns(a.indexColumns()).Gap(14).Children(func() {
		for _, b := range badges {
			day, ok := earned[b.ID]
			tile := ui.Column(c.Key(b.ID)).Padding(18).Gap(8).Radius(14).Background(t.Surface).Border(1, t.Border)
			if !ok {
				tile.Opacity(0.55)
			}
			tile.Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					circle := ui.Column(c).Size(44, 44).Radius(22).Justify(ui.Center).AlignItems(ui.Center)
					col := t.TextMuted
					if ok {
						circle.Background(t.Accent.Alpha(0.14))
						col = t.Accent
					} else {
						circle.Background(t.Background)
					}
					circle.Children(func() { ui.Icon(c, b.Icon).FontSize(22).TextColor(col) })
					ui.Column(c).Grow(1).Gap(2).Children(func() {
						ui.Text(c, a.tx(b.Name)).FontSize(15).Bold()
						note := a.tx(txtNotYet)
						if when, err := time.Parse("2006-01-02", day); ok && err == nil {
							note = fmt.Sprintf(a.tx(txtBadgeEarned), a.longDate(when))
						}
						ui.Text(c, note).FontSize(12).TextColor(t.TextMuted)
					})
				})
				ui.Text(c, a.tx(b.Desc)).FontSize(13).TextColor(t.TextMuted).LineHeight(1.4)
				if ok && labeledButton(c, iconShare, a.tx(txtShare), false).AlignSelf(ui.Start).Clicked() {
					a.openShare(ShareCard{Kind: CardBadge, Badge: b})
				}
			})
		}
	})
}

// statCard is a number with its title and unit, and what more builds.
func (a *app) statCard(c *ui.Context, title, value, unit string, more func()) {
	t := c.Theme()
	ui.Column(c.Key(title)).Grow(1).Basis(0).Padding(16).Gap(6).Radius(12).
		Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, title).FontSize(12).Bold().TextColor(t.TextMuted)
		ui.Text(c, value).FontSize(30).Bold().TextColor(t.Accent).FontFeatures("tnum")
		ui.Text(c, unit).FontSize(12).TextColor(t.TextMuted)
		if more != nil {
			more()
		}
	})
}

// statsChart draws the verses read on each of the last 14 days.
func (a *app) statsChart(c *ui.Context, now time.Time) {
	t := c.Theme()
	const days, height = 14, 120
	most := 1
	for i := range days {
		if d := a.reading.stats.Days[dayKey(now.AddDate(0, 0, -i))]; d != nil {
			most = max(most, d.Read)
		}
	}
	ui.Column(c).Gap(10).Padding(16).Radius(12).Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, a.tx(txtLast14)).FontSize(13).Bold()
		ui.Row(c).Gap(6).AlignItems(ui.End).Height(height + 34).Children(func() {
			for i := days - 1; i >= 0; i-- {
				day := now.AddDate(0, 0, -i)
				n := 0
				if d := a.reading.stats.Days[dayKey(day)]; d != nil {
					n = d.Read
				}
				ui.Column(c.Key(i)).Grow(1).Basis(0).Gap(4).AlignItems(ui.Center).Justify(ui.End).Children(func() {
					if n > 0 {
						ui.Text(c, fmt.Sprint(n)).FontSize(10).TextColor(t.TextMuted)
					}
					bar := ui.Box(c).Width(18).Height(max(2, float32(height*n/most))).Radius(4)
					if i == 0 {
						bar.Background(t.Accent)
					} else {
						bar.Background(t.Accent.Alpha(0.45))
					}
					ui.Text(c, day.Format("2")).FontSize(10).TextColor(t.TextMuted)
				})
			}
		})
	})
}

// juzProgress shows how much of each juz has been read.
func (a *app) juzProgress(c *ui.Context) {
	t := c.Theme()
	total, read := [31]int{}, [31]int{}
	for _, v := range a.quran.All {
		total[v.Juz]++
		if a.reading.isRead(v.Index) {
			read[v.Juz]++
		}
	}
	ui.Column(c).Gap(10).Padding(16).Radius(12).Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, a.tx(txtByJuz)).FontSize(13).Bold()
		ui.Grid(c).Columns(10).Gap(6).Children(func() {
			for j := 1; j <= 30; j++ {
				f := float32(read[j]) / float32(max(total[j], 1))
				cell := ui.Column(c.Key(j)).Height(46).Radius(8).Justify(ui.Center).AlignItems(ui.Center).
					Background(t.Accent.Alpha(0.06 + 0.6*f)).Tooltip(fmt.Sprintf("%s %d: %d/%d", a.tx(txtJuz), j, read[j], total[j]))
				cell.Children(func() {
					txt := t.Text
					if f > 0.6 {
						txt = t.AccentText
					}
					ui.Text(c, fmt.Sprint(j)).FontSize(13).Bold().TextColor(txt)
					ui.Textf(c, "%d%%", int(100*f)).FontSize(10).TextColor(txt.Alpha(0.75))
				})
			}
		})
	})
}
