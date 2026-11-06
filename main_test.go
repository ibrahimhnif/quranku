package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func segmentKeys(segs []Segment) string {
	var got []string
	for _, s := range segs {
		key := "-"
		if s.Rule != nil {
			key = s.Rule.Key
		}
		got = append(got, key+":"+s.Text)
	}
	return strings.Join(got, "|")
}

func TestVerseSegments(t *testing.T) {
	text := []rune("رَيْبَ ۛ فِيهِ")
	segs := verseSegments(text, []tajweedRange{
		{0, 3, rulesByKey["qalaqah"]},
		{1, 2, rulesByKey["slnt"]}, // starts inside: takes its part
	})
	// The pause mark joins the word before, uncolored.
	want := "qalaqah:ر|slnt:َ|qalaqah:ي|-:ْبَۛ فِيهِ"
	if got := segmentKeys(segs); got != want {
		t.Errorf("segments %q, want %q", got, want)
	}
}

func TestQuranText(t *testing.T) {
	q, err := loadQuran()
	if err != nil {
		t.Fatal(err)
	}
	// Verse 1 of Al-Baqarah without the basmalah Tanzil opens it with.
	if got := segmentText(q.Verse(2, 1).Segments); got != "الٓمٓ" {
		t.Errorf("2:1 is %q", got)
	}
	if got := segmentText(q.Verse(1, 1).Segments); !strings.HasPrefix(got, "بِسْمِ") {
		t.Errorf("1:1 is %q", got)
	}
	// No pause mark stands alone as a word.
	for _, v := range q.All {
		for _, w := range splitWords(v.Segments) {
			if r := []rune(segmentText(w)); len(r) == 1 && isPauseMark(r[0]) {
				t.Fatalf("%s has a pause mark alone", v.Key)
			}
		}
	}
	if q.Verse(2, 255).Pickthall == "" || !strings.HasPrefix(q.Verse(1, 2).Pickthall, "Praise be to Allah") {
		t.Errorf("Pickthall 1:2 is %q", q.Verse(1, 2).Pickthall)
	}
	rules := 0
	for _, s := range q.Surahs {
		for _, n := range s.RuleCounts {
			rules += n
		}
	}
	if rules < 50000 {
		t.Errorf("only %d tajweed segments", rules)
	}
}

func segmentText(segs []Segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	return b.String()
}

func TestLoadQuran(t *testing.T) {
	q, err := loadQuran()
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Surahs) != 114 {
		t.Fatalf("%d surahs", len(q.Surahs))
	}
	total := 0
	for _, s := range q.Surahs {
		if len(s.Verses) != s.VerseCount {
			t.Errorf("surah %d has %d verses, want %d", s.ID, len(s.Verses), s.VerseCount)
		}
		total += len(s.Verses)
	}
	if total != 6236 {
		t.Errorf("%d verses", total)
	}
	if got := q.Search("baqar"); len(got) != 1 || got[0].ID != 2 {
		t.Errorf("search baqar: %v", got)
	}
	if got := q.Search("114"); len(got) != 1 || got[0].ID != 114 {
		t.Errorf("search 114: %v", got)
	}
}

