package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Sections of the app, chosen in the sidebar.
const (
	SectionHome = iota
	SectionRead // the reader, full window
	SectionShuffle
	SectionStats
	SectionAudio
	SectionSettings
)

// app is the state the window shows.
type app struct {
	quran *Quran
	surah *Surah
	set   Settings
	file  *jsonFile // the settings' file; nil in tests
	saved Settings  // what the file last saved
	shown Settings  // the settings the last pass built with

	section    int
	query      string
	verses     ui.ListState
	surahs     ui.ListState
	audioList  ui.ListState
	pageScroll ui.ScrollState
	focus      *Rule // the rule the guide explains; other rules fade while set

	editing    *themeDraft // the theme being made, in the settings
	transQuery string

	onboardStep    int
	indexTab       int  // the home's index: surahs or juz
	picker         bool // the reader's surah picker shows
	pickerQuery    string
	quick          bool // the reader's display menu shows
	homeScroll     ui.ScrollState
	mushafArea     ui.Handle // the Mushaf's scroll area, to fit the page in
	mushafMaxZoom  float64   // the zoom at which the page is as wide as it
	contentArea    ui.Handle // the sections' pages, to fit their columns
	indexWidth     float32
	obView, obBody ui.Handle // the introduction's window and content
	obTop          float32
	settingsScroll ui.ScrollState
	settingsTab    int
	statsTab       int
	statsScroll    ui.ScrollState
	recapPeriod    int

	// The share dialog: the card, its style and format.
	share        *ShareCard
	shareOpen    bool
	shareStyle   int
	shareFormat  int
	shareTajweed bool
	newBadges    []*Badge // earned, to announce
	toast        string   // a message from background work, to show

	shuf          shuffle
	shuffleScroll ui.ScrollState
	revealed      bool // the hidden passage shows

	reading      *reading
	confirmReset bool

	trans *translations
	rec   *recitation

	widths map[lineKey]float32    // natural widths of Mushaf lines, in ems
	lines  map[lineKey]*justified // Mushaf lines set with kashidas
	pal    palette

	// post runs a function on the UI thread (Window.Update); now is the
	// clock. Tests replace them.
	post               func(func())
	media              bool // the system's media controls show what plays
	now                func() time.Time
	apiBase, audioBase string
}

// newApp makes the app's state; dataDir keeps downloads and statistics,
// "" for none.
func newApp(q *Quran, set Settings, dataDir string) *app {
	set = set.valid()
	a := &app{
		quran: q, set: set, surah: q.Surahs[set.Surah-1],
		widths: map[lineKey]float32{}, lines: map[lineKey]*justified{},
		trans: newTranslations(dataDir), rec: newRecitation(dataDir),
		post: func(f func()) { f() }, now: time.Now,
		apiBase: quranAPI, audioBase: audioCDN,
	}
	statsFile := ""
	if dataDir != "" {
		statsFile = filepath.Join(dataDir, "stats.json")
	}
	a.reading = newReading(newJSONFile(statsFile))
	a.shuf = shuffle{juz: 30, surah: a.surah.ID, rng: rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 7))}
	a.section = SectionHome
	a.shown = set.clone()
	a.saved = set.clone()
	return a
}

// tx returns a text in the app's language.
func (a *app) tx(s T) string { return s[a.set.Lang] }

// open shows a surah, from its start.
func (a *app) open(s *Surah) {
	a.section = SectionRead
	a.shuf.surah = s.ID
	a.set.LastVerse = s.Verses[0].Index
	if a.set.Mode == ModeMushaf {
		a.goPage(s.Page)
		a.surah = s
		return
	}
	if a.surah == s {
		return
	}
	a.surah = s
	a.verses.ScrollTo(0, ui.Start)
}

// goPage shows page n of the Mushaf (and the one after, as a spread).
func (a *app) goPage(n int) {
	n = max(1, min(n, PageCount))
	if a.set.Spread && n%2 == 0 {
		n-- // a spread starts on an odd page, on the right
	}
	if n != a.set.Page {
		a.pageScroll.Y = 0
	}
	a.set.Page = n
	if v := a.quran.Pages[n].FirstVerse(); v != nil {
		a.surah = a.quran.Surahs[v.Surah-1]
	}
}

