package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/mobile"
)

// onPhone reports whether the app runs on a phone or a tablet. On a
// desktop, QURANKU_PHONE=1 shows the phone's layout, to work on it.
func onPhone() bool {
	return runtime.GOOS == "ios" || runtime.GOOS == "android" || os.Getenv("QURANKU_PHONE") == "1"
}

// phoneTabs are the sections of the bar of tabs, in order. Shuffle opens
// from the home page.
var phoneTabs = []int{SectionHome, SectionRead, SectionAudio, SectionStats, SectionSettings}

// phoneTab is the tab of a section.
func phoneTab(section int) int {
	for i, s := range phoneTabs {
		if s == section {
			return i
		}
	}
	return 0 // shuffle, under home
}

// phoneView is the app on a phone: a section at a time, over the bar of
// tabs, with what plays between them.
func (a *app) phoneView(c *ui.Context) {
	t := c.Theme()
	a.sectionOpened = a.section != a.lastSection
	a.lastSection = a.section
	ui.Column(c).Fill().Background(t.Background).Children(func() {
		ui.Column(c).Grow(1).MinHeight(0).Bind(&a.contentArea).Children(func() {
			switch a.section {
			case SectionHome:
				a.homePage(c)
			case SectionRead:
				a.phoneReader(c)
			case SectionShuffle:
				a.shufflePage(c)
			case SectionStats:
				a.statsPage(c)
			case SectionAudio:
				a.recitationPage(c)
			case SectionSettings:
				a.settingsPage(c)
			}
		})
		a.playerBar(c)
		tab := phoneTab(a.section)
		icon := func(svg *ui.SVG) func(p *ui.Painter, r ui.Rect, col ui.Color) {
			return func(p *ui.Painter, r ui.Rect, col ui.Color) { p.Icon(svg, r, col) }
		}
		bar := mobile.BottomBar(c, &tab,
			mobile.BarButton{Label: a.tx(txtHome), Icon: icon(iconHome)},
			mobile.BarButton{Label: a.tx(txtRead), Icon: icon(iconRead)},
			mobile.BarButton{Label: a.tx(txtRecitations), Icon: icon(iconListen)},
			mobile.BarButton{Label: a.tx(txtStats), Icon: icon(iconStats)},
			mobile.BarButton{Label: a.tx(txtSettings), Icon: icon(iconSettings)},
		)
		if bar.Changed() {
			if phoneTabs[tab] == SectionRead {
				a.continueReading(a.set.Mode)
			} else {
				a.section = phoneTabs[tab]
			}
		}
	})
	// A colored letter tapped anywhere explains its rule.
	mobile.Sheet(c, &a.guideOpen, mobile.SheetOptions{Detents: []float32{0.6, 0.92}}, func() {
		ui.Scroll(c).Grow(1).Children(func() { a.guideContent(c) })
	})
}

// phoneReader is the reader on a phone: its bar above the text, and the
// surah picker, the display and the guide to a rule in sheets.
func (a *app) phoneReader(c *ui.Context) {
	a.phoneReaderBar(c)
	if a.set.Mode == ModeMushaf {
		a.mushaf(c)
	} else {
		a.reader(c)
	}
	half := mobile.SheetOptions{Detents: []float32{0.6, 0.92}}
	mobile.Sheet(c, &a.picker, half, func() { a.surahPicker(c) })
	mobile.Sheet(c, &a.quick, half, func() {
		ui.Scroll(c).Grow(1).Children(func() { a.quickDisplay(c) })
	})
}

// phoneReaderBar shows the surah, which opens the picker, and what changes
// the reading.
func (a *app) phoneReaderBar(c *ui.Context) {
	t := c.Theme()
	today, goal := a.reading.day(a.now()).Read, int(a.set.DailyGoal)
	ui.Row(c).Padding(6, 8, 6, 12).Gap(4).AlignItems(ui.Center).Children(func() {
		title := ui.Column(c.Key("title")).Grow(1).MinWidth(0).Gap(1).Padding(4).Radius(8).Label(a.tx(txtChooseSurah))
		if title.Clicked() {
			a.picker = true
			a.pickerQuery = ""
		}
		title.Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Textf(c, "%d. %s", a.surah.ID, a.surah.Name).FontSize(16).Bold().SingleLine()
				ui.Icon(c, iconChevron).FontSize(14).TextColor(t.TextMuted)
			})
			where := a.tx(a.surah.Meaning)
			if a.set.Mode == ModeMushaf {
				p := a.quran.Pages[a.set.Page]
				where = fmt.Sprintf("%s %d · %s %d", a.tx(txtPage), p.Number, a.tx(txtJuz), p.Juz)
			}
			ui.Textf(c, "%s · %s", where, a.todayLine(today, goal)).FontSize(11).TextColor(t.TextMuted).SingleLine()
		})
		a.phonePlayButton(c)
		// The other way to read.
		other, label := ModeMushaf, txtMushafMode
		if a.set.Mode == ModeMushaf {
			other, label = ModeVerses, txtVersesMode
		}
		if iconButton(c, iconBook, a.tx(label)).Clicked() {
			a.set.Mode = other
		}
		if iconButton(c, iconText, a.tx(txtDisplay)).Clicked() {
			a.quick = true
		}
	})
	ui.Progress(c, min(float64(today)/float64(max(goal, 1)), 1)).Height(3).Radius(0).Label(a.tx(txtGoal))
}

