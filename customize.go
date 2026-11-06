package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

// themeDraft is a theme being made or changed.
type themeDraft struct {
	ID                          string // empty for a new theme
	Name                        string
	Page, Surface, Text, Accent ui.Color
}

func draftOf(p Paper) themeDraft {
	return themeDraft{ID: p.ID, Name: p.Name[0], Page: p.Page, Surface: p.Surface, Text: p.Text, Accent: p.Accent}
}

func (d themeDraft) theme(id string) CustomTheme {
	name := strings.TrimSpace(d.Name)
	if name == "" {
		name = "Tema"
	}
	return CustomTheme{ID: id, Name: name, Page: hexColor(d.Page), Surface: hexColor(d.Surface), Text: hexColor(d.Text), Accent: hexColor(d.Accent)}
}

// Texts of the display settings.
var (
	txtArabicText   = T{"Arabic text", "Teks Arab"}
	txtLineSpacing  = T{"Line spacing", "Jarak baris"}
	txtTransSize    = T{"Translation size", "Ukuran terjemahan"}
	txtWhichTrans   = T{"Translation", "Terjemahan"}
	txtFollowLang   = T{"Follow the app's language", "Ikuti bahasa aplikasi"}
	txtMoreTrans    = T{"More translations", "Terjemahan lainnya"}
	txtChooseTrans  = T{"Choose a translation", "Pilih terjemahan"}
	txtLoadingList  = T{"Loading the list…", "Memuat daftar…"}
	txtSearchTrans  = T{"Search translations", "Cari terjemahan"}
	txtNewTheme     = T{"New theme", "Buat tema"}
	txtEditTheme    = T{"Edit theme", "Ubah tema"}
	txtThemeName    = T{"Name", "Nama"}
	txtPaper        = T{"Page", "Kertas"}
	txtPanel        = T{"Panels", "Panel"}
	txtTextColor    = T{"Text", "Teks"}
	txtAccent       = T{"Accent", "Aksen"}
	txtSave         = T{"Save", "Simpan"}
	txtPreview      = T{"Preview", "Pratinjau"}
	txtMushafOpts   = T{"Mushaf", "Mushaf"}
	txtDone         = T{"Done", "Selesai"}
	txtDownloading  = T{"Downloading…", "Mengunduh…"}
	txtUse          = T{"Use", "Pakai"}
	txtInUse        = T{"In use", "Dipakai"}
	txtCustomThemes = T{"Make your own theme from four colors; the rest follow from them.", "Buat tema sendiri dari empat warna; warna lainnya mengikuti."}
)

// newThemeSwatch starts a theme of the user's own, from the one in use.
func (a *app) newThemeSwatch(c *ui.Context) {
	t := c.Theme()
	add := ui.Column(c.Key("new")).Gap(4).AlignItems(ui.Center).Cursor(ui.CursorPointer).Width(64).Label(a.tx(txtNewTheme))
	if add.Clicked() {
		base := paperByID(a.set.Theme, a.set.Themes)
		if base == nil || base.Auto {
			base = &papers[2]
		}
		d := draftOf(*base)
		d.ID, d.Name = "", ""
		a.editing = &d
	}
	add.Children(func() {
		ui.Column(c).Size(36, 36).Radius(18).Border(1.5, t.Border).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, iconAdd).FontSize(18).TextColor(t.TextMuted)
		})
		ui.Text(c, a.tx(txtNewTheme)).FontSize(11)
	})
}

func (a *app) sliderRow(c *ui.Context, label string, v *float64, lo, hi float64) {
	t := c.Theme()
	value := fmt.Sprintf("%.0f", *v) // a size
	if hi <= 3 {
		value = fmt.Sprintf("%.1f×", *v) // a spacing
	}
	ui.Row(c.Key(label)).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(13).Width(130)
		ui.Slider(c, v, lo, hi).Grow(1).Label(label)
		ui.Text(c, value).FontSize(12).TextColor(t.TextMuted).Width(36).TextAlign(ui.End).FontFeatures("tnum")
	})
}

