package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// phoneTester shows the app on an iPhone's screen: 393×852 points, its
// status bar and home indicator around the safe area.
func phoneTester(a *app) *ui.Tester {
	a.touch = true
	tt := ui.NewTester(a.view, 393, 852)
	tt.SetInsets(59, 0, 34, 0)
	tt.Frame()
	return tt
}

// TestPhoneScreens builds each section on a phone's screen. With
// QURANKU_SHOTS set to a directory, it saves them there, to look at.
func TestPhoneScreens(t *testing.T) {
	a := newApp(loadTestQuran(t), defaultSettings(), "")
	a.set.Onboarded = true
	tt := phoneTester(a)
	shots := []struct {
		name    string
		section int
		mode    int
	}{
		{"home", SectionHome, 0},
		{"read-verses", SectionRead, ModeVerses},
		{"read-mushaf", SectionRead, ModeMushaf},
		{"read-mushaf-50", SectionRead, ModeMushaf},
		{"recitations", SectionAudio, 0},
		{"stats", SectionStats, 0},
		{"settings", SectionSettings, 0},
		{"shuffle", SectionShuffle, 0},
	}
	save := func(name string) {
		if dir := os.Getenv("QURANKU_SHOTS"); dir != "" {
			f, err := os.Create(filepath.Join(dir, "phone-"+name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			png.Encode(f, tt.Image())
			f.Close()
		}
	}
	for _, s := range shots {
		a.section = s.section
		if s.section == SectionRead {
			a.set.Mode = s.mode
		}
		if s.name == "read-mushaf-50" {
			a.goPage(50)
		}
		for range 3 {
			tt.Frame()
		}
		save(s.name)
	}
	// The statistics' recap, after some reading, and the settings' tabs
	// scrolled to their end.
	for n := 1; n <= 12; n++ {
		a.reading.markRead(a.quran.Verse(2, n), a.now())
	}
	a.section, a.statsTab = SectionStats, 1
	for range 3 {
		tt.Frame()
	}
	save("stats-recap")
	a.statsTab = 0
	a.section = SectionSettings
	for tab := TabLook; tab <= TabGeneral; tab++ {
		a.settingsTab = tab
		for range 3 {
			tt.Frame()
		}
		save(fmt.Sprint("settings-tab-", tab))
	}
	a.settingsTab = TabLook

	// The sheets, risen.
	sheet := func(name string, open *bool) {
		*open = true
		for range 3 {
			tt.Frame()
			tt.Advance(time.Second)
		}
		tt.Frame()
		save(name)
		*open = false
		for range 3 {
			tt.Frame()
			tt.Advance(time.Second)
		}
	}
	a.section, a.set.Mode = SectionRead, ModeVerses
	sheet("sheet-picker", &a.picker)
	sheet("sheet-display", &a.quick)
	a.focus = rules[0]
	sheet("sheet-guide", &a.guideOpen)
	a.share = &ShareCard{Kind: CardVerse, Verse: a.quran.Verse(2, 255)}
	sheet("sheet-share", &a.shareOpen)

	// The introduction, on a first run.
	a.set.Onboarded = false
	for range 3 {
		tt.Frame()
	}
	save("onboarding")
	a.set.Onboarded = true
	a.section, a.set.Mode = SectionRead, ModeVerses
	a.open(a.quran.Surahs[1])
	for range 3 {
		tt.Frame()
	}
	save("read-baqarah")
	a.set.Onboarded = false
	for step := 1; step < onboardSteps; step++ {
		a.onboardStep = step
		for range 3 {
			tt.Frame()
		}
		save(fmt.Sprint("onboarding-", step))
	}
}

// fingerDrag moves a finger from one point to another, and lets what it
// threw come to rest.
func fingerDrag(tt *ui.Tester, x0, y0, x1, y1 float32) {
	tt.FingerDown(x0, y0)
	for i := 1; i <= 10; i++ {
		tt.Advance(16 * time.Millisecond)
		f := float32(i) / 10
		tt.FingerMove(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	tt.FingerUp(x1, y1)
	for range 30 {
		tt.Advance(50 * time.Millisecond)
		tt.Frame()
	}
}

// TestPhoneGestures scrolls with a finger, and turns the Mushaf's pages as
// a book read right to left.
func TestPhoneGestures(t *testing.T) {
	a := newApp(loadTestQuran(t), defaultSettings(), "")
	a.set.Onboarded = true
	tt := phoneTester(a)
	fingerDrag(tt, 200, 600, 200, 300)
	if a.homeScroll.Y <= 0 {
		t.Errorf("home did not scroll: Y %v", a.homeScroll.Y)
	}

	// Switching to the Mushaf opens it on the page being read; then a jump.
	a.section, a.set.Mode = SectionRead, ModeMushaf
	tt.Frame()
	a.goPage(10)
	for range 3 {
		tt.Frame()
	}
	turn := func(dx float32, want int) {
		t.Helper()
		fingerDrag(tt, 200, 500, 200+dx, 500)
		if a.set.Page != want {
			t.Fatalf("swiped %v: page %d, want %d", dx, a.set.Page, want)
		}
	}
	turn(250, 11)  // to the right: the next page, from the left
	turn(-250, 10) // back
	// A jump, as the picker makes, and a swipe from there.
	a.goPage(100)
	for range 3 {
		tt.Frame()
	}
	turn(250, 101)
}

// TestTabletScreens builds the sections on an iPad's screen, across and
// upright. With QURANKU_SHOTS set to a directory, it saves them there.
func TestTabletScreens(t *testing.T) {
	for _, o := range []struct {
		name string
		w, h int
		rail bool
	}{{"landscape", 1180, 820, false}, {"portrait", 820, 1180, true}} {
		a := newApp(loadTestQuran(t), defaultSettings(), "")
		a.set.Onboarded, a.touch = true, true
		for n := 1; n <= 12; n++ {
			a.reading.markRead(a.quran.Verse(2, n), a.now())
		}
		tt := ui.NewTester(a.view, o.w, o.h)
		tt.SetInsets(24, 0, 20, 0)
		tt.Frame()
		if a.phone || a.rail != o.rail {
			t.Fatalf("%s: phone %v, rail %v", o.name, a.phone, a.rail)
		}
		for _, s := range []struct {
			name    string
			section int
			mode    int
		}{
			{"home", SectionHome, 0},
			{"read-verses", SectionRead, ModeVerses},
			{"read-mushaf", SectionRead, ModeMushaf},
			{"recitations", SectionAudio, 0},
			{"stats", SectionStats, 0},
			{"settings", SectionSettings, 0},
		} {
			a.section = s.section
			if s.section == SectionRead {
				a.set.Mode = s.mode
				if s.mode == ModeMushaf {
					tt.Frame()
					a.goPage(50)
				}
			}
			for range 3 {
				tt.Frame()
			}
			if dir := os.Getenv("QURANKU_SHOTS"); dir != "" {
				f, err := os.Create(filepath.Join(dir, "tablet-"+o.name+"-"+s.name+".png"))
				if err != nil {
					t.Fatal(err)
				}
				png.Encode(f, tt.Image())
				f.Close()
			}
		}
	}
}