func TestView(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Lang, set.Onboarded = English, true
	a := newApp(q, set, "")
	tt := ui.NewTester(a.view, 1320, 880)
	// Home: the continue point, then the surahs.
	if !tt.HasText("Assalamu alaikum") || !tt.HasText("Begin with Al-Fatihah") {
		t.Fatalf("home texts %q", tt.Texts())
	}
	if err := tt.Click("Al-Baqarah"); err != nil {
		t.Fatal(err)
	}
	if a.section != SectionRead || a.surah.ID != 2 {
		t.Fatalf("section %d, surah %d after clicking Al-Baqarah", a.section, a.surah.ID)
	}
	if err := tt.Click("Qalqalah"); err != nil {
		t.Fatal(err)
	}
	if a.focus == nil || a.focus.Key != "qalaqah" {
		t.Fatalf("focus %v after clicking Qalqalah", a.focus)
	}
	if err := tt.Click("Show all"); err != nil {
		t.Fatal(err)
	}
	if a.focus != nil {
		t.Errorf("focus %v after Show all", a.focus.Key)
	}
	a.set.Lang = Indonesian
	tt.Frame()
	if !tt.HasText("Panduan Tajwid") || !tt.HasText("2. Al-Baqarah") {
		t.Errorf("texts in Indonesian: %q", tt.Texts())
	}
	// Esc goes home, which now continues at Al-Baqarah.
	tt.Key(0, ui.KeyEscape)
	if a.section != SectionHome || a.lastVerse().Surah != 2 {
		t.Errorf("after Esc: section %d, last verse %s", a.section, a.lastVerse().Key)
	}
	if err := tt.Click("Lanjutkan"); err != nil {
		t.Fatal(err)
	}
	if a.section != SectionRead || a.surah.ID != 2 {
		t.Errorf("after Continue: section %d, surah %d", a.section, a.surah.ID)
	}
	if path := os.Getenv("SCREENSHOT"); path != "" {
		tt.Frame()
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOnboarding(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Lang = English
	a := newApp(q, set, "")
	tt := ui.NewTester(a.view, 1320, 880)
	if !tt.HasText("Begin") || !tt.HasText("Skip, use the defaults") {
		t.Fatalf("welcome texts %q", tt.Texts())
	}
	for _, step := range []struct{ click, then string }{
		{"Begin", "How do you like to read?"},
		{"Mushaf", ""},
		{"Next", "Make it comfortable to read"},
		{"Night", ""},
		{"Next", "Choose a reciter"},
		{"Abdur-Rahman as-Sudais", ""},
		{"Next", "How much a day?"},
		{"30 verses", ""},
		{"Start reading", "Assalamu alaikum"},
	} {
		if err := tt.Click(step.click); err != nil {
			t.Fatalf("%s: %v (texts %q)", step.click, err, tt.Texts())
		}
		if step.then != "" && !tt.HasText(step.then) {
			t.Fatalf("after %s: no %q in %q", step.click, step.then, tt.Texts())
		}
	}
	if !a.set.Onboarded || a.set.Mode != ModeMushaf || a.set.Theme != "night" || a.set.Reciter != 3 || a.set.DailyGoal != 30 || a.section != SectionHome {
		t.Errorf("after onboarding: %+v, section %d", a.set, a.section)
	}
	// Skipping keeps the defaults.
	b := newApp(q, set, "")
	tb := ui.NewTester(b.view, 1320, 880)
	if err := tb.Click("Skip, use the defaults"); err != nil {
		t.Fatal(err)
	}
	if !b.set.Onboarded || b.set.Theme != "auto" || b.section != SectionHome {
		t.Errorf("after skipping: %+v", b.set)
	}
}

func TestMushafLayout(t *testing.T) {
	q, err := loadQuran()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range q.Pages[1:] {
		empty := 0
		for _, l := range p.Lines {
			if l.Kind == LineEmpty {
				empty++
			}
		}
		// The first two pages hold 8 lines; every other page is full.
		want := 0
		if p.Number <= 2 {
			want = 7
		}
		if empty != want {
			t.Errorf("page %d has %d empty lines, want %d", p.Number, empty, want)
		}
		if p.Juz == 0 {
			t.Errorf("page %d has no juz", p.Number)
		}
	}
	if q.Surahs[1].Page != 2 || q.Surahs[113].Page != 604 {
		t.Errorf("Al-Baqarah on page %d, An-Nas on %d", q.Surahs[1].Page, q.Surahs[113].Page)
	}
	if l := q.Pages[604].Lines[9]; l.Kind != LineHeader || l.Surah.ID != 114 {
		t.Errorf("page 604 line 10: %+v", l)
	}
}

func TestMushafMode(t *testing.T) {
	q, err := loadQuran()
	if err != nil {
		t.Fatal(err)
	}
	set := defaultSettings()
	set.Lang, set.Onboarded = English, true
	a := newApp(q, set, "")
	tt := ui.NewTester(a.view, 1280, 860)
	a.query = "mulk"
	tt.Frame()
	if err := tt.Click("Al-Mulk"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Mushaf"); err != nil {
		t.Fatal(err)
	}
	if a.set.Mode != ModeMushaf || a.set.Page != q.Surahs[66].Page || a.surah.ID != 67 {
		t.Fatalf("mode %d page %d surah %d after choosing Mushaf on Al-Mulk", a.set.Mode, a.set.Page, a.surah.ID)
	}
	if !tt.HasText("سُورَةُ الملك") {
		t.Errorf("no header of Al-Mulk on page %d: %q", a.set.Page, tt.Texts())
	}
	page := a.set.Page
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	if a.set.Page != page+1 {
		t.Errorf("page %d after Next, want %d", a.set.Page, page+1)
	}
	tt.Key(0, ui.KeyRight) // back, as the book reads right to left
	if a.set.Page != page {
		t.Errorf("page %d after Right, want %d", a.set.Page, page)
	}
	a.set.Spread = true
	tt.Frame()
	if a.set.Page%2 != 1 {
		t.Errorf("a spread starts on page %d", a.set.Page)
	}
	if err := tt.Click("Verses"); err != nil {
		t.Fatal(err)
	}
	if a.set.Mode != ModeVerses {
		t.Errorf("mode %d after choosing Verses", a.set.Mode)
	}
}

func TestSettingsFile(t *testing.T) {
	f := newJSONFile(t.TempDir() + "/settings.json")
	want := defaultSettings()
	want.Theme, want.Font, want.Page, want.Mode = "custom-1", 1, 300, ModeMushaf
	want.Themes = []CustomTheme{{ID: "custom-1", Name: "Senja", Page: "#FFF4E6", Surface: "#FBE7CF", Text: "#3A2414", Accent: "#C2410C"}}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(f.path, data); err != nil {
		t.Fatal(err)
	}
	if got := loadSettings(f); !got.equal(want) {
		t.Errorf("loaded %+v, want %+v", got, want)
	}
	bad := Settings{Theme: "gone", Font: -1, Page: 9999, Size: 5, Surah: 0, Reciter: 999}
	got := bad.valid()
	if got.Theme != "auto" || got.Font != 0 || got.Page != 1 || got.Size != defaultSettings().Size || got.Surah != 1 || got.Reciter != 7 {
		t.Errorf("valid() kept %+v", got)
	}
	// A custom theme may not take a built-in's ID.
	if got := (Settings{Theme: "night", Themes: []CustomTheme{{ID: "night"}}}).valid(); len(got.Themes) != 0 || got.Theme != "night" {
		t.Errorf("valid() kept a theme named as a built-in: %+v", got.Themes)
	}
}

func TestCustomTheme(t *testing.T) {
	light := CustomTheme{ID: "c", Name: "Senja", Page: "#FFF4E6", Surface: "#FBE7CF", Text: "#3A2414", Accent: "#C2410C"}.paper()
	if light.Dark || !light.Custom || hexColor(light.Accent) != "#C2410C" {
		t.Errorf("light theme: %+v", light)
	}
	dark := CustomTheme{ID: "d", Page: "#101418", Surface: "#161B22", Text: "#E6EDF3", Accent: "#58A6FF"}.paper()
	if !dark.Dark {
		t.Error("a dark page makes no dark theme")
	}
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Onboarded = true
	a := newApp(q, set, "")
	a.saveTheme(themeDraft{Name: "Senja", Page: light.Page, Surface: light.Surface, Text: light.Text, Accent: light.Accent})
	if len(a.set.Themes) != 1 || a.set.Theme != a.set.Themes[0].ID || a.set.Themes[0].Name != "Senja" {
		t.Fatalf("after saving: theme %q, themes %+v", a.set.Theme, a.set.Themes)
	}
	// Changing it keeps one theme.
	d := draftOf(a.set.Themes[0].paper())
	d.Name = "Senja 2"
	a.saveTheme(d)
	if len(a.set.Themes) != 1 || a.set.Themes[0].Name != "Senja 2" {
		t.Errorf("after changing: %+v", a.set.Themes)
	}
	tt := ui.NewTester(a.view, 1280, 860)
	tt.Frame() // builds with the custom theme
}

var testQuran *Quran

func loadTestQuran(t *testing.T) *Quran {
	t.Helper()
	if testQuran == nil {
		q, err := loadQuran()
		if err != nil {
			t.Fatal(err)
		}
		testQuran = q
	}
	return testQuran
}

func TestShuffle(t *testing.T) {
	q := loadTestQuran(t)
	s := shuffle{rng: rand.New(rand.NewPCG(1, 2)), juz: 30, surah: 2}
	for _, c := range []struct {
		scope, length, count int
	}{{ScopeQuran, LengthOne, 1}, {ScopeJuz, LengthThree, 3}, {ScopeSurah, LengthFive, 5}} {
		s.scope, s.length = c.scope, c.length
		for range 200 {
			p := s.pick(q)
			s.current = &p
			vs := s.verses(q)
			if len(vs) == 0 || len(vs) > c.count {
				t.Fatalf("scope %d length %d: %d verses", c.scope, c.length, len(vs))
			}
			for _, v := range vs {
				if v.Surah != vs[0].Surah {
					t.Fatalf("passage %+v crosses surahs", p)
				}
				if c.scope == ScopeJuz && v.Juz != 30 || c.scope == ScopeSurah && v.Surah != 2 {
					t.Fatalf("scope %d: verse %s out of scope", c.scope, v.Key)
				}
			}
		}
	}
	s.scope, s.length = ScopeQuran, LengthPage
	p := s.pick(q)
	if p.Page == 0 || q.All[p.First].Page != p.Page || q.All[p.First+p.Count-1].Page != p.Page {
		t.Errorf("page passage %+v", p)
	}
}

func TestReadingCredit(t *testing.T) {
	q := loadTestQuran(t)
	r := newReading(newJSONFile(t.TempDir() + "/stats.json"))
	now := time.Date(2026, 11, 6, 9, 0, 0, 0, time.Local)
	vs := []*Verse{&q.Surahs[0].Verses[0], &q.Surahs[0].Verses[1], &q.Surahs[0].Verses[2]}
	// One second is not enough for the first verse; the next seconds read
	// the verses in order.
	r.credit(vs, 1, now)
	if r.day(now).Read != 0 {
		t.Fatalf("read %d after a second", r.day(now).Read)
	}
	need := readNeed(vs[0]) + readNeed(vs[1])
	r.credit(vs, need-1, now)
	if r.day(now).Read != 2 || !r.isRead(vs[0].Index) || !r.isRead(vs[1].Index) || r.isRead(vs[2].Index) {
		t.Errorf("read %d; verses read %v %v %v", r.day(now).Read, r.isRead(vs[0].Index), r.isRead(vs[1].Index), r.isRead(vs[2].Index))
	}
	// Counted verses are not counted again this session.
	r.credit(vs[:2], 100, now)
	if r.day(now).Read != 2 {
		t.Errorf("read %d after reading the same verses again", r.day(now).Read)
	}
	// The streak: today and the two days before.
	r.stats.Days[dayKey(now.AddDate(0, 0, -1))] = &DayStats{Read: 3}
	r.stats.Days[dayKey(now.AddDate(0, 0, -2))] = &DayStats{Read: 1}
	r.stats.Days[dayKey(now.AddDate(0, 0, -4))] = &DayStats{Read: 9}
	if n := r.streak(now); n != 3 {
		t.Errorf("streak %d, want 3", n)
	}
	if n := r.streak(now.AddDate(0, 0, 1)); n != 3 {
		t.Errorf("streak the next morning %d, want 3 (yesterday counts)", n)
	}
	// Saved and read back.
	r.dirty = true
	r.save()
	waitFor(t, func() bool {
		r2 := newReading(r.file)
		return r2.isRead(vs[1].Index) && r2.stats.Days[dayKey(now)] != nil && r2.stats.Days[dayKey(now)].Read == 2
	})
}

// waitFor waits for the background writes a condition depends on.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for range 200 {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out")
}

// fakeSound plays nothing, and ends when the test says.
type fakeSound struct {
	path    string
	playing bool
	end     func()
}

func (f *fakeSound) Play()  { f.playing = true }
func (f *fakeSound) Pause() { f.playing = false }
func (f *fakeSound) Close() { f.playing = false }
func (f *fakeSound) OnEnd(fn func()) func() {
	f.end = fn
	return func() {}
}

// testServer serves translations, recitation lists and their audio, as
// quran.com does.
func testServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/resources/translations", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"translations":[{"id":134,"name":"King Fahad Quran Complex","author_name":"KFQC","language_name":"indonesian"},{"id":20,"name":"Saheeh International","author_name":"Saheeh","language_name":"english"}]}`)
	})
	mux.HandleFunc("/api/quran/translations/134", func(w http.ResponseWriter, r *http.Request) {
		var b strings.Builder
		b.WriteString(`{"translations":[`)
		for i := range VerseCount {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, `{"text":"Ayat %d<sup foot_note=1>1</sup>"}`, i+1)
		}
		b.WriteString(`]}`)
		fmt.Fprint(w, b.String())
	})
	mux.HandleFunc("/api/recitations/7/by_chapter/", func(w http.ResponseWriter, r *http.Request) {
		surah := strings.TrimPrefix(r.URL.Path, "/api/recitations/7/by_chapter/")
		n := map[string]int{"1": 7, "112": 4}[surah]
		var files []string
		for v := 1; v <= n; v++ {
			files = append(files, fmt.Sprintf(`{"verse_key":"%s:%d","url":"Alafasy/mp3/%03s%03d.mp3"}`, surah, v, surah, v))
		}
		fmt.Fprintf(w, `{"audio_files":[%s],"pagination":{"next_page":null}}`, strings.Join(files, ","))
	})
	mux.HandleFunc("/audio/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ID3 fake mp3 "+r.URL.Path)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// syncApp makes an app whose background work is done when the test