// phonePlayButton plays the surah from the verse in view, or pauses it.
func (a *app) phonePlayButton(c *ui.Context) {
	pb := a.rec.play
	if pb != nil && pb.preview {
		pb = nil
	}
	switch {
	case pb == nil:
		if iconButton(c, iconPlay, a.tx(txtPlaySurah)).Clicked() {
			a.playHere()
		}
	case pb.loading:
		iconButton(c, iconPause, a.tx(txtLoading)).Disabled(true)
	case pb.paused:
		if iconButton(c, iconPlay, a.tx(txtResume)).Clicked() {
			a.togglePause()
		}
	default:
		if iconButton(c, iconPause, a.tx(txtPause)).Clicked() {
			a.togglePause()
		}
	}
}

// phonePageNav turns the Mushaf's pages on a phone, the next on the left.
func (a *app) phonePageNav(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Padding(4, 12).Gap(8).AlignItems(ui.Center).Children(func() {
		if iconButton(c, iconBack, a.tx(txtNextPage)).Disabled(a.set.Page >= PageCount).Clicked() {
			a.goPage(a.set.Page + 1)
		}
		ui.Textf(c, "%s %d / %d", a.tx(txtPage), a.set.Page, PageCount).FontSize(13).TextColor(t.TextMuted).
			FontFeatures("tnum").Grow(1).TextAlign(ui.Center)
		if iconButton(c, iconForward, a.tx(txtPrevPage)).Disabled(a.set.Page <= 1).Clicked() {
			a.goPage(a.set.Page - 1)
		}
	})
}

// spread reports whether the Mushaf shows two pages side by side: never on
// a phone, whose screen holds one.
func (a *app) spread() bool { return a.set.Spread && !a.phone }

// phoneShare previews a card in a sheet, to share through the system's
// share sheet or save to the phone.
func (a *app) phoneShare(c *ui.Context, styles []Paper, style Paper) {
	mobile.Sheet(c, &a.shareOpen, mobile.SheetOptions{Detents: []float32{0.92}, Label: a.tx(txtShare)}, func() {
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(8, 20, 24, 20).Gap(14).Children(func() {
				unit := float32(0.4)
				if a.shareFormat == FormatPost {
					unit = 0.6
				}
				ui.Column(c).AlignSelf(ui.Center).Radius(10).Clip().Shadow(0, 8, 28, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
					a.cardView(c, *a.share, style, a.shareFormat, unit)
				})
				a.shareOptions(c, styles)
				ui.Row(c).Gap(10).Children(func() {
					if labeledButton(c, iconShare, a.tx(txtShare), true).Grow(1).Basis(0).Clicked() {
						a.shareCardFile(c, *a.share, style, a.shareFormat, false)
					}
					if labeledButton(c, iconDownload, a.tx(txtSaveImage), false).Grow(1).Basis(0).Clicked() {
						a.shareCardFile(c, *a.share, style, a.shareFormat, true)
					}
				})
			})
		})
	})
}

// shareCardFile writes the card to a file of the app's, then shares it
// through the system's share sheet, or has the user save a copy of it.
func (a *app) shareCardFile(c *ui.Context, card ShareCard, p Paper, format int, save bool) {
	data, err := a.cardPNG(card, p, format)
	if err != nil {
		c.Toast(err.Error())
		return
	}
	name := a.cardFile(card)
	path := filepath.Join(os.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		c.Toast(err.Error())
		return
	}
	saved := a.tx(txtSaved)
	go func() {
		var err error
		var where string
		if save {
			where, err = mygo.Dialog.SaveCopy(path, name)
		} else {
			err = mygo.Share(mygo.ShareOptions{Files: []string{path}})
		}
		a.post(func() {
			switch {
			case err != nil:
				a.toast = err.Error()
			case where != "":
				a.toast = saved
			}
		})
	}()
}

