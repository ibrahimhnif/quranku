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

// onTouch reports whether the app runs on a phone or a tablet, which a
// finger drives. On a desktop, QURANKU_PHONE=1 or QURANKU_TABLET=1 shows
// their layouts, to work on them.
func onTouch() bool {
	return runtime.GOOS == "ios" || runtime.GOOS == "android" || os.Getenv("QURANKU_PHONE") == "1" || os.Getenv("QURANKU_TABLET") == "1"
}

// Below compactWidth a touch screen is a phone's: a section at a time over
// a bar of tabs. Wider, a tablet's: the sections beside a rail, or below
// railWidth the sidebar, as on a desktop.
const (
	compactWidth = 700
	railWidth    = 1000
)

// fitScreen lays a touch screen out for its width, as it turns or splits.
func (a *app) fitScreen(c *ui.Context) {
	if !a.touch {
		return
	}
	w, _ := c.Size()
	a.phone = w < compactWidth
	a.rail = !a.phone && w < railWidth
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
	a.guideSheet(c)
}

// guideSheet explains the rule of a colored letter tapped anywhere, in a
// sheet, where the screen has no room for the guide beside the text.
func (a *app) guideSheet(c *ui.Context) {
	mobile.Sheet(c, &a.guideOpen, mobile.SheetOptions{Detents: []float32{0.6, 0.92}}, func() {
		ui.Scroll(c).Grow(1).Children(func() { a.guideContent(c) })
	})
}