// asks, on the test's goroutine.
func syncApp(t *testing.T, srv *httptest.Server) (*app, func()) {
	q := loadTestQuran(t)
	a := newApp(q, defaultSettings(), t.TempDir())
	a.apiBase, a.audioBase = srv.URL+"/api", srv.URL+"/audio/"
	queue := make(chan func(), 1000)
	a.post = func(f func()) { queue <- f }
	settle := func() {
		for range 300 {
			select {
			case f := <-queue:
				f()
				continue
			case <-time.After(20 * time.Millisecond):
			}
			if len(a.rec.jobs) == 0 && len(a.trans.fetching) == 0 && (a.rec.play == nil || !a.rec.play.loading) {
				return
			}
		}
		t.Fatal("background work did not settle")
	}
	return a, settle
}

func TestTranslationDownload(t *testing.T) {
	a, settle := syncApp(t, testServer(t))
	a.loadCatalog()
	settle()
	if len(a.trans.catalog) != 2 || a.trans.catalog[0].ID != 134 {
		t.Fatalf("catalog %+v (indonesian first)", a.trans.catalog)
	}
	a.downloadTranslation(134)
	settle()
	if a.set.TranslationID != 134 || !a.trans.installed(134) {
		t.Fatalf("translation %d, installed %v, err %q", a.set.TranslationID, a.trans.installed(134), a.trans.err)
	}
	if got := a.translationText(a.quran.Verse(2, 255)); got != fmt.Sprintf("Ayat %d", a.quran.Verse(2, 255).Index+1) {
		t.Errorf("2:255 reads %q", got)
	}
	// A new run reads it from disk.
	b := newApp(a.quran, a.set, filepath.Dir(a.trans.dir))
	if got := b.translationText(a.quran.Verse(1, 1)); got != "Ayat 1" {
		t.Errorf("1:1 reads %q after a restart", got)
	}
	a.deleteTranslation(134)
	if a.set.TranslationID != 0 || a.trans.installed(134) {
		t.Errorf("after deleting: translation %d, installed %v", a.set.TranslationID, a.trans.installed(134))
	}
}