// pageStep is how far the next and previous buttons turn.
func (a *app) pageStep() int {
	if a.set.Spread {
		return 2
	}
	return 1
}

// settingsChanged follows changes the controls made to the settings:
// switching modes keeps the place being read.
func (a *app) settingsChanged() {
	old := a.shown
	switch {
	case old.Mode == ModeVerses && a.set.Mode == ModeMushaf:
		page := a.surah.Page
		if first, _ := a.verses.Visible(); first > 0 && first <= len(a.surah.Verses) {
			page = a.surah.Verses[first-1].Page
		}
		s := a.surah
		a.goPage(page)
		if page == s.Page {
			a.surah = s
		}
	case old.Mode == ModeMushaf && a.set.Mode == ModeVerses:
		row := 0
		if v := a.quran.Pages[a.set.Page].FirstVerse(); v != nil && v.Surah == a.surah.ID && v.Number > 1 {
			row = v.Number // the list's row of the verse: row 0 is the surah's header
		}
		a.verses.ScrollTo(row, ui.Start)
	}
	if a.set.Spread != old.Spread {
		a.goPage(a.set.Page)
	}
	if a.set.Font != old.Font {
		clear(a.lines) // kashidas are set for a font
	}
	a.shown = a.set.clone()
}

func (a *app) view(c *ui.Context) {
	if !a.set.equal(a.shown) {
		a.settingsChanged()
	}
	a.ensureTranslation()
	a.pal = a.applyTheme(c)

	if !a.set.Onboarded {
		a.onboarding(c)
		a.finishPass(c)
		return
	}
	if a.section == SectionRead {
		a.readerView(c)
	} else {
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			a.sidebar(c)
			ui.Column(c).Grow(1).Bind(&a.contentArea).Children(func() {
				switch a.section {
				case SectionHome:
					a.homePage(c)
				case SectionShuffle:
					a.shufflePage(c)
				case SectionStats:
					a.statsPage(c)
				case SectionAudio:
					a.recitationPage(c)
				case SectionSettings:
					a.settingsPage(c)
				}
				a.playerBar(c)
			})
			if a.set.Guide && a.section == SectionShuffle {
				a.guidePanel(c)
			}
		})
	}
	a.shareDialog(c)
	a.announce(c)
	a.finishPass(c)
}

// announce shows what background work and new badges have to say.
func (a *app) announce(c *ui.Context) {
	if a.toast != "" {
		c.Toast(a.toast)
		a.toast = ""
	}
	for _, b := range a.newBadges {
		c.ToastAction(fmt.Sprintf(a.tx(txtNewBadge), a.tx(b.Name)), a.tx(txtShare), func() {
			a.openShare(ShareCard{Kind: CardBadge, Badge: b})
		})
	}
	a.newBadges = nil
}

// finishPass ends a pass: the settings saved.
func (a *app) finishPass(c *ui.Context) {
	a.save()
}

// readerView is the reader in the whole window: a slim bar above the text,
// Esc back home.
func (a *app) readerView(c *ui.Context) {
	c.OnShortcut(0, ui.KeyEscape, func() { a.section = SectionHome })
	c.OnShortcut(0, ui.KeySpace, a.playPause)
	// Zoom the text, as reading apps do: ⌘+ and ⌘− by a step, ⌘0 back.
	zoom := func(steps int) func() {
		return func() { a.zoomText(steps) }
	}
	c.OnShortcut(ui.Cmd, ui.KeyEqual, zoom(1))
	c.OnShortcut(ui.Cmd|ui.Shift, ui.KeyEqual, zoom(1)) // ⌘+ on keyboards where + is shifted
	c.OnShortcut(ui.Cmd, ui.KeyMinus, zoom(-1))
	c.OnShortcut(ui.Cmd, ui.Key0, zoom(0))
	if a.set.Mode == ModeMushaf {
		// The pages turn as a book read right to left: Left is forward.
		c.OnShortcut(0, ui.KeyLeft, func() { a.goPage(a.set.Page + a.pageStep()) })
		c.OnShortcut(0, ui.KeyRight, func() { a.goPage(a.set.Page - a.pageStep()) })
		c.OnShortcut(0, ui.KeyPageDown, func() { a.goPage(a.set.Page + a.pageStep()) })
		c.OnShortcut(0, ui.KeyPageUp, func() { a.goPage(a.set.Page - a.pageStep()) })
	}
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Grow(1).Children(func() {
			a.readerBar(c)
			if a.set.Mode == ModeMushaf {
				a.mushaf(c)
			} else {
				a.reader(c)
			}
			a.playerBar(c)
		})
		if a.set.Guide {
			a.guidePanel(c)
		}
	})
}