// themeSwatch is a theme's button: a circle of its paper and its name. A
// custom theme's opens its editor on a second click.
func (a *app) themeSwatch(c *ui.Context, p Paper) {
	t := c.Theme()
	chosen := a.set.Theme == p.ID
	col := ui.Column(c.Key(p.ID)).Gap(4).AlignItems(ui.Center).Cursor(ui.CursorPointer).Width(64)
	if col.Clicked() {
		if chosen && p.Custom {
			d := draftOf(p)
			a.editing = &d
		}
		a.set.Theme = p.ID
	}
	if p.Custom {
		col.Tooltip(a.tx(txtEditTheme))
	}
	col.Children(func() {
		dot := ui.Row(c).Size(36, 36).Radius(18).Clip()
		if p.Auto {
			// Half the light paper, half the night.
			dot.Children(func() {
				ui.Box(c).Size(18, 36).Background(papers[2].Page)
				ui.Box(c).Size(18, 36).Background(papers[4].Page)
			})
		} else {
			dot.Background(p.Page)
		}
		if chosen {
			dot.Border(3, t.Accent)
		} else {
			dot.Border(1, t.Border)
		}
		if !p.Auto {
			dot.Children(func() {
				ui.Box(c).Size(10, 10).Radius(5).Background(p.Accent).Margin(13)
			})
		}
		name := ui.Text(c, a.tx(p.Name)).FontSize(11).SingleLine()
		if chosen {
			name.Bold().TextColor(t.Accent)
		}
	})
}

// themeEditor makes or changes a custom theme, with a preview.
func (a *app) themeEditor(c *ui.Context) {
	d := a.editing
	t := c.Theme()
	preview := CustomTheme{Page: hexColor(d.Page), Surface: hexColor(d.Surface), Text: hexColor(d.Text), Accent: hexColor(d.Accent)}.paper()
	title := txtNewTheme
	if d.ID != "" {
		title = txtEditTheme
	}
	ui.Column(c).Gap(12).Padding(16).Radius(12).Background(t.Background).Border(1, t.Border).Children(func() {
		ui.Text(c, a.tx(title)).FontSize(16).Bold()
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.tx(txtThemeName)).FontSize(13).Width(80)
			ui.TextInput(c, &d.Name).Placeholder("Tema saya").Grow(1).AutoFocus()
		})
		ui.Row(c).Gap(16).Children(func() {
			for _, w := range []struct {
				label T
				color *ui.Color
			}{{txtPaper, &d.Page}, {txtPanel, &d.Surface}, {txtTextColor, &d.Text}, {txtAccent, &d.Accent}} {
				ui.Column(c.Key(w.label[0])).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.ColorWell(c, w.color).Label(a.tx(w.label))
					ui.Text(c, a.tx(w.label)).FontSize(12)
				})
			}
		})
		// Preview.
		ui.Text(c, a.tx(txtPreview)).FontSize(12).Bold().TextColor(t.TextMuted)
		ui.Row(c).Radius(10).Border(1, preview.Border).Clip().Height(170).Children(func() {
			ui.Column(c).Width(120).Background(preview.Surface).Padding(12).Gap(6).Children(func() {
				ui.Text(c, "Quranku").Bold().TextColor(preview.Text)
				ui.Text(c, "Al-Fatihah").FontSize(12).TextColor(preview.Accent).Bold()
				ui.Text(c, "Al-Baqarah").FontSize(12).TextColor(preview.Muted)
			})
			ui.Column(c).Grow(1).Background(preview.Page).Padding(14).Gap(8).Children(func() {
				ui.RichText(c).Font(quranFonts[a.set.Font].Family).FontSize(26).TextColor(preview.Text).Children(func() {
					ui.Text(c, "بِسْمِ ")
					ui.Text(c, "ٱ").TextColor(rulesByKey["ham_wasl"].Color(&ui.Theme{Dark: preview.Dark}))
					ui.Text(c, "للَّهِ ")
					ui.Text(c, "ٱلرَّحْمَـٰنِ").TextColor(preview.Text)
				})
				ui.Text(c, "Dengan nama Allah Yang Maha Pengasih").FontSize(12).TextColor(preview.Muted)
				ui.Text(c, "1:1").FontSize(11).Bold().TextColor(preview.Accent).Background(preview.Accent.Alpha(0.12)).Radius(99).Padding(2, 8).AlignSelf(ui.Start)
			})
		})
		ui.Row(c).Gap(8).Children(func() {
			if d.ID != "" && ui.Button(c, a.tx(txtDelete)).Clicked() {
				a.set.Themes = slices.DeleteFunc(slices.Clone(a.set.Themes), func(x CustomTheme) bool { return x.ID == d.ID })
				if a.set.Theme == d.ID {
					a.set.Theme = "auto"
				}
				a.editing = nil
				return
			}
			ui.Spacer(c)
			if ui.Button(c, a.tx(txtCancel)).Clicked() {
				a.editing = nil
			}
			if ui.PrimaryButton(c, a.tx(txtSave)).Clicked() {
				a.saveTheme(*d)
				a.editing = nil
			}
		})
	})
}