func TestRecitation(t *testing.T) {
	a, settle := syncApp(t, testServer(t))
	var sounds []*fakeSound
	a.rec.open = func(path string) (sound, error) {
		s := &fakeSound{path: path}
		sounds = append(sounds, s)
		return s, nil
	}
	ikhlas := a.quran.Surahs[111]
	a.downloadSurah(7, ikhlas)
	settle()
	if n := a.audioCount(7, ikhlas); n != 4 {
		t.Fatalf("%d of Al-Ikhlas's 4 verses downloaded, err %q", n, a.rec.err)
	}
	// Play it through: each verse plays when the one before ends.
	a.playFrom(112, 1, 0)
	for v := 1; v <= 4; v++ {
		if a.rec.play == nil || a.rec.play.verse != v {
			t.Fatalf("playing %+v, want verse %d", a.rec.play, v)
		}
		s := sounds[len(sounds)-1]
		if !s.playing || !strings.HasSuffix(s.path, fmt.Sprintf("112%03d.mp3", v)) {
			t.Fatalf("verse %d plays %+v", v, s)
		}
		s.end()
		settle()
	}
	if a.rec.play != nil {
		t.Errorf("still playing %+v after the last verse", a.rec.play)
	}
	if n := a.reading.day(a.now()).Listened; n != 4 {
		t.Errorf("listened %d verses, want 4", n)
	}
	// Repeating a verse plays it again.
	a.set.Repeat = RepeatVerse
	a.playFrom(112, 2, 0)
	sounds[len(sounds)-1].end()
	settle()
	if a.rec.play.verse != 2 {
		t.Errorf("verse %d after a repeat", a.rec.play.verse)
	}
	a.stopPlayback()
	// A verse not downloaded is fetched as it plays.
	a.set.Repeat = RepeatOff
	a.playFrom(1, 3, 3)
	settle()
	if a.rec.play == nil || a.rec.play.loading || a.rec.play.err != "" || !a.rec.has(7, 1, 3) {
		t.Fatalf("streaming 1:3: %+v", a.rec.play)
	}
	a.deleteSurahAudio(7, ikhlas)
	if n := a.audioCount(7, ikhlas); n != 0 {
		t.Errorf("%d verses left after deleting", n)
	}
}