// save stores the settings when they change.
func (a *app) save() {
	cur := a.set.clone()
	cur.Surah = a.surah.ID
	if a.file != nil && !cur.equal(a.saved) {
		a.saved = cur
		a.file.save(cur)
	}
}

// Texts of the sections.
var (
	txtRead = T{"Read", "Baca"}
)

// sidebar holds the sections.
func (a *app) sidebar(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(224).Padding(20, 14).Gap(4).Background(t.Surface).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Padding(0, 8, 18, 8).Children(func() {
			ui.Text(c, "Quranku").FontSize(19).Bold()
			ui.Text(c, "قرآنكو").Font(quranFont).FontSize(19).TextColor(t.TextMuted)
		})
		a.navItem(c, SectionHome, iconHome, a.tx(txtHome))
		a.navItem(c, SectionRead, iconRead, a.tx(txtRead))
		a.navItem(c, SectionShuffle, iconShuffle, a.tx(txtShuffle))
		a.navItem(c, SectionStats, iconStats, a.tx(txtStats))
		a.navItem(c, SectionAudio, iconListen, a.tx(txtRecitations))
		ui.Spacer(c)
		a.navItem(c, SectionSettings, iconSettings, a.tx(txtSettings))
	})
}

func (a *app) navItem(c *ui.Context, id int, svg *ui.SVG, label string) {
	t := c.Theme()
	row := ui.Row(c.Key(fmt.Sprint("nav", id))).Padding(8, 10).Gap(12).AlignItems(ui.Center).Radius(8).Cursor(ui.CursorPointer).Label(label)
	chosen := a.section == id
	switch {
	case chosen:
		row.Background(t.Accent.Alpha(0.13))
	case row.Hovered():
		row.Background(t.SurfaceHover)
	}
	if row.Clicked() {
		if id == SectionRead {
			a.continueReading(a.set.Mode)
			return
		}
		a.section = id
	}
	row.Children(func() {
		col := t.TextMuted
		if chosen {
			col = t.Accent
		}
		ui.Icon(c, svg).FontSize(17).TextColor(col)
		l := ui.Text(c, label).FontSize(14)
		if chosen {
			l.Bold().TextColor(t.Accent)
		}
	})
}

// Texts of the reader's bar.
var (
	txtBackHome    = T{"Home (Esc)", "Beranda (Esc)"}
	txtAllSettings = T{"All settings", "Semua pengaturan"}
	txtChooseSurah = T{"Choose a surah", "Pilih surah"}
	txtSpaceKey    = T{"Space to play or pause", "Spasi untuk putar/jeda"}
)

