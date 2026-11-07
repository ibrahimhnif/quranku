package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Texts of the first-run introduction.
var (
	txtWelcome       = T{"Calm daily reading, with tajweed colors and recitation.", "Tilawah harian yang tenang, dengan warna tajwid dan murattal."}
	txtSetupTime     = T{"Five quick choices, about a minute. You can change all of them later in Settings.", "Lima pilihan singkat, sekitar satu menit. Semuanya bisa diubah nanti di Pengaturan."}
	txtChooseLang    = T{"Language", "Bahasa"}
	txtBegin         = T{"Begin", "Mulai"}
	txtSkipSetup     = T{"Skip, use the defaults", "Lewati, pakai pengaturan bawaan"}
	txtStepOf        = T{"Step %d of %d", "Langkah %d dari %d"}
	txtBackStep      = T{"Back", "Kembali"}
	txtNextStep      = T{"Next", "Lanjut"}
	txtHowRead       = T{"How do you like to read?", "Bagaimana Anda suka membaca?"}
	txtHowReadNote   = T{"You can switch at any time while reading.", "Bisa diganti kapan saja saat membaca."}
	txtVerseByVerse  = T{"Verse by verse", "Ayat demi ayat"}
	txtVerseByVerseD = T{"Each verse with its translation and its own play button.", "Setiap ayat dengan terjemahannya dan tombol putar sendiri."}
	txtMushafD       = T{"The Madinah Mushaf page, 15 lines, as printed.", "Halaman Mushaf Madinah, 15 baris, seperti cetakan."}
	txtShowTrans     = T{"Show the translation", "Tampilkan terjemahan"}
	txtHowLook       = T{"Make it comfortable to read", "Atur supaya nyaman dibaca"}
	txtHowLookNote   = T{"Pick a paper, a typeface and a size. The preview is the real text.", "Pilih kertas, jenis huruf dan ukuran. Pratinjau memakai teks aslinya."}
	txtWhichReciter  = T{"Choose a reciter", "Pilih qari"}
	txtReciterNote   = T{"Press play to hear Al-Fatihah's first verse. Recitations stream as they play, and can be downloaded later.", "Tekan putar untuk mendengar ayat pertama Al-Fatihah. Murattal diputar langsung dari internet, dan bisa diunduh nanti."}
	txtDailyGoalQ    = T{"How much a day?", "Berapa ayat sehari?"}
	txtDailyGoalNote = T{"A small goal kept every day beats a large one kept now and then. A Mushaf page holds about 15 verses.", "Target kecil yang rutin lebih baik daripada target besar yang jarang. Satu halaman mushaf berisi sekitar 15 ayat."}
	txtVersesADay    = T{"%d verses", "%d ayat"}
	txtStartReading  = T{"Start reading", "Mulai membaca"}
	txtGentle        = T{"Gentle start", "Awal yang ringan"}
	txtSteady        = T{"Steady", "Rutin"}
	txtAPage         = T{"About a page", "Sekitar satu halaman"}
	txtStrong        = T{"Strong", "Bersemangat"}
)

const onboardSteps = 5

// onboarding is the first-run introduction: language, how to read, the
// look, the reciter and the daily goal, each shown as it will look.
func (a *app) onboarding(c *ui.Context) {
	t := c.Theme()
	// The welcome sits in the middle of the window; the steps after it
	// share one top, so their bar and buttons stay where they were.
	view, body := a.obView.Bounds(c), a.obBody.Bounds(c)
	top := max(40, (view.H-body.H)/2)
	if a.onboardStep > 0 {
		top = 56
	}
	ui.Scroll(c).Fill().Background(t.Background).Bind(&a.obView).Children(func() {
		ui.Column(c).Padding(top, 40, 40, 40).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Width(720).Gap(24).Bind(&a.obBody).Children(func() {
				if a.onboardStep > 0 {
					a.onboardProgress(c)
				}
				switch a.onboardStep {
				case 0:
					a.onboardWelcome(c)
				case 1:
					a.onboardReading(c)
				case 2:
					a.onboardLook(c)
				case 3:
					a.onboardReciter(c)
				case 4:
					a.onboardGoal(c)
				}
				if a.onboardStep > 0 {
					a.onboardNav(c)
				}
			})
		})
	})
	if top != a.obTop {
		a.obTop = top
		c.Invalidate()
	}
}