func TestSections(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Lang, set.Onboarded = English, true
	a := newApp(q, set, "")
	tt := ui.NewTester(a.view, 1320, 880)
	for _, s := range []struct{ nav, want string }{
		{"Shuffle", "Press Shuffle now to begin."},
		{"Statistics", "Progress by juz"},
		{"Recitations", "Download all"},
		{"Settings", "Make your own theme from four colors; the rest follow from them."},
		{"Today", "Begin with Al-Fatihah"},
	} {
		if err := tt.Click(s.nav); err != nil {
			t.Fatal(err)
		}
		if !tt.HasText(s.want) {
			t.Errorf("%s: no %q in %q", s.nav, s.want, tt.Texts())
		}
	}
	if err := tt.Click("Read"); err != nil {
		t.Fatal(err)
	}
	if a.section != SectionRead {
		t.Fatalf("section %d after Read", a.section)
	}
	tt.Key(0, ui.KeyEscape)
	if err := tt.Click("Shuffle"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Shuffle now"); err != nil {
		t.Fatal(err)
	}
	if a.shuf.current == nil {
		t.Fatal("no passage after shuffling")
	}
	if err := tt.Click("Open in Mushaf"); err != nil {
		t.Fatal(err)
	}
	v := q.All[a.shuf.current.First]
	if a.section != SectionRead || a.set.Mode != ModeMushaf || (a.set.Page != v.Page) {
		t.Errorf("opened section %d mode %d page %d, want page %d", a.section, a.set.Mode, a.set.Page, v.Page)
	}
}

