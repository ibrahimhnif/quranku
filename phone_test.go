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
	a.phone = true
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
	for step := 1; step < onboardSteps; step++ {
		a.onboardStep = step
		for range 3 {
			tt.Frame()
		}
		save(fmt.Sprint("onboarding-", step))
	}
}