// finishOnboarding leaves the introduction for the home page.
func (a *app) finishOnboarding() {
	a.stopPlayback()
	a.set.Onboarded = true
	a.section = SectionHome
}

func (a *app) onboardProgress(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Row(c).Gap(6).Grow(1).Children(func() {
			for i := 1; i < onboardSteps; i++ {
				bar := ui.Box(c.Key(i)).Height(4).Grow(1).Radius(2)
				if i <= a.onboardStep {
					bar.Background(t.Accent)
				} else {
					bar.Background(t.Border)
				}
			}
		})
		ui.Textf(c, a.tx(txtStepOf), a.onboardStep, onboardSteps-1).FontSize(12).TextColor(t.TextMuted).FontFeatures("tnum")
	})
}

func (a *app) onboardNav(c *ui.Context) {
	ui.Row(c).Gap(10).AlignItems(ui.Center).Padding(8, 0, 0, 0).Children(func() {
		if labeledButton(c, iconBack, a.tx(txtBackStep), false).Clicked() {
			a.onboardStep--
		}
		ui.Spacer(c)
		if a.onboardStep == onboardSteps-1 {
			if labeledButton(c, iconRead, a.tx(txtStartReading), true).Padding(10, 22).Clicked() {
				a.finishOnboarding()
			}
			return
		}
		if ui.PrimaryButton(c, a.tx(txtNextStep)).Padding(10, 22).Clicked() {
			a.onboardStep++
		}
	})
}

// onboardHeading is a step's question and its note.
func onboardHeading(c *ui.Context, title, note string) {
	t := c.Theme()
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).FontSize(26).Bold()
		ui.Text(c, note).FontSize(14).TextColor(t.TextMuted).LineHeight(1.5)
	})
}

func (a *app) onboardWelcome(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Gap(28).Padding(40, 0, 0, 0).Children(func() {
		ui.Column(c).Gap(10).Children(func() {
			ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Quranku").FontSize(40).Bold()
				ui.Text(c, "قرآنكو").Font(quranFont).FontSize(40).TextColor(t.TextMuted)
			})
			ui.Text(c, a.tx(txtWelcome)).FontSize(18).LineHeight(1.5)
			ui.Text(c, a.tx(txtSetupTime)).FontSize(14).TextColor(t.TextMuted).LineHeight(1.5)
		})
		// The basmalah, as the app shows it.
		f := quranFonts[a.set.Font]
		ui.RichText(c).Font(f.Family).FontSize(38).LineHeight(2).TextAlign(ui.Center).AlignSelf(ui.Stretch).
			Padding(18).Radius(16).Background(t.Surface).Children(func() {
			a.segments(c, a.quran.Basmalah)
		})
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.tx(txtChooseLang)).FontSize(14).Width(90)
			ui.Segmented(c, &a.set.Lang, "English", "Bahasa Indonesia").Label(a.tx(txtChooseLang))
		})
		ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
			if ui.PrimaryButton(c, a.tx(txtBegin)).Padding(12, 32).Clicked() {
				a.onboardStep = 1
			}
			skip := ui.Text(c, a.tx(txtSkipSetup)).FontSize(13).TextColor(t.TextMuted).Cursor(ui.CursorPointer).Underline()
			if skip.Clicked() {
				a.finishOnboarding()
			}
		})
	})
}