// TestMushafLineWidths reports the natural widths of the Mushaf's lines,
// which QuranFont.Width follows.
func TestMushafLineWidths(t *testing.T) {
	q, err := loadQuran()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range quranFonts {
		var ws []float64
		for _, p := range q.Pages[3:] {
			for i := range p.Lines {
				if p.Lines[i].Kind == LineText {
					w, _ := ui.ShapeText(p.Lines[i].Text(), ui.Font{Family: f.Family, Size: 100}, 0).Size()
					ws = append(ws, float64(w)/100)
				}
			}
		}
		sort.Float64s(ws)
		p95 := ws[len(ws)*95/100]
		t.Logf("%s: median %.1f, 95th percentile %.1f, widest %.1f ems; page %.1f", f.Name, ws[len(ws)/2], p95, ws[len(ws)-1], f.Width)
		if math.Abs(p95-float64(f.Width)) > 1 {
			t.Errorf("%s: page width %.1f, but 95%% of lines fit in %.1f ems", f.Name, f.Width, p95)
		}
	}
}

func TestKashidas(t *testing.T) {
	words := [][]Segment{
		{{Text: "قُلْ"}},
		{{Text: "يَ"}, {Text: "ـٰ", Rule: rulesByKey["madda_normal"]}, {Text: "أَيُّهَا"}},
		{{Text: "ٱللَّهِ"}}, // the name of Allah stays whole
		{{Text: "ٱلْكَـٰفِرُونَ"}},
	}
	out, n := addKashidas(words, 5)
	if n != 5 {
		t.Fatalf("added %d kashidas, want 5", n)
	}
	count := 0
	for wi, w := range out {
		var got, was strings.Builder
		for si, s := range w {
			got.WriteString(s.Text)
			if s.Rule != words[wi][si].Rule {
				t.Errorf("word %d segment %d changed rule", wi, si)
			}
		}
		for _, s := range words[wi] {
			was.WriteString(s.Text)
		}
		count += strings.Count(got.String(), "ـ") - strings.Count(was.String(), "ـ")
		if strings.ReplaceAll(got.String(), "ـ", "") != strings.ReplaceAll(was.String(), "ـ", "") {
			t.Errorf("word %d is %q, was %q", wi, got.String(), was.String())
		}
		if wi == 2 && got.String() != was.String() {
			t.Errorf("the name of Allah got kashidas: %q", got.String())
		}
	}
	if count != 5 {
		t.Errorf("%d kashidas in the words, want 5", count)
	}
	if spots := kashidaSpots([][]rune{[]rune("لَا")}); len(spots) != 0 {
		t.Errorf("kashida spots in lam-alif: %v", spots)
	}
	if spots := kashidaSpots([][]rune{[]rune("قَالَ")}); len(spots) != 1 {
		t.Errorf("kashida spots in قَالَ: %v, want one, between qaf and alif", spots)
	}
}

func TestReaderPlayToggle(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Lang, set.Onboarded = English, true
	a := newApp(q, set, t.TempDir())
	var sounds []*fakeSound
	a.rec.open = func(path string) (sound, error) {
		s := &fakeSound{path: path}
		sounds = append(sounds, s)
		return s, nil
	}
	// A verse on disk, so it plays at once.
	if err := writeFileAtomic(a.rec.path(7, 1, 1), []byte("mp3")); err != nil {
		t.Fatal(err)
	}
	a.open(q.Surahs[0])
	tt := ui.NewTester(a.view, 1320, 880)
	if !tt.HasText("Play") {
		t.Fatalf("no Play in %q", tt.Texts())
	}
	tt.Key(0, ui.KeySpace)
	if a.rec.play == nil || len(sounds) != 1 || !sounds[0].playing {
		t.Fatalf("Space did not play: %+v", a.rec.play)
	}
	tt.Frame()
	if !tt.HasText("Pause") {
		t.Errorf("the bar's button does not offer Pause: %q", tt.Texts())
	}
	if err := tt.Click("Pause"); err != nil {
		t.Fatal(err)
	}
	if !a.rec.play.paused || sounds[0].playing {
		t.Errorf("not paused after Pause")
	}
	if !tt.HasText("Resume") {
		t.Errorf("no Resume after pausing: %q", tt.Texts())
	}
	tt.Key(0, ui.KeySpace)
	if a.rec.play.paused || !sounds[0].playing {
		t.Errorf("Space did not resume")
	}
}