// readerBar shows where the reader is, with what changes the reading.
func (a *app) readerBar(c *ui.Context) {
	t := c.Theme()
	today, goal := a.reading.day(a.now()).Read, int(a.set.DailyGoal)
	ui.Row(c).Padding(8, 16).Gap(12).AlignItems(ui.Center).Children(func() {
		if iconButton(c, iconBack, a.tx(txtBackHome)).Clicked() {
			a.section = SectionHome
		}
		// The surah, which opens the picker.
		title := ui.Row(c.Key("title")).Gap(8).AlignItems(ui.Center).Padding(6, 10).Radius(8).Cursor(ui.CursorPointer).Label(a.tx(txtChooseSurah)).MinWidth(0)
		if title.Hovered() {
			title.Background(t.SurfaceHover)
		}
		if title.Clicked() {
			a.picker = !a.picker
			a.pickerQuery = ""
		}
		title.Children(func() {
			ui.Column(c).Gap(1).MinWidth(0).Children(func() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Textf(c, "%d. %s", a.surah.ID, a.surah.Name).FontSize(15).Bold().SingleLine()
					ui.Icon(c, iconChevron).FontSize(14).TextColor(t.TextMuted)
				})
				where := a.tx(a.surah.Meaning)
				if vs := a.visibleVerses(); len(vs) > 0 {
					where = fmt.Sprintf("%s · %s %d", where, a.tx(txtJuz), vs[0].Juz)
				}
				if a.set.Mode == ModeMushaf {
					p := a.quran.Pages[a.set.Page]
					where = fmt.Sprintf("%s %d · %s %d", a.tx(txtPage), p.Number, a.tx(txtJuz), p.Juz)
				}
				ui.Textf(c, "%s · %s", where, a.todayLine(today, goal)).FontSize(12).TextColor(t.TextMuted).SingleLine()
			})
		})
		ui.Popover(c, title, &a.picker, func() { a.surahPicker(c) })
		ui.Spacer(c)
		a.playToggle(c)
		ui.Segmented(c, &a.set.Mode, a.tx(txtVersesMode), a.tx(txtMushafMode)).Label(a.tx(txtMode))
		aa := iconButton(c, iconText, a.tx(txtDisplay)+" (⌘+ ⌘− ⌘0)")
		if aa.Clicked() {
			a.quick = !a.quick
		}
		ui.Popover(c, aa, &a.quick, func() { a.quickDisplay(c) })
	})
	// Today's reading, along the bar's edge.
	ui.Progress(c, min(float64(today)/float64(max(goal, 1)), 1)).Height(3).Radius(0).Label(a.tx(txtGoal))
}

// todayLine says how today's reading stands against the goal.
func (a *app) todayLine(today, goal int) string {
	if today >= goal {
		return fmt.Sprintf(a.tx(txtGoalReached), today)
	}
	return fmt.Sprintf(a.tx(txtOfToday), today, goal)
}

// surahPicker finds a surah to read, without leaving the reader.
func (a *app) surahPicker(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(340).Gap(8).Padding(4).Children(func() {
		ui.SearchField(c, &a.pickerQuery).Placeholder(a.tx(txtSearch)).AutoFocus()
		results := a.quran.Search(a.pickerQuery)
		ui.List(c, &a.surahs, len(results), func(i int) {
			s := results[i]
			row := ui.Row(c.Key(s.ID)).Padding(7, 10).Gap(10).AlignItems(ui.Center).Radius(8).Cursor(ui.CursorPointer)
			switch {
			case s == a.surah:
				row.Background(t.Accent.Alpha(0.13))
			case row.Hovered():
				row.Background(t.SurfaceHover)
			}
			if row.Clicked() {
				a.picker = false
				a.open(s)
			}
			row.Children(func() {
				ui.Text(c, fmt.Sprint(s.ID)).FontSize(12).TextColor(t.TextMuted).Width(26).FontFeatures("tnum")
				ui.Column(c).Grow(1).Children(func() {
					ui.Text(c, s.Name).FontSize(14).SingleLine()
					ui.Text(c, a.tx(s.Meaning)).FontSize(11).TextColor(t.TextMuted).SingleLine()
				})
				ui.Text(c, s.Arabic).Font(quranFont).FontSize(17)
			})
		}).Height(380)
	})
}

