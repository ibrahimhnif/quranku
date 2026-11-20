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

// screen is a device the app is laid out for in TestScreens.
type screen struct {
	name        string
	w, h        int
	top, bottom float32 // its safe area's insets
	phone, rail bool    // the layout it should get
}

var screens = []screen{
	{"phone", 393, 852, 59, 34, true, false},
	{"tablet-portrait", 820, 1180, 24, 20, false, true},
	{"tablet-landscape", 1180, 820, 24, 20, false, false},
}

// TestScreens builds every screen on a phone and a tablet, across and
// upright, with something read. With QURANKU_SHOTS set to a directory, it
// saves each there, to look at.
func TestScreens(t *testing.T) {
	for _, d := range screens {
		t.Run(d.name, func(t *testing.T) {
			a := newApp(loadTestQuran(t), defaultSettings(), "")
			a.set.Onboarded, a.touch = true, true
			for n := 1; n <= 30; n++ {
				a.reading.markRead(a.quran.Verse(2, n), a.now())
			}
			a.reading.markListened(7, a.now())
			tt := ui.NewTester(a.view, d.w, d.h)
			tt.SetInsets(d.top, 0, d.bottom, 0)
			settle := func() {
				for range 4 {
					tt.Frame()
					tt.Advance(time.Second)
				}
				tt.Frame()
			}
			settle()
			if a.phone != d.phone || a.rail != d.rail {
				t.Fatalf("phone %v, rail %v", a.phone, a.rail)
			}
			save := func(name string) {
				if dir := os.Getenv("QURANKU_SHOTS"); dir != "" {
					f, err := os.Create(filepath.Join(dir, d.name+"-"+name+".png"))
					if err != nil {
						t.Fatal(err)
					}
					png.Encode(f, tt.Image())
					f.Close()
				}
			}
			show := func(name string, set func()) {
				set()
				settle()
				save(name)
			}
			show("home", func() { a.section = SectionHome })
			show("read-verses", func() { a.section, a.set.Mode = SectionRead, ModeVerses; a.open(a.quran.Surahs[1]) })
			show("read-mushaf", func() { a.set.Mode = ModeMushaf })
			show("read-mushaf-50", func() { a.goPage(50) })
			a.set.Mode = ModeVerses
			show("shuffle", func() { a.section = SectionShuffle; a.shuffleNext() })
			show("stats", func() { a.section, a.statsTab = SectionStats, 0 })
			show("stats-recap", func() { a.statsTab = 1 })
			show("stats-badges", func() { a.statsTab = 2 })
			show("recitations", func() { a.section = SectionAudio })
			for tab := TabLook; tab <= TabGeneral; tab++ {
				show(fmt.Sprint("settings-", tab), func() { a.section, a.settingsTab = SectionSettings, tab })
			}
			// What comes up over the reader.
			a.section, a.set.Mode = SectionRead, ModeVerses
			settle()
			open := func(name string, flag *bool) {
				show(name, func() { *flag = true })
				*flag = false
				settle()
			}
			open("picker", &a.picker)
			open("display", &a.quick)
			a.focus = rules[0]
			if a.sideGuide() {
				open("guide", &a.set.Guide)
				a.set.Guide = true
			} else {
				open("guide", &a.guideOpen)
			}
			a.focus = nil
			a.openShare(ShareCard{Kind: CardVerse, Verse: a.quran.Verse(1, 1)})
			show("share", func() {})
			a.shareOpen = false
			settle()
			a.share = nil
			// The introduction.
			a.set.Onboarded = false
			for step := 0; step < onboardSteps; step++ {
				show(fmt.Sprint("onboarding-", step), func() { a.onboardStep = step })
			}
		})
	}
}
