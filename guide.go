package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// ruleCounts counts the rules where the reader is: the surah, or the
// page(s) of the Mushaf.
func (a *app) ruleCounts() (map[*Rule]int, T) {
	if a.set.Mode != ModeMushaf {
		return a.surah.RuleCounts, txtInSurah
	}
	counts := map[*Rule]int{}
	for _, p := range a.shownPages() {
		for r, n := range p.RuleCounts {
			counts[r] += n
		}
	}
	return counts, txtOnPage
}

// guidePanel lists the rules, and explains the one chosen.
func (a *app) guidePanel(c *ui.Context) {
	t := c.Theme()
	counts, countText := a.ruleCounts()
	ui.Scroll(c).Width(300).Background(t.Surface).BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		ui.Column(c).Padding(16).Gap(14).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.tx(txtGuideTitle)).FontSize(16).Bold().Grow(1)
				if a.focus != nil && ui.Button(c, a.tx(txtShowAll)).Clicked() {
					a.focus = nil
				}
			})
			if a.focus != nil {
				a.ruleDetail(c, a.focus, counts[a.focus], countText)
			} else {
				ui.Text(c, a.tx(txtHint)).FontSize(12).TextColor(t.TextMuted)
			}
			for _, g := range ruleGroups {
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, a.tx(*g)).FontSize(11).Bold().TextColor(t.TextMuted).Padding(4, 0)
					for _, r := range rules {
						if r.Group == g {
							a.ruleRow(c, r, counts[r])
						}
					}
				})
			}
			ui.Text(c, a.tx(txtCredits)).FontSize(10).TextColor(t.TextMuted)
		})
	})
}

func (a *app) ruleRow(c *ui.Context, r *Rule, n int) {
	t := c.Theme()
	row := ui.Row(c.Key(r.Key)).Padding(6, 8).Gap(10).AlignItems(ui.Center).Radius(6).Cursor(ui.CursorPointer)
	switch {
	case a.focus == r:
		row.Background(r.Color(t).Alpha(0.16))
	case row.Hovered():
		row.Background(t.SurfaceHover)
	}
	if row.Clicked() {
		if a.focus == r {
			a.focus = nil
		} else {
			a.focus = r
			a.set.Colors = true
		}
	}
	if n == 0 && a.focus != r {
		row.Opacity(0.45)
	}
	row.Children(func() {
		ui.Box(c).Size(12, 12).Radius(6).Background(r.Color(t))
		ui.Text(c, a.tx(r.Name)).FontSize(13).Grow(1)
		ui.Text(c, fmt.Sprint(n)).FontSize(11).TextColor(t.TextMuted).FontFeatures("tnum")
	})
}

func (a *app) ruleDetail(c *ui.Context, r *Rule, n int, countText T) {
	t := c.Theme()
	col := r.Color(t)
	ui.Column(c).Padding(14).Gap(8).Radius(10).Background(t.Background).Border(1, col.Alpha(0.5)).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(12, 12).Radius(6).Background(col)
			ui.Text(c, a.tx(r.Name)).Bold().FontSize(15).Grow(1)
		})
		ui.Text(c, r.Arabic).Font(quranFont).FontSize(26).TextColor(col)
		ui.Text(c, a.tx(r.Summary)).FontSize(13).Bold()
		ui.Text(c, a.tx(r.Detail)).FontSize(13).LineHeight(1.5)
		ui.Textf(c, a.tx(countText), n).FontSize(11).TextColor(t.TextMuted)
	})
}
