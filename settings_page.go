package main

import (
	"github.com/egoist/mygo/ui"
)

// Texts of the settings page.
var (
	txtSettings       = T{"Settings", "Pengaturan"}
	txtLookSection    = T{"Look", "Tampilan"}
	txtReadingSection = T{"Reading", "Membaca"}
	txtRecitation     = T{"Recitation", "Murattal"}
	txtManageAudio    = T{"Manage downloads", "Kelola unduhan"}
	txtRepeatDefault  = T{"Repeat", "Ulangi"}
	txtAbout          = T{"About", "Tentang"}
	txtRerunIntro     = T{"Show the introduction again", "Tampilkan pengenalan awal lagi"}
	txtTajweedColors  = T{"Tajweed colors", "Warna tajwid"}
	txtTajweedGuide   = T{"Tajweed guide beside the text", "Panduan tajwid di samping teks"}
	txtGoalSection    = T{"Daily goal", "Target harian"}
)

// Settings tabs.
const (
	TabLook = iota
	TabArabic
	TabTranslation
	TabRecitation
	TabGeneral
)

var (
	txtGeneral  = T{"General", "Umum"}
	txtPlayback = T{"Playback", "Pemutaran"}
)

// settingsPage holds every preference in tabs, each a column of one width,
// applied as it changes.
func (a *app) settingsPage(c *ui.Context) {
	ui.Scroll(c).Grow(1).TrackScroll(&a.settingsScroll).Children(func() {
		pageColumn(c, 760, func() {
			ui.Column(c).Gap(16).Children(func() {
				sectionTitle(c, a.tx(txtSettings))
				if ui.Tabs(c, &a.settingsTab, a.tx(txtLookSection), a.tx(txtArabicText), a.tx(txtWhichTrans), a.tx(txtRecitation), a.tx(txtGeneral)).Changed() {
					a.settingsScroll.Y = 0
				}
			})
			switch a.settingsTab {
			case TabLook:
				a.settingsLook(c)
			case TabArabic:
				a.settingsArabic(c)
			case TabTranslation:
				a.settingsSection(c, txtWhichTrans, func() {
					a.settingRow(c, "translation", &a.set.Translation, a.tx(txtShowTrans))
					a.sliderRow(c, a.tx(txtTransSize), &a.set.TransSize, 12, 24)
				})
				a.settingsSection(c, txtChooseTrans, func() {
					a.translationChooser(c)
				})
			case TabRecitation:
				a.settingsSection(c, txtReciter, func() {
					a.reciterRows(c)
				})
				a.settingsSection(c, txtPlayback, func() {
					a.fieldRow(c, a.tx(txtRepeatDefault), func() {
						ui.Segmented(c, &a.set.Repeat, a.tx(txtRepeatOff), a.tx(txtRepeatVerse), a.tx(txtRepeatSurah)).Label(a.tx(txtRepeat))
					})
					if labeledButton(c, iconDownload, a.tx(txtManageAudio), false).AlignSelf(ui.Start).Clicked() {
						a.section = SectionAudio
					}
				})
			case TabGeneral:
				a.settingsGeneral(c)
			}
		})
	})
}

func (a *app) settingsLook(c *ui.Context) {
	t := c.Theme()
	a.settingsSection(c, txtTheme, func() {
		ui.Row(c).Gap(10).Wrap().Children(func() {
			for _, p := range allPapers(a.set.Themes) {
				a.themeSwatch(c, p)
			}
			a.newThemeSwatch(c)
		})
		ui.Text(c, a.tx(txtCustomThemes)).FontSize(12).TextColor(t.TextMuted)
	})
	if a.editing != nil {
		a.themeEditor(c)
	}
}

func (a *app) settingsArabic(c *ui.Context) {
	t := c.Theme()
	a.settingsSection(c, txtArabicText, func() {
		names := make([]string, len(quranFonts))
		for i, f := range quranFonts {
			names[i] = f.Name
		}
		a.fieldRow(c, a.tx(txtFont), func() {
			ui.Segmented(c, &a.set.Font, names...).Label(a.tx(txtFont))
		})
		a.sliderRow(c, a.tx(txtTextSize), &a.set.Size, minSize, maxSize)
		a.sliderRow(c, a.tx(txtLineSpacing), &a.set.LineSpacing, 1.5, 3)
		// The text as it reads now.
		f := quranFonts[a.set.Font]
		ui.RichText(c).Font(f.Family).FontSize(float32(a.set.Size)).LineHeight(float32(a.set.LineSpacing)).
			Padding(16, 20).Radius(10).Background(t.Background).Children(func() {
			for n := 1; n <= 2; n++ {
				a.segments(c, a.quran.Verse(112, n).Segments, "settings")
				ui.Text(c, " ")
				a.ayahMark(c, n, "settings")
				ui.Text(c, " ")
			}
		})
	})
	a.settingsSection(c, txtTajweed, func() {
		a.settingRow(c, "colors", &a.set.Colors, a.tx(txtTajweedColors))
		a.settingRow(c, "guide", &a.set.Guide, a.tx(txtTajweedGuide))
	})
	a.settingsSection(c, txtMushafOpts, func() {
		a.settingRow(c, "spread", &a.set.Spread, a.tx(txtSpread))
	})
}

func (a *app) settingsGeneral(c *ui.Context) {
	t := c.Theme()
	a.settingsSection(c, txtLanguage, func() {
		a.fieldRow(c, a.tx(txtLanguage), func() {
			ui.Segmented(c, &a.set.Lang, "English", "Bahasa Indonesia").Label(a.tx(txtLanguage))
		})
	})
	a.settingsSection(c, txtGoalSection, func() {
		a.goalStepper(c)
	})
	a.settingsSection(c, txtAbout, func() {
		ui.Text(c, a.tx(txtCredits)).FontSize(12).TextColor(t.TextMuted).LineHeight(1.5)
		if ui.Button(c, a.tx(txtRerunIntro)).AlignSelf(ui.Start).Clicked() {
			a.set.Onboarded = false
			a.onboardStep = 0
		}
	})
}

// settingsSection is a titled group of settings.
func (a *app) settingsSection(c *ui.Context, title T, fn func()) {
	t := c.Theme()
	ui.Column(c.Key(title[0])).Gap(12).Children(func() {
		ui.Text(c, a.tx(title)).FontSize(16).Bold()
		ui.Column(c).Gap(12).Padding(18, 20).Radius(14).Background(t.Surface).Border(1, t.Border).Children(fn)
	})
}

// fieldRow is a label with its control.
func (a *app) fieldRow(c *ui.Context, label string, fn func()) {
	ui.Row(c.Key(label)).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(13).Width(150)
		fn()
	})
}

// goalStepper sets the daily goal, five verses a step.
func (a *app) goalStepper(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, a.tx(txtGoal)).FontSize(13).Width(150)
		if iconButton(c, iconMinus, a.tx(txtGoal)+" −").Clicked() {
			a.set.DailyGoal = max(a.set.DailyGoal-5, 5)
		}
		ui.Textf(c, "%d", int(a.set.DailyGoal)).FontSize(16).Bold().Width(44).TextAlign(ui.Center).FontFeatures("tnum")
		if iconButton(c, iconAdd, a.tx(txtGoal)+" +").Clicked() {
			a.set.DailyGoal = min(a.set.DailyGoal+5, 1000)
		}
		ui.Text(c, a.tx(txtVerses)).FontSize(13).TextColor(t.TextMuted)
	})
}
