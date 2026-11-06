package main

import (
	"sort"
	"time"

	"github.com/egoist/mygo/ui"
)

// Badge marks a habit kept: earned once, kept with its day.
type Badge struct {
	ID         string
	Name, Desc T
	Icon       *ui.SVG
	earned     func(f badgeFacts) bool
}

// badgeFacts are what the badges are judged on, gathered once.
type badgeFacts struct {
	read, listened int // all days
	unique         int
	longestStreak  int
	juzDone        [31]bool
	kahfiFriday    bool
	dawnDays       int // days with verses read from 3 to 6 in the morning
	goalDays       int // days the goal was met
}

var badges = []*Badge{
	{ID: "first", Icon: icon("footprints"),
		Name: T{"First step", "Langkah pertama"}, Desc: T{"Read your first verse in Quranku.", "Membaca ayat pertama di Quranku."},
		earned: func(f badgeFacts) bool { return f.read >= 1 }},
	{ID: "streak3", Icon: icon("flame"),
		Name: T{"Three days", "Tiga hari"}, Desc: T{"Read three days in a row.", "Membaca tiga hari berturut-turut."},
		earned: func(f badgeFacts) bool { return f.longestStreak >= 3 }},
	{ID: "streak7", Icon: icon("calendar-check"),
		Name: T{"A full week", "Sepekan penuh"}, Desc: T{"Read seven days in a row.", "Membaca tujuh hari berturut-turut."},
		earned: func(f badgeFacts) bool { return f.longestStreak >= 7 }},
	{ID: "streak30", Icon: icon("medal"),
		Name: T{"A month of habit", "Sebulan istiqamah"}, Desc: T{"Read thirty days in a row.", "Membaca tiga puluh hari berturut-turut."},
		earned: func(f badgeFacts) bool { return f.longestStreak >= 30 }},
	{ID: "goal7", Icon: icon("target"),
		Name: T{"Goal keeper", "Tepat target"}, Desc: T{"Met the daily goal on seven days.", "Mencapai target harian di tujuh hari."},
		earned: func(f badgeFacts) bool { return f.goalDays >= 7 }},
	{ID: "dawn", Icon: icon("sunrise"),
		Name: T{"Dawn reader", "Pembaca Subuh"}, Desc: T{"Read at dawn on seven days.", "Membaca di waktu Subuh di tujuh hari."},
		earned: func(f badgeFacts) bool { return f.dawnDays >= 7 }},
	{ID: "kahfi", Icon: icon("moon"),
		Name: T{"Friday's Al-Kahf", "Al-Kahfi hari Jumat"}, Desc: T{"Read Al-Kahf on a Friday.", "Membaca Al-Kahfi di hari Jumat."},
		earned: func(f badgeFacts) bool { return f.kahfiFriday }},
	{ID: "juz", Icon: icon("award"),
		Name: T{"A whole juz", "Satu juz tuntas"}, Desc: T{"Read every verse of a juz.", "Membaca seluruh ayat satu juz."},
		earned: func(f badgeFacts) bool {
			for _, d := range f.juzDone[1:] {
				if d {
					return true
				}
			}
			return false
		}},
	{ID: "amma", Icon: icon("gem"),
		Name: T{"Juz 'Amma", "Juz 'Amma"}, Desc: T{"Read all of the 30th juz.", "Membaca seluruh juz ke-30."},
		earned: func(f badgeFacts) bool { return f.juzDone[30] }},
	{ID: "listener", Icon: icon("headphones"),
		Name: T{"Good listener", "Pendengar setia"}, Desc: T{"Listened to 500 verses recited.", "Mendengarkan 500 ayat murattal."},
		earned: func(f badgeFacts) bool { return f.listened >= 500 }},
	{ID: "thousand", Icon: icon("trophy"),
		Name: T{"A thousand verses", "Seribu ayat"}, Desc: T{"Read 1,000 different verses.", "Membaca 1.000 ayat berbeda."},
		earned: func(f badgeFacts) bool { return f.unique >= 1000 }},
	{ID: "khatam", Icon: icon("crown"),
		Name: T{"Khatam", "Khatam"}, Desc: T{"Read every verse of the Quran.", "Membaca seluruh ayat Al-Qur'an."},
		earned: func(f badgeFacts) bool { return f.unique >= VerseCount }},
}

func badgeByID(id string) *Badge {
	for _, b := range badges {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// facts gathers what the badges are judged on.
func (a *app) facts() badgeFacts {
	r := a.reading
	var f badgeFacts
	f.unique = r.uniqueRead(nil)
	keys := make([]string, 0, len(r.stats.Days))
	for k, d := range r.stats.Days {
		keys = append(keys, k)
		f.read += d.Read
		f.listened += d.Listened
		if d.Read >= int(a.set.DailyGoal) {
			f.goalDays++
		}
		if d.Hours[3]+d.Hours[4]+d.Hours[5] > 0 {
			f.dawnDays++
		}
		if t, err := time.Parse("2006-01-02", k); err == nil && t.Weekday() == time.Friday && d.Surahs[18] >= 10 {
			f.kahfiFriday = true
		}
	}
	// The longest run of days with verses read.
	streak := 0
	var prev time.Time
	sort.Strings(keys)
	for _, k := range keys {
		d := r.stats.Days[k]
		t, err := time.Parse("2006-01-02", k)
		if err != nil || d.Read == 0 {
			streak = 0
			continue
		}
		if !prev.IsZero() && t.Sub(prev) <= 25*time.Hour {
			streak++
		} else {
			streak = 1
		}
		prev = t
		f.longestStreak = max(f.longestStreak, streak)
	}
	var total, read [31]int
	for _, v := range a.quran.All {
		total[v.Juz]++
		if r.isRead(v.Index) {
			read[v.Juz]++
		}
	}
	for j := 1; j <= 30; j++ {
		f.juzDone[j] = total[j] > 0 && read[j] == total[j]
	}
	return f
}

// checkBadges awards the badges newly earned, to show them.
func (a *app) checkBadges(now time.Time) {
	r := a.reading
	f := a.facts()
	for _, b := range badges {
		if _, ok := r.stats.Badges[b.ID]; ok || !b.earned(f) {
			continue
		}
		if r.stats.Badges == nil {
			r.stats.Badges = map[string]string{}
		}
		r.stats.Badges[b.ID] = dayKey(now)
		r.dirty = true
		a.newBadges = append(a.newBadges, b)
	}
}