// saveTheme adds a new theme, or changes one, and uses it.
func (a *app) saveTheme(d themeDraft) {
	themes := slices.Clone(a.set.Themes)
	if d.ID == "" {
		d.ID = fmt.Sprintf("custom-%d", a.now().UnixNano())
		themes = append(themes, d.theme(d.ID))
	} else {
		for i := range themes {
			if themes[i].ID == d.ID {
				themes[i] = d.theme(d.ID)
			}
		}
	}
	a.set.Themes = themes
	a.set.Theme = d.ID
}

// translationChooser lists the translations: those the app has, and those
// quran.com offers to download.
func (a *app) translationChooser(c *ui.Context) {
	t := c.Theme()
	tr := a.trans
	row := func(key string, label, detail string, chosen bool, action func()) {
		r := ui.Row(c.Key(key)).Padding(6, 10).Gap(10).AlignItems(ui.Center).Radius(8)
		if chosen {
			r.Background(t.Accent.Alpha(0.12))
		}
		r.Children(func() {
			ui.Column(c).Grow(1).Children(func() {
				name := ui.Text(c, label).FontSize(13)
				if chosen {
					name.Bold().TextColor(t.Accent)
				}
				if detail != "" {
					ui.Text(c, detail).FontSize(11).TextColor(t.TextMuted)
				}
			})
			action()
		})
	}
	use := func(id int) func() {
		return func() {
			if a.set.TranslationID == id {
				ui.Text(c, a.tx(txtInUse)).FontSize(11).TextColor(t.Accent)
				return
			}
			if ui.Button(c, a.tx(txtUse)).Clicked() {
				a.set.TranslationID = id
				a.set.Translation = true
			}
		}
	}
	row("auto", a.tx(txtFollowLang), "Kemenag RI / Pickthall", a.set.TranslationID == 0, use(0))
	for _, b := range builtinTranslations {
		row(fmt.Sprint(b.ID), b.Name, strings.Title(b.Lang), a.set.TranslationID == b.ID, use(b.ID))
	}
	ui.Text(c, a.tx(txtMoreTrans)).FontSize(12).Bold().TextColor(t.TextMuted).Padding(6, 0, 0, 0)
	a.loadCatalog()
	if tr.err != "" {
		ui.Text(c, tr.err).FontSize(11).TextColor(t.Danger)
	}
	if tr.catalog == nil {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Spinner(c)
			ui.Text(c, a.tx(txtLoadingList)).FontSize(12).TextColor(t.TextMuted)
		})
		return
	}
	ui.SearchField(c, &a.transQuery).Placeholder(a.tx(txtSearchTrans))
	q := strings.ToLower(strings.TrimSpace(a.transQuery))
	shown := 0
	for _, info := range tr.catalog {
		if info.ID == pickthallID {
			continue
		}
		hay := strings.ToLower(info.Name + " " + info.Author + " " + info.Lang)
		if q != "" && !strings.Contains(hay, q) {
			continue
		}
		if q == "" && shown >= 12 && !tr.installed(info.ID) {
			continue // the first ones, until a search
		}
		shown++
		row(fmt.Sprint(info.ID), info.Name, strings.Title(info.Lang)+" · "+info.Author, a.set.TranslationID == info.ID, func() {
			switch {
			case tr.fetching[info.ID]:
				ui.Spinner(c)
				ui.Text(c, a.tx(txtDownloading)).FontSize(11).TextColor(t.TextMuted)
			case tr.installed(info.ID):
				use(info.ID)()
				if iconButton(c, iconTrash, a.tx(txtDelete)).Clicked() {
					a.deleteTranslation(info.ID)
				}
			default:
				if labeledButton(c, iconDownload, a.tx(txtDownload), false).Disabled(tr.dir == "").Clicked() {
					a.downloadTranslation(info.ID)
				}
			}
		})
	}
}