func TestRecapAndBadges(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Onboarded = true
	a := newApp(q, set, "")
	friday := time.Date(2026, 11, 6, 5, 0, 0, 0, time.Local) // a Friday, at dawn
	a.now = func() time.Time { return friday }
	r := a.reading
	// Al-Kahf's first ten verses at dawn, then Al-Fatihah.
	for i := range 10 {
		r.markRead(&q.Surahs[17].Verses[i], friday)
	}
	for i := range 7 {
		r.markRead(&q.Surahs[0].Verses[i], friday)
	}
	// Two days before, to make a streak of three.
	for d := 1; d <= 2; d++ {
		r.markRead(&q.Surahs[1].Verses[d], friday.AddDate(0, 0, -d))
	}
	r.markListened(3, friday)
	rc := a.recap(PeriodWeek, friday)
	if rc.Read != 19 || rc.ActiveDays != 3 || rc.LongestStreak != 3 {
		t.Errorf("week recap: read %d, days %d, streak %d", rc.Read, rc.ActiveDays, rc.LongestStreak)
	}
	if len(rc.TopSurahs) == 0 || rc.TopSurahs[0].Surah.ID != 18 || rc.TopSurahs[0].N != 10 {
		t.Errorf("top surahs %+v", rc.TopSurahs)
	}
	if rc.Persona.Name != personas[0].Name {
		t.Errorf("persona %v, want the dawn reader", rc.Persona.Name)
	}
	if rc.Reciter == nil || rc.Reciter.ID != 3 {
		t.Errorf("reciter %+v", rc.Reciter)
	}
	if got := periodStart(PeriodWeek, friday); got.Weekday() != time.Monday || got.Day() != 2 {
		t.Errorf("the week starts %v", got)
	}
	a.checkBadges(friday)
	for _, id := range []string{"first", "streak3", "kahfi"} {
		if _, ok := r.stats.Badges[id]; !ok {
			t.Errorf("no badge %q: %v", id, r.stats.Badges)
		}
	}
	if _, ok := r.stats.Badges["streak7"]; ok {
		t.Error("a seven-day badge after three days")
	}
	n := len(a.newBadges)
	a.checkBadges(friday)
	if len(a.newBadges) != n {
		t.Errorf("badges awarded twice: %d then %d", n, len(a.newBadges))
	}
	// The cards render to images of their sizes.
	for _, card := range []ShareCard{{Kind: CardSummary, Recap: rc}, {Kind: CardVerse, Verse: q.Verse(2, 255)}, {Kind: CardBadge, Badge: badgeByID("kahfi")}} {
		for format, want := range map[int][2]int{FormatStory: {1080, 1920}, FormatPost: {1080, 1080}} {
			data, err := a.cardPNG(card, papers[2], format)
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if b := img.Bounds(); b.Dx() != want[0] || b.Dy() != want[1] {
				t.Errorf("card %d format %d is %v", card.Kind, format, b)
			}
		}
	}
	// The tabs build.
	tt := ui.NewTester(a.view, 1440, 900)
	a.section = SectionStats
	for i := range 3 {
		a.statsTab = i
		tt.Frame()
	}
	a.openShare(ShareCard{Kind: CardSummary, Recap: rc})
	tt.Frame()
	if !tt.HasText(a.tx(txtSaveImage)) {
		t.Errorf("no share dialog: %q", tt.Texts())
	}
}

func TestReaderZoom(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Lang, set.Onboarded = English, true
	a := newApp(q, set, "")
	a.open(q.Surahs[0])
	tt := ui.NewTester(a.view, 1320, 880)
	tt.Key(ui.Cmd, ui.KeyEqual)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyEqual)
	if a.set.Size != 38 || a.set.TransSize != 17 {
		t.Errorf("after zooming in twice: size %v, translation %v", a.set.Size, a.set.TransSize)
	}
	tt.Key(ui.Cmd, ui.KeyMinus)
	if a.set.Size != 36 {
		t.Errorf("after zooming out: %v", a.set.Size)
	}
	for range 20 {
		tt.Key(ui.Cmd, ui.KeyEqual)
	}
	if a.set.Size != maxSize {
		t.Errorf("zoomed past the largest size: %v", a.set.Size)
	}
	tt.Key(ui.Cmd, ui.Key0)
	if a.set.Size != defaultSettings().Size || a.set.TransSize != defaultSettings().TransSize {
		t.Errorf("⌘0 left size %v, translation %v", a.set.Size, a.set.TransSize)
	}
	// The home page does not zoom.
	tt.Key(0, ui.KeyEscape)
	tt.Key(ui.Cmd, ui.KeyEqual)
	if a.set.Size != defaultSettings().Size {
		t.Errorf("the home page zoomed the text to %v", a.set.Size)
	}
}

