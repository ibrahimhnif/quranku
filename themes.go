package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Paper is a reading theme: the color of the page and what is around it.
type Paper struct {
	ID   string
	Name T
	Auto bool // follows the system's light or dark appearance
	Dark bool
	// Page is the paper; Desk is behind the Mushaf's pages; Surface the
	// sidebars.
	Page, Desk, Surface, Hover, Border, Text, Muted, Accent ui.Color
	Custom                                                  bool
}

var papers = []Paper{
	{ID: "auto", Name: T{"Auto", "Otomatis"}, Auto: true},
	{
		ID: "white", Name: T{"White", "Putih"},
		Page: ui.Hex("#FFFFFF"), Desk: ui.Hex("#EEF0F0"), Surface: ui.Hex("#F6F7F7"), Hover: ui.Hex("#EAECEC"),
		Border: ui.Hex("#E1E4E4"), Text: ui.Hex("#1B1D1E"), Muted: ui.Hex("#6B7275"), Accent: ui.Hex("#0F766E"),
	},
	{
		ID: "cream", Name: T{"Cream", "Krem"},
		Page: ui.Hex("#FBF5E6"), Desk: ui.Hex("#EDE3CB"), Surface: ui.Hex("#F4ECD8"), Hover: ui.Hex("#EADFC4"),
		Border: ui.Hex("#E0D3B3"), Text: ui.Hex("#2B2418"), Muted: ui.Hex("#7A6A50"), Accent: ui.Hex("#8A5A14"),
	},
	{
		ID: "green", Name: T{"Green", "Hijau"},
		Page: ui.Hex("#F1F7EF"), Desk: ui.Hex("#DCE9D8"), Surface: ui.Hex("#E7F0E4"), Hover: ui.Hex("#D8E6D3"),
		Border: ui.Hex("#CBDCC5"), Text: ui.Hex("#1A2A1E"), Muted: ui.Hex("#5C705F"), Accent: ui.Hex("#1F6F43"),
	},
	{
		ID: "night", Name: T{"Night", "Malam"}, Dark: true,
		Page: ui.Hex("#16191D"), Desk: ui.Hex("#0E1013"), Surface: ui.Hex("#1C2026"), Hover: ui.Hex("#262B32"),
		Border: ui.Hex("#2C323A"), Text: ui.Hex("#E8E6E3"), Muted: ui.Hex("#9AA0A6"), Accent: ui.Hex("#2DD4BF"),
	},
}

func builtinPaper(id string) *Paper {
	for i := range papers {
		if papers[i].ID == id {
			return &papers[i]
		}
	}
	return nil
}

// paper makes a custom theme's paper: its four colors, and the others
// mixed from them.
func (t CustomTheme) paper() Paper {
	page, surface, text, accent := ui.Hex(t.Page), ui.Hex(t.Surface), ui.Hex(t.Text), ui.Hex(t.Accent)
	return Paper{
		ID: t.ID, Name: T{t.Name, t.Name}, Custom: true, Dark: luminance(page) < 0.5,
		Page: page, Surface: surface, Text: text, Accent: accent,
		Desk:   surface.Mix(text, 0.08),
		Hover:  surface.Mix(text, 0.07),
		Border: surface.Mix(text, 0.14),
		Muted:  text.Mix(page, 0.42),
	}
}

// allPapers returns the built-in themes, then the user's.
func allPapers(custom []CustomTheme) []Paper {
	out := append([]Paper(nil), papers...)
	for _, t := range custom {
		out = append(out, t.paper())
	}
	return out
}

func paperByID(id string, custom []CustomTheme) *Paper {
	if p := builtinPaper(id); p != nil {
		return p
	}
	for _, t := range custom {
		if t.ID == id {
			p := t.paper()
			return &p
		}
	}
	return nil
}

// luminance is a color's relative luminance, from 0 (black) to 1 (white).
func luminance(c ui.Color) float64 {
	return (0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)) / 255
}

func hexColor(c ui.Color) string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }

// palette is the current theme's colors for the Mushaf's pages.
type palette struct {
	Page, Desk, Frame ui.Color
}

// applyTheme sets the window's theme from the chosen paper, and returns
// the page colors.
func (a *app) applyTheme(c *ui.Context) palette {
	p := paperByID(a.set.Theme, a.set.Themes)
	if p == nil {
		p = &papers[0]
	}
	var t ui.Theme
	if p.Auto {
		t = *c.Theme()
		accent := ui.Hex("#0F766E")
		if t.Dark {
			accent = ui.Hex("#2DD4BF")
		}
		t.Accent, t.AccentHover, t.AccentPressed = accent, accent, accent
		if t.Dark {
			t.AccentText = ui.Hex("#062925")
		}
	} else {
		if p.Dark {
			t = *ui.DarkTheme()
		} else {
			t = *ui.LightTheme()
		}
		t.Background, t.Surface, t.SurfaceHover, t.SurfacePressed = p.Page, p.Surface, p.Hover, p.Hover
		t.Border, t.Text, t.TextMuted = p.Border, p.Text, p.Muted
		t.Accent, t.AccentHover, t.AccentPressed = p.Accent, p.Accent, p.Accent
		// Text on the accent: dark on a light accent, white on a dark one.
		t.AccentText = ui.Hex("#FFFFFF")
		if luminance(p.Accent) > 0.6 {
			t.AccentText = ui.Hex("#111111")
		}
	}
	t.Radius = 8
	c.SetTheme(&t)
	pal := palette{Page: t.Background, Desk: t.Surface, Frame: t.Accent}
	if !p.Auto {
		pal.Desk = p.Desk
	}
	return pal
}