// sideGuide reports whether the tajweed guide can show beside the text:
// not on a phone, nor on a narrow tablet.
func (a *app) sideGuide() bool { return !a.phone && !a.rail }

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
	mobile.Sheet(c, &a.picker, half, func() {
		ui.Column(c).Grow(1).Padding(4, 12, 0, 12).Children(func() { a.surahPicker(c) })
	})
	mobile.Sheet(c, &a.quick, mobile.SheetOptions{Detents: []float32{0.42, 0.7}}, func() {
		ui.Scroll(c).Grow(1).Padding(4, 14, 16, 14).Children(func() { a.quickDisplay(c) })
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

// swipePages reports whether a finger turns the Mushaf's pages: on a touch
// screen, a page at a time.
func (a *app) swipePages() bool { return a.phone || a.touch && !a.spread() }

// phoneShare previews a card in a sheet, to share through the system's
// share sheet or save to the phone.
func (a *app) phoneShare(c *ui.Context, styles []Paper, style Paper) {
	// The card as large as the sheet holds over its controls, and the sheet
	// as tall as they are.
	w, h := c.Size()
	_, _, bottom, _ := c.SafeArea()
	const controls = 290 // the grabber, style, format, tajweed and buttons
	unit := max(min((h*0.92-controls-bottom)/cardHeight(a.shareFormat), (w-40)/cardWidth), 0.3)
	detent := min((unit*cardHeight(a.shareFormat)+controls+bottom)/h, 0.92)
	mobile.Sheet(c, &a.shareOpen, mobile.SheetOptions{Detents: []float32{detent}, Label: a.tx(txtShare)}, func() {
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(8, 20, 16, 20).Gap(12).Children(func() {
				ui.Column(c).AlignSelf(ui.Center).Radius(10).Clip().Shadow(0, 8, 28, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
					a.cardView(c, *a.share, style, a.shareFormat, unit)
				})
				a.shareOptions(c, styles)
				ui.Row(c).Gap(10).Children(func() {
					if labeledButton(c, iconShare, a.tx(txtShare), true).Grow(1).Basis(0).Clicked() {
						a.shareCardFile(c, *a.share, style, a.shareFormat, false)
					}
					if labeledButton(c, iconDownload, a.tx(txtSave), false).Grow(1).Basis(0).Clicked() {
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
	// Cards of a size the screen shows whole, with the next one's edge,
	// and their button.
	_, h := c.Size()
	w := max(min(a.contentWidth(c, 600)*0.62, (h-340)*cardWidth/cardHeight(FormatStory)), 180)
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

// navRail holds the sections on a narrow tablet: an icon over its name,
// down the side, as Material's navigation rail.
func (a *app) navRail(c *ui.Context) {
	t := c.Theme()
	// Its color goes on under the system's bars.
	top, _, bottom, _ := c.SafeArea()
	ui.Column(c).Width(84).Padding(16+top, 6, 16+bottom, 6).Margin(-top, 0, -bottom, 0).Gap(6).AlignItems(ui.Stretch).Background(t.Surface).
		BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Role(ui.RoleTabList).Children(func() {
		ui.Text(c, "قرآنكو").Font(quranFont).FontSize(20).TextColor(t.Accent).AlignSelf(ui.Center).Padding(0, 0, 10, 0)
		a.railItem(c, SectionHome, iconHome, a.tx(txtHome))
		a.railItem(c, SectionRead, iconRead, a.tx(txtRead))
		a.railItem(c, SectionShuffle, iconShuffle, a.tx(txtShuffle))
		a.railItem(c, SectionStats, iconStats, a.tx(txtStats))
		a.railItem(c, SectionAudio, iconListen, a.tx(txtRecitations))
		ui.Spacer(c)
		a.railItem(c, SectionSettings, iconSettings, a.tx(txtSettings))
	})
}

func (a *app) railItem(c *ui.Context, id int, svg *ui.SVG, label string) {
	t := c.Theme()
	chosen := a.section == id
	item := ui.Column(c.Key(fmt.Sprint("rail", id))).Gap(4).Padding(4, 0, 6, 0).AlignItems(ui.Center).
		Role(ui.RoleTab).Label(label).Checked(chosen)
	if item.Clicked() {
		if id == SectionRead {
			a.continueReading(a.set.Mode)
		} else {
			a.section = id
		}
	}
	item.Children(func() {
		col := t.TextMuted
		pill := ui.Row(c).Width(56).Height(32).Radius(16).Justify(ui.Center).AlignItems(ui.Center)
		if chosen {
			col = t.Accent
			pill.Background(t.Accent.Alpha(0.14))
		}
		pill.Children(func() { ui.Icon(c, svg).FontSize(20).TextColor(col) })
		l := ui.Text(c, label).FontSize(11).TextColor(col).NoWrap()
		if chosen {
			l.Bold()
		}
	})
}

// phoneBadges shows the badges two a row, each its icon, name and what
// earns it.
func (a *app) phoneBadges(c *ui.Context, earned map[string]string) {
	t := c.Theme()
	ui.Grid(c).Columns(2).Gap(10).Children(func() {
		for _, b := range badges {
			_, ok := earned[b.ID]
			tile := ui.Column(c.Key(b.ID)).Padding(14, 12).Gap(6).Radius(14).Background(t.Surface).Border(1, t.Border).AlignItems(ui.Center)
			if !ok {
				tile.Opacity(0.55)
			}
			if ok && tile.Clicked() {
				a.openShare(ShareCard{Kind: CardBadge, Badge: b})
			}
			tile.Children(func() {
				circle := ui.Column(c).Size(40, 40).Radius(20).Justify(ui.Center).AlignItems(ui.Center)
				col := t.TextMuted
				if ok {
					circle.Background(t.Accent.Alpha(0.14))
					col = t.Accent
				} else {
					circle.Background(t.Background)
				}
				circle.Children(func() { ui.Icon(c, b.Icon).FontSize(20).TextColor(col) })
				ui.Text(c, a.tx(b.Name)).FontSize(13).Bold().TextAlign(ui.Center)
				ui.Text(c, a.tx(b.Desc)).FontSize(11).TextColor(t.TextMuted).TextAlign(ui.Center).LineHeight(1.35)
			})
		}
	})
}