// phoneMushaf shows the Mushaf a page at a time, which a finger turns as a
// book read right to left: the next page comes from the left. A page longer
// than the screen scrolls down.
func (a *app) phoneMushaf(c *ui.Context) {
	// The pager's pages run left to right, from the last page to the first.
	index := func(page int) int { return PageCount - page }
	if a.pager.Page != index(a.set.Page) {
		// The buttons, the picker or the recitation turned it: a new pager
		// opens on the page, rather than sliding past those between.
		a.pager = mobile.PageState{Page: index(a.set.Page)}
		a.pagerGen++
	}
	ui.Column(c).Grow(1).Background(a.pal.Desk).Children(func() {
		a.pageNav(c)
		at := a.pager.Page
		ui.Column(c.Key(fmt.Sprint("pager", a.pagerGen))).Grow(1).Children(func() {
			a.phonePages(c, at)
		})
		if a.pager.Page != at {
			a.goPage(PageCount - a.pager.Page) // a finger turned it
		}
	})
}

// phonePages are the Mushaf's pages a finger swipes between.
func (a *app) phonePages(c *ui.Context, at int) {
	{
		mobile.PageView(c, &a.pager, PageCount, func(i int) {
			p := a.quran.Pages[PageCount-i]
			scroll := ui.Scroll(c.Key(p.Number)).Fill()
			if i == at {
				scroll.TrackScroll(&a.pageScroll).Bind(&a.mushafArea)
			}
			scroll.Children(func() {
				ui.Row(c).Justify(ui.Center).AlignItems(ui.Start).Padding(8, 8, 24, 8).Children(func() {
					a.mushafPage(c, p)
				})
			})
		}).Grow(1).FillWidth()
	}
}

// tabRow chooses *selected of the tabs, and reports a new choice. On a
// phone, a row wider than the screen scrolls across from edge to edge, and
// keeps the chosen tab in view as it is chosen or the section opens.
func (a *app) tabRow(c *ui.Context, selected *int, labels ...string) bool {
	if !a.phone {
		return ui.Tabs(c, selected, labels...).Changed()
	}
	t := c.Theme()
	changed := false
	ui.ScrollHorizontal(c).Margin(0, -16, 0, -16).Padding(0, 12).Gap(4).Role(ui.RoleTabList).Children(func() {
		for i, label := range labels {
			on := i == *selected
			tab := ui.Column(c.Key(i)).Padding(8, 10, 0, 10).Gap(8).AlignItems(ui.Center).
				Role(ui.RoleTab).Label(label).Checked(on)
			if tab.Clicked() && !on {
				*selected, changed, on = i, true, true
			}
			tab.Children(func() {
				text := ui.Text(c, label).FontSize(15).NoWrap()
				line := ui.Box(c).Height(2).FillWidth().Radius(1)
				if on {
					text.Bold()
					line.Background(t.Accent)
				} else {
					text.TextColor(t.TextMuted)
				}
			})
			if on && (changed || a.sectionOpened) {
				tab.ScrollIntoView(ui.ScrollIntoViewOptions{Inline: ui.ScrollNearest, Block: ui.ScrollNearest})
			}
		}
	})
	return changed
}

// recapCarousel shows the recap's cards side by side, a finger swiping from
// one to the next, the next peeking in at the edge.
func (a *app) recapCarousel(c *ui.Context, cards []ShareCard, style Paper) {
	// Cards of a size the screen shows whole, with the next one's edge.
	w := max(a.contentWidth(c, 600)*0.62, 180)
	unit := w / cardWidth
	ui.ScrollHorizontal(c).ScrollSnap(ui.SnapMandatory).Margin(0, -16, 0, -16).Padding(0, 16, 4, 16).Gap(12).Children(func() {
		for i, card := range cards {
			ui.Column(c.Key(i)).Gap(10).AlignItems(ui.Center).SnapAlign(ui.Start).Children(func() {
				shot := ui.Column(c).Radius(8).Clip().Shadow(0, 4, 16, 0, ui.RGBA(0, 0, 0, 0.12)).Label(a.tx(txtShare))
				if shot.Clicked() {
					a.openShare(card)
				}
				shot.Children(func() {
					a.cardView(c, card, style, FormatStory, unit)
				})
				if labeledButton(c, iconShare, a.tx(txtShare), true).Clicked() {
					a.openShare(card)
				}
			})
		}
	})
}

// lineHeight is the Arabic text's line height, as the settings say it. A
// phone's screen holds few words a line, so the lines sit closer: 2.0 in
// the settings is 1.7 there.
func (a *app) lineHeight() float32 {
	h := float32(a.set.LineSpacing)
	if a.phone {
		h = 1 + (h-1)*0.7
	}
	return h
}