// choiceCard is a large option to choose, with a preview inside.
func (a *app) choiceCard(c *ui.Context, key string, chosen bool, title, note string, preview func()) ui.Element {
	t := c.Theme()
	card := ui.Column(c.Key(key)).Grow(1).Basis(0).Padding(18).Gap(12).Radius(14).Cursor(ui.CursorPointer).Background(t.Surface)
	switch {
	case chosen:
		card.Border(2, t.Accent)
	case card.Hovered():
		card.Border(1, t.TextMuted)
	default:
		card.Border(1, t.Border)
	}
	card.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, title).FontSize(16).Bold().Grow(1)
			if chosen {
				ui.Icon(c, iconCheck).FontSize(18).TextColor(t.Accent)
			}
		})
		ui.Text(c, note).FontSize(13).TextColor(t.TextMuted).LineHeight(1.4)
		preview()
	})
	return card
}

func (a *app) onboardReading(c *ui.Context) {
	t := c.Theme()
	f := quranFonts[a.set.Font]
	v := a.quran.Verse(1, 2)
	onboardHeading(c, a.tx(txtHowRead), a.tx(txtHowReadNote))
	ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
		if a.choiceCard(c, "verses", a.set.Mode == ModeVerses, a.tx(txtVerseByVerse), a.tx(txtVerseByVerseD), func() {
			ui.Column(c).Gap(6).Padding(12).Radius(10).Background(t.Background).Children(func() {
				ui.RichText(c).Font(f.Family).FontSize(24).LineHeight(1.8).Children(func() {
					a.segments(c, v.Segments)
					ui.Text(c, " ")
					a.ayahMark(c, 2)
				})
				ui.Text(c, a.translationText(v)).FontSize(12).TextColor(t.TextMuted)
			})
		}).Clicked() {
			a.set.Mode = ModeVerses
		}
		if a.choiceCard(c, "mushaf", a.set.Mode == ModeMushaf, a.tx(txtMushafMode), a.tx(txtMushafD), func() {
			ui.Column(c).Gap(4).Padding(10).Radius(10).Background(a.pal.Page).Border(1.5, a.pal.Frame.Alpha(0.5)).Children(func() {
				ui.Text(c, "سُورَةُ الفاتحة").Font(quranFont).FontSize(15).TextColor(t.Accent).AlignSelf(ui.Center)
				for _, n := range []int{1, 2} {
					ui.RichText(c.Key(n)).Font(f.Family).FontSize(19).LineHeight(1.7).TextAlign(ui.Center).Children(func() {
						a.segments(c, a.quran.Verse(1, n).Segments)
						ui.Text(c, " ")
						a.ayahMark(c, n)
					})
				}
			})
		}).Clicked() {
			a.set.Mode = ModeMushaf
		}
	})
	a.settingRow(c, "ob-translation", &a.set.Translation, a.tx(txtShowTrans))
}

func (a *app) onboardLook(c *ui.Context) {
	t := c.Theme()
	onboardHeading(c, a.tx(txtHowLook), a.tx(txtHowLookNote))
	ui.Row(c).Gap(10).Wrap().Children(func() {
		for _, p := range papers {
			a.themeSwatch(c, p)
		}
	})
	names := make([]string, len(quranFonts))
	for i, f := range quranFonts {
		names[i] = f.Name
	}
	ui.Segmented(c, &a.set.Font, names...).Label(a.tx(txtFont)).AlignSelf(ui.Start)
	a.sliderRow(c, a.tx(txtTextSize), &a.set.Size, minSize, maxSize)
	// The preview: the start of Al-Fatihah in the settings chosen.
	f := quranFonts[a.set.Font]
	ui.Column(c).Padding(20, 24).Gap(14).Radius(14).Background(t.Background).Border(1, t.Border).Children(func() {
		for n := 2; n <= 3; n++ {
			v := a.quran.Verse(1, n)
			ui.Column(c.Key(n)).Gap(4).Children(func() {
				ui.RichText(c).Font(f.Family).FontSize(float32(a.set.Size)).LineHeight(float32(a.set.LineSpacing)).Children(func() {
					a.segments(c, v.Segments)
					ui.Text(c, " ")
					a.ayahMark(c, n)
				})
				if a.set.Translation {
					ui.Text(c, a.translationText(v)).FontSize(float32(a.set.TransSize)).TextColor(t.TextMuted)
				}
			})
		}
	})
}