// quickDisplay changes how the text looks without leaving it.
func (a *app) quickDisplay(c *ui.Context) {
	ui.Column(c).Width(400).Gap(12).Padding(6).Children(func() {
		ui.Row(c).Gap(6).Wrap().Children(func() {
			for _, p := range allPapers(a.set.Themes) {
				a.themeSwatch(c, p)
			}
		})
		a.sliderRow(c, a.tx(txtTextSize), &a.set.Size, minSize, maxSize)
		a.sliderRow(c, a.tx(txtLineSpacing), &a.set.LineSpacing, 1.5, 3)
		a.settingRow(c, "q-translation", &a.set.Translation, a.tx(txtTranslation))
		a.settingRow(c, "q-colors", &a.set.Colors, a.tx(txtTajweedColors))
		a.settingRow(c, "q-guide", &a.set.Guide, a.tx(txtTajweedGuide))
		if a.set.Mode == ModeMushaf {
			a.settingRow(c, "q-spread", &a.set.Spread, a.tx(txtSpread))
		}
		if labeledButton(c, iconSettings, a.tx(txtAllSettings), false).AlignSelf(ui.Start).Clicked() {
			a.quick = false
			a.section = SectionSettings
		}
	})
}

// playToggle is the reader's play button: it plays from where the reader
// is, then pauses and resumes what plays, as the bar below does (Space too).
func (a *app) playToggle(c *ui.Context) {
	pb := a.rec.play
	if pb != nil && pb.preview {
		pb = nil
	}
	tip := reciterByID(a.set.Reciter).label() + " · " + a.tx(txtSpaceKey)
	switch {
	case pb == nil:
		if labeledButton(c, iconPlay, a.tx(txtPlaySurah), false).Tooltip(tip).Clicked() {
			a.playHere()
		}
	case pb.loading:
		labeledButton(c, iconPause, a.tx(txtLoading), false).Disabled(true)
	case pb.paused:
		if labeledButton(c, iconPlay, a.tx(txtResume), false).Tooltip(tip).Clicked() {
			a.togglePause()
		}
	default:
		if labeledButton(c, iconPause, a.tx(txtPause), true).Tooltip(tip).Clicked() {
			a.togglePause()
		}
	}
}

// Texts of the menu.
var (
	txtMenuView  = T{"View", "Tampilan"}
	txtZoomIn    = T{"Larger Text", "Perbesar Teks"}
	txtZoomOut   = T{"Smaller Text", "Perkecil Teks"}
	txtZoomReset = T{"Actual Size", "Ukuran Bawaan"}
)

// appMenu is the menu bar: the system's menus, with a View menu whose zoom
// sizes the reader's text. update runs a change on the UI thread.
func (a *app) appMenu(update func(func())) *mygo.Menu {
	zoom := func(steps int) func(*mygo.MenuItem, *mygo.Window) {
		return func(*mygo.MenuItem, *mygo.Window) {
			update(func() {
				if a.section != SectionRead {
					return // only the reader's text zooms
				}
				a.zoomText(steps)
			})
		}
	}
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Role: mygo.RoleFileMenu},
		{Role: mygo.RoleEditMenu},
		{Label: a.tx(txtMenuView), Submenu: []*mygo.MenuItem{
			{ID: "zoom-reset", Label: a.tx(txtZoomReset), Accelerator: "CmdOrCtrl+0", Click: zoom(0)},
			{ID: "zoom-in", Label: a.tx(txtZoomIn), Accelerator: "CmdOrCtrl+=", Click: zoom(1)},
			{ID: "zoom-out", Label: a.tx(txtZoomOut), Accelerator: "CmdOrCtrl+-", Click: zoom(-1)},
			mygo.Separator(),
			{Role: mygo.RoleToggleFullScreen},
		}},
		{Role: mygo.RoleWindowMenu},
	})
}

// zoomText makes the reading larger by steps, or smaller; zero sets it
// back. In the Mushaf it scales the page, by a tenth a step from the size
// that fits the window; verse by verse, it sizes the Arabic text and its
// translation.
func (a *app) zoomText(steps int) {
	d := defaultSettings()
	if a.set.Mode == ModeMushaf {
		z := d.MushafZoom
		if steps != 0 {
			z = math.Round((a.set.MushafZoom+0.1*float64(steps))*10) / 10
		}
		most := maxMushafZoom
		if a.mushafMaxZoom >= 1 {
			most = min(most, math.Floor(a.mushafMaxZoom*10)/10) // as wide as the window
		}
		a.set.MushafZoom = min(max(z, minMushafZoom), most)
		return
	}
	if steps == 0 {
		a.set.Size, a.set.TransSize = d.Size, d.TransSize
		return
	}
	a.set.Size = min(max(a.set.Size+2*float64(steps), minSize), maxSize)
	a.set.TransSize = min(max(a.set.TransSize+float64(steps), 12), 24)
}