func TestMigrateDataDir(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, "quaranku-app Dev")
	if err := os.MkdirAll(filepath.Join(old, "audio", "7"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"settings.json", "stats.json", filepath.Join("audio", "7", "001001.mp3")} {
		if err := os.WriteFile(filepath.Join(old, f), []byte(f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dir := filepath.Join(root, "Quranku Dev")
	migrateDataDir(dir)
	for _, f := range []string{"settings.json", "stats.json", filepath.Join("audio", "7", "001001.mp3")} {
		if data, err := os.ReadFile(filepath.Join(dir, f)); err != nil || string(data) != f {
			t.Errorf("%s after moving: %q, %v", f, data, err)
		}
	}
	// A folder in use is left alone.
	os.WriteFile(filepath.Join(old, "settings.json"), []byte("old"), 0o644)
	migrateDataDir(dir)
	if data, _ := os.ReadFile(filepath.Join(dir, "settings.json")); string(data) != "settings.json" {
		t.Errorf("settings overwritten: %q", data)
	}
}

func TestZoomMenu(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Onboarded = true
	a := newApp(q, set, "")
	m := a.appMenu(func(f func()) { f() })
	for id, accel := range map[string]string{"zoom-in": "CmdOrCtrl+=", "zoom-out": "CmdOrCtrl+-", "zoom-reset": "CmdOrCtrl+0"} {
		it := m.ItemByID(id)
		if it == nil || it.Accelerator != accel || it.Click == nil {
			t.Fatalf("menu item %s: %+v", id, it)
		}
	}
	click := func(id string) { it := m.ItemByID(id); it.Click(it, nil) }
	// Away from the reader the text keeps its size.
	click("zoom-in")
	if a.set.Size != defaultSettings().Size {
		t.Errorf("zoomed on the home page: %v", a.set.Size)
	}
	a.open(q.Surahs[0])
	click("zoom-in")
	click("zoom-in")
	click("zoom-out")
	if a.set.Size != defaultSettings().Size+2 || a.toast != "" {
		t.Errorf("after zooming in twice and out once: %v, toast %q", a.set.Size, a.toast)
	}
	click("zoom-reset")
	if a.set.Size != defaultSettings().Size {
		t.Errorf("after the reset: %v", a.set.Size)
	}
}

func TestMushafZoom(t *testing.T) {
	q := loadTestQuran(t)
	set := defaultSettings()
	set.Lang, set.Onboarded, set.Mode = English, true, ModeMushaf
	a := newApp(q, set, "")
	a.open(q.Surahs[66])
	tt := ui.NewTester(a.view, 1440, 900)
	tt.Frame()
	var fit float32
	// The page at a zoom of 1 fits the window: its size is what fits.
	sizeAt := func() float32 {
		var s float32
		tt2 := ui.NewTester(func(c *ui.Context) {
			a.view(c)
			s = a.mushafSize(c)
		}, 1440, 900)
		tt2.Frame()
		tt2.Frame()
		return s
	}
	fit = sizeAt()
	tt.Key(ui.Cmd, ui.KeyEqual)
	tt.Key(ui.Cmd, ui.KeyEqual)
	if a.set.MushafZoom != 1.2 || a.set.Size != defaultSettings().Size {
		t.Fatalf("after ⌘= twice in the Mushaf: zoom %v, text size %v", a.set.MushafZoom, a.set.Size)
	}
	if got := sizeAt(); got < fit*1.19 || got > fit*1.21 {
		t.Errorf("page text %v at 1.2, %v at 1", got, fit)
	}
	tt.Key(ui.Cmd, ui.KeyMinus)
	if a.set.MushafZoom != 1.1 {
		t.Errorf("after ⌘−: %v", a.set.MushafZoom)
	}
	for range 30 {
		tt.Key(ui.Cmd, ui.KeyMinus)
	}
	if a.set.MushafZoom != minMushafZoom {
		t.Errorf("zoomed below the least: %v", a.set.MushafZoom)
	}
	tt.Key(ui.Cmd, ui.Key0)
	if a.set.MushafZoom != 1 {
		t.Errorf("⌘0 left %v", a.set.MushafZoom)
	}
}