func (a *app) onboardReciter(c *ui.Context) {
	onboardHeading(c, a.tx(txtWhichReciter), a.tx(txtReciterNote))
	a.reciterList(c)
}

// reciterList lists the reciters, each with a sample to hear and the one
// chosen marked.
func (a *app) reciterList(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Gap(2).Radius(14).Background(t.Surface).Border(1, t.Border).Padding(6).Children(func() {
		a.reciterRows(c)
	})
}

// reciterRows are the reciters' rows, for a container of the caller's.
func (a *app) reciterRows(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Gap(2).Children(func() {
		for _, r := range reciters {
			chosen := a.set.Reciter == r.ID
			row := ui.Row(c.Key(r.ID)).Padding(8, 12).Gap(12).AlignItems(ui.Center).Radius(10).Cursor(ui.CursorPointer)
			switch {
			case chosen:
				row.Background(t.Accent.Alpha(0.12))
			case row.Hovered():
				row.Background(t.SurfaceHover)
			}
			if row.Clicked() {
				a.set.Reciter = r.ID
			}
			row.Children(func() {
				pb := a.rec.play
				previewing := pb != nil && pb.preview && pb.reciter == r.ID
				switch {
				case previewing && pb.loading:
					ui.Spinner(c).Width(32)
				case previewing && !pb.paused:
					if iconButton(c, iconPause, a.tx(txtPause)).Clicked() {
						a.stopPlayback()
					}
				default:
					if iconButton(c, iconPlay, fmt.Sprintf("%s %s", a.tx(txtPlay), r.Name)).Clicked() {
						a.previewReciter(r.ID)
						a.set.Reciter = r.ID
					}
				}
				ui.Column(c).Grow(1).Gap(1).Children(func() {
					name := ui.Text(c, r.Name).FontSize(14)
					if chosen {
						name.Bold().TextColor(t.Accent)
					}
					if r.Style != "" {
						ui.Text(c, r.Style).FontSize(12).TextColor(t.TextMuted)
					}
				})
				if previewing && pb.err != "" {
					ui.Text(c, pb.err).FontSize(11).TextColor(t.Danger).MaxWidth(240).SingleLine()
				}
				if chosen {
					ui.Icon(c, iconCheck).FontSize(18).TextColor(t.Accent)
				}
			})
		}
	})
}

func (a *app) onboardGoal(c *ui.Context) {
	t := c.Theme()
	onboardHeading(c, a.tx(txtDailyGoalQ), a.tx(txtDailyGoalNote))
	ui.Row(c).Gap(12).AlignItems(ui.Stretch).Children(func() {
		for _, g := range []struct {
			n    int
			name T
		}{{5, txtGentle}, {10, txtSteady}, {15, txtAPage}, {30, txtStrong}} {
			chosen := int(a.set.DailyGoal) == g.n
			card := ui.Column(c.Key(g.n)).Grow(1).Basis(0).Padding(18).Gap(6).Radius(14).Cursor(ui.CursorPointer).Background(t.Surface)
			switch {
			case chosen:
				card.Border(2, t.Accent)
			case card.Hovered():
				card.Border(1, t.TextMuted)
			default:
				card.Border(1, t.Border)
			}
			if card.Clicked() {
				a.set.DailyGoal = float64(g.n)
			}
			card.Children(func() {
				num := ui.Textf(c, a.tx(txtVersesADay), g.n).FontSize(20).Bold()
				if chosen {
					num.TextColor(t.Accent)
				}
				ui.Text(c, a.tx(g.name)).FontSize(13).TextColor(t.TextMuted)
			})
		}
	})
}