// playPause plays from the reader's place, or pauses and resumes.
func (a *app) playPause() {
	if pb := a.rec.play; pb != nil && !pb.preview {
		a.togglePause()
		return
	}
	a.playHere()
}

// playHere plays from where the reader is: the first verse in view.
func (a *app) playHere() {
	if vs := a.visibleVerses(); len(vs) > 0 {
		v := vs[0]
		if a.set.Mode == ModeVerses && v.Number <= 2 {
			v = &a.surah.Verses[0]
		}
		a.playFrom(v.Surah, v.Number, 0)
		return
	}
	a.playFrom(a.surah.ID, 1, 0)
}

// setting is a labeled switch; clicking the label toggles it too.
func (a *app) setting(c *ui.Context, key string, on *bool, label string) {
	ui.Row(c.Key(key)).Gap(8).AlignItems(ui.Center).Children(func() {
		if ui.Text(c, label).FontSize(13).Cursor(ui.CursorPointer).Clicked() {
			*on = !*on
		}
		ui.Switch(c, on).Label(label)
	})
}

// settingRow is a setting across a row, the switch at its end.
func (a *app) settingRow(c *ui.Context, key string, on *bool, label string) {
	ui.Row(c.Key(key)).Gap(8).AlignItems(ui.Center).Children(func() {
		if ui.Text(c, label).FontSize(13).Grow(1).Cursor(ui.CursorPointer).Clicked() {
			*on = !*on
		}
		ui.Switch(c, on).Label(label)
	})
}

// migrateDataDir moves the user's data from the folder of the app's first
// name, "quaranku-app", into the folder of its name now, the first time it
// runs with it: settings, statistics, translations and recitations.
func migrateDataDir(dir string) {
	if dir == "" {
		return
	}
	old := filepath.Join(filepath.Dir(dir), strings.Replace(filepath.Base(dir), "Quranku", "quaranku-app", 1))
	if old == dir {
		return
	}
	if _, err := os.Stat(filepath.Join(dir, "settings.json")); err == nil {
		return // already in use
	}
	entries, err := os.ReadDir(old)
	if err != nil {
		return // nothing to move
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("moving user data: %v", err)
		return
	}
	for _, e := range entries {
		from, to := filepath.Join(old, e.Name()), filepath.Join(dir, e.Name())
		if _, err := os.Stat(to); err == nil {
			continue
		}
		if err := os.Rename(from, to); err != nil {
			log.Printf("moving user data: %v", err)
		}
	}
}

func main() {
	q, err := loadQuran()
	if err != nil {
		log.Fatal(err)
	}
	dataDir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Printf("user data: %v", err)
		dataDir = ""
	}
	migrateDataDir(dataDir)
	settingsFile := newJSONFile(filepath.Join(dataDir, "settings.json"))
	if dataDir == "" {
		settingsFile = nil
	}
	a := newApp(q, loadSettings(settingsFile), dataDir)
	a.file = settingsFile
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "Quranku",
			Width:     1320,
			Height:    880,
			MinWidth:  960,
			MinHeight: 600,
			StateKey:  "main",
			Content:   ui.View(a.view),
		})
		a.post = win.Update
		a.listenMedia()
		if runtime.GOOS == "darwin" {
			// macOS's View menu would take ⌘+, ⌘− and ⌘0 for a zoom native
			// UI does not do: Quranku's zooms the text instead.
			mygo.App.SetMenu(a.appMenu(win.Update))
		}
		a.reading.last = time.Now()
		// Reading time: what is on screen, each second.
		go func() {
			for now := range time.Tick(time.Second) {
				win.Update(func() { a.tick(now, win.IsFocused()) })
			}
		}()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
