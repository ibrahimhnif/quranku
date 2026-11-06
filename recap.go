package main

import (
	"fmt"
	"sort"
	"time"
)

// Recap periods.
const (
	PeriodWeek = iota
	PeriodMonth
	PeriodYear
)

// SurahCount is a surah and the verses of it read.
type SurahCount struct {
	Surah *Surah
	N     int
}

// Recap sums the reading of a period, as the recap cards tell it.
type Recap struct {
	Period         int
	From, To       time.Time // To is the day after the last
	Read, Listened int
	Minutes        int
	ActiveDays     int
	LongestStreak  int
	TopSurahs      []SurahCount // the most read first, five at most
	Hours          [24]int
	Persona        Persona
	Reciter        *Reciter // the most heard, or nil
	ReciterVerses  int
}

// Persona names when someone reads most.
type Persona struct {
	Name     T
	From, To int // hours
}

var personas = []Persona{
	{T{"Dawn reader", "Pembaca Subuh"}, 3, 6},
	{T{"Morning reader", "Pembaca Pagi"}, 6, 11},
	{T{"Midday reader", "Pembaca Siang"}, 11, 15},
	{T{"Afternoon reader", "Pembaca Sore"}, 15, 18},
	{T{"Night reader", "Pembaca Malam"}, 18, 27}, // to 3 in the morning
}

// periodStart is the first day of the period holding now: the week from
// Monday, the month, or the year.
func periodStart(p int, now time.Time) time.Time {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch p {
	case PeriodWeek:
		back := (int(day.Weekday()) + 6) % 7 // days since Monday
		return day.AddDate(0, 0, -back)
	case PeriodMonth:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
}

// recap sums the period holding now.
func (a *app) recap(p int, now time.Time) Recap {
	r := Recap{Period: p, From: periodStart(p, now)}
	r.To = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	surahs := map[int]int{}
	reciters := map[int]int{}
	streak := 0
	for d := r.From; d.Before(r.To); d = d.AddDate(0, 0, 1) {
		day := a.reading.stats.Days[dayKey(d)]
		if day == nil || day.Read == 0 {
			streak = 0
			if day == nil {
				continue
			}
		} else {
			streak++
			r.ActiveDays++
			r.LongestStreak = max(r.LongestStreak, streak)
		}
		r.Read += day.Read
		r.Listened += day.Listened
		r.Minutes += int(day.Seconds / 60)
		for s, n := range day.Surahs {
			surahs[s] += n
		}
		for h, n := range day.Hours {
			r.Hours[h] += n
		}
		for id, n := range day.Reciters {
			reciters[id] += n
		}
	}
	for s, n := range surahs {
		if s >= 1 && s <= len(a.quran.Surahs) {
			r.TopSurahs = append(r.TopSurahs, SurahCount{a.quran.Surahs[s-1], n})
		}
	}
	sort.Slice(r.TopSurahs, func(i, j int) bool {
		if r.TopSurahs[i].N != r.TopSurahs[j].N {
			return r.TopSurahs[i].N > r.TopSurahs[j].N
		}
		return r.TopSurahs[i].Surah.ID < r.TopSurahs[j].Surah.ID
	})
	if len(r.TopSurahs) > 5 {
		r.TopSurahs = r.TopSurahs[:5]
	}
	r.Persona = personaOf(r.Hours)
	for id, n := range reciters {
		if rec := reciterByID(id); rec != nil && (n > r.ReciterVerses || n == r.ReciterVerses && r.Reciter != nil && id < r.Reciter.ID) {
			r.Reciter, r.ReciterVerses = rec, n
		}
	}
	return r
}

// personaOf names the part of the day with the most verses read.
func personaOf(hours [24]int) Persona {
	best, most := personas[0], -1
	for _, p := range personas {
		n := 0
		for h := p.From; h < p.To; h++ {
			n += hours[h%24]
		}
		if n > most {
			best, most = p, n
		}
	}
	return best
}

// periodName names a recap's period: "November 2026", "2026", or the
// week's days.
func (a *app) periodName(r Recap) string {
	switch r.Period {
	case PeriodWeek:
		last := r.To.AddDate(0, 0, -1)
		return fmt.Sprintf("%d–%d %s %d", r.From.Day(), last.Day(), a.monthName(last.Month()), last.Year())
	case PeriodMonth:
		return fmt.Sprintf("%s %d", a.monthName(r.From.Month()), r.From.Year())
	}
	return fmt.Sprint(r.From.Year())
}

func (a *app) monthName(m time.Month) string {
	if a.set.Lang == Indonesian {
		return idMonths[m-1]
	}
	return m.String()
}
