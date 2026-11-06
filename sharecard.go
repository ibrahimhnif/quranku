package main

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
)

// Kinds of share card.
const (
	CardSummary = iota // a period's verses, minutes, days
	CardSurahs         // the surahs read most
	CardTime           // when the reading happens
	CardHabit          // streak and reciter
	CardVerse          // one verse
	CardBadge          // a badge earned
)

// Formats of share card, in DIPs; exported at twice the size.
const (
	FormatStory = iota // 1080 x 1920, for stories
	FormatPost         // 1080 x 1080, for posts
)

const cardWidth = 540

func cardHeight(format int) float32 {
	if format == FormatPost {
		return 540
	}
	return 960
}

// ShareCard is what a card shows.
type ShareCard struct {
	Kind  int
	Recap Recap
	Verse *Verse
	Badge *Badge
}

// cardStyles are the papers a card can be printed on: the user's theme
// first.
func (a *app) cardStyles() []Paper {
	base := builtinPaper("cream")
	if p := paperByID(a.set.Theme, a.set.Themes); p != nil && !p.Auto {
		base = p
	}
	out := []Paper{drenched(*base)}
	if base.ID != "cream" {
		out = append(out, *base)
	}
	for _, id := range []string{"cream", "night", "green", "white"} {
		if id != a.set.Theme {
			out = append(out, *builtinPaper(id))
		}
	}
	return out
}

// Texts of the cards.
var (
	txtShare       = T{"Share", "Bagikan"}
	txtShareVerse  = T{"Share this verse", "Bagikan ayat ini"}
	txtInvite      = T{"Let's read the Quran together", "Yuk tilawah bareng"}
	txtCardVerses  = T{"verses read", "ayat dibaca"}
	txtCardMinutes = T{"minutes", "menit"}
	txtCardDays    = T{"days reading", "hari membaca"}
	txtCardStreak  = T{"days in a row", "hari beruntun"}
	txtTopSurahs   = T{"The surahs I read most", "Surah yang paling sering kubaca"}
	txtMostlyAt    = T{"Most of my verses were read between %02d.00 and %02d.00", "Kebanyakan ayatku dibaca pukul %02d.00–%02d.00"}
	txtLongest     = T{"My longest streak", "Runtutan terpanjangku"}
	txtFavReciter  = T{"Reciter I listened to most", "Qari yang paling sering kudengar"}
	txtVersesHeard = T{"%d verses heard", "%d ayat didengar"}
	txtBadgeEarned = T{"Earned on %s", "Diraih %s"}
	txtStyle       = T{"Style", "Gaya"}
	txtFormat      = T{"Format", "Format"}
	txtSaveImage   = T{"Save image…", "Simpan gambar…"}
	txtCopyImage   = T{"Copy image", "Salin gambar"}
	txtCopied      = T{"Image copied. Paste it in your story or chat.", "Gambar disalin. Tempel di story atau chat."}
	txtSaved       = T{"Image saved", "Gambar tersimpan"}
	txtImageSize   = T{"%d × %d pixels", "%d × %d piksel"}
	txtCardTajweed = T{"Tajweed colors", "Warna tajwid"}
	txtNewBadge    = T{"New badge: %s", "Lencana baru: %s"}
	txtRecapOf     = T{"My Quran reading · %s", "Tilawahku · %s"}
	txtClose       = T{"Close", "Tutup"}
)

// starOrnament draws the eight-pointed star of the Mushaf's ornaments in
// outline: the outline of two squares, one turned an eighth, in a circle,
// with a small circle at its heart.
func starOrnament(c *ui.Context, size, stroke float32, col ui.Color) ui.Element {
	return ui.Box(c).Size(size, size).Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		R := r.W/2 - stroke
		half := R * 0.62 / float32(math.Sqrt2) // half a square's side
		point := func(radius, deg float64) (float32, float32) {
			a := deg * math.Pi / 180
			return cx + float32(radius*math.Cos(a)), cy + float32(radius*math.Sin(a))
		}
		// Corners of the squares at every 45 degrees, and where their
		// sides cross, between them.
		star := &ui.Path{}
		outer := float64(half) * math.Sqrt2
		inner := float64(half) / math.Cos(math.Pi/8)
		for k := range 16 {
			radius := outer
			if k%2 == 1 {
				radius = inner
			}
			x, y := point(radius, float64(k)*22.5)
			if k == 0 {
				star.MoveTo(x, y)
			} else {
				star.LineTo(x, y)
			}
		}
		p.StrokePath(star.Close(), stroke, col)
		circle := func(radius float64) {
			ring := &ui.Path{}
			for k := 0; k <= 96; k++ {
				x, y := point(radius, float64(k)*360/96)
				if k == 0 {
					ring.MoveTo(x, y)
				} else {
					ring.LineTo(x, y)
				}
			}
			p.StrokePath(ring.Close(), stroke, col)
		}
		circle(float64(R))
		circle(float64(half) * 0.55)
	})
}

// drenched is a paper's accent as the whole card, its text on it: the
// boldest of the styles, as cards in a feed are.
func drenched(p Paper) Paper {
	text := ui.Hex("#FFFFFF")
	if luminance(p.Accent) > 0.6 {
		text = ui.Hex("#141414")
	}
	return Paper{
		ID: "bold", Name: T{"Bold", "Berwarna"}, Dark: luminance(p.Accent) <= 0.6,
		Page: p.Accent, Surface: p.Accent, Text: text, Muted: text.Alpha(0.78), Accent: text,
		Border: text.Alpha(0.3),
	}
}

// cardView draws a card, unit DIPs to a point: 1 to export, less for a
// preview. Every size scales with it, so the preview is the image.
func (a *app) cardView(c *ui.Context, card ShareCard, p Paper, format int, unit float32) {
	u := func(v float32) float32 { return v * unit }
	w, h := u(cardWidth), u(cardHeight(format))
	ink, muted, accent := p.Text, p.Muted, p.Accent
	ui.Column(c).Size(w, h).Background(p.Page).Padding(u(22)).Clip().Children(func() {
		// The eight-pointed star of the Mushaf's ornaments, large and
		// faint, as the cards' mark.
		starOrnament(c, u(440), u(2.5), accent.Alpha(0.16)).Absolute().Right(-u(150)).Bottom(-u(150)).PassThrough()
		ui.Column(c).Grow(1).Radius(u(16)).Border(u(1.5), accent.Alpha(0.5)).Padding(u(30), u(30), u(24), u(30)).Gap(u(18)).
			TextColor(ink).Children(func() {
			// Header: the app, and what the card is of.
			ui.Row(c).Gap(u(8)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Quranku").FontSize(u(17)).Bold().TextColor(ink)
				ui.Text(c, "قرآنكو").Font(quranFont).FontSize(u(19)).TextColor(muted)
				ui.Spacer(c)
				ui.Text(c, a.cardLabel(card)).FontSize(u(12)).TextColor(muted).SingleLine()
			})
			ui.Column(c).Grow(1).Justify(ui.Center).Gap(u(16)).Children(func() {
				a.cardBody(c, card, p, format, unit)
			})
			// Footer: an invitation, not a boast.
			ui.Box(c).Height(u(1)).Background(accent.Alpha(0.35))
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.tx(txtInvite)).FontSize(u(13)).TextColor(muted).Grow(1)
				ui.Text(c, "quranku").FontSize(u(13)).Bold().TextColor(accent)
			})
		})
	})
}

func (a *app) cardLabel(card ShareCard) string {
	switch card.Kind {
	case CardVerse:
		s := a.quran.Surahs[card.Verse.Surah-1]
		return fmt.Sprintf("QS %s [%d]: %d", s.Name, s.ID, card.Verse.Number)
	case CardBadge:
		return a.tx(txtBadges)
	}
	return a.periodName(card.Recap)
}

func (a *app) cardBody(c *ui.Context, card ShareCard, p Paper, format int, unit float32) {
	u := func(v float32) float32 { return v * unit }
	ink, muted, accent := p.Text, p.Muted, p.Accent
	post := format == FormatPost
	big := func(v float32) float32 { // big type is smaller on a square card
		if post {
			return u(v * 0.72)
		}
		return u(v)
	}
	r := card.Recap
	stat := func(n string, label string) {
		ui.Column(c.Key(label)).Grow(1).Basis(0).Gap(u(2)).Children(func() {
			ui.Text(c, n).FontSize(big(30)).Bold().TextColor(ink).FontFeatures("tnum")
			ui.Text(c, label).FontSize(u(12)).TextColor(muted)
		})
	}
	switch card.Kind {
	case CardSummary:
		ui.Text(c, fmt.Sprintf(a.tx(txtRecapOf), a.periodName(r))).FontSize(u(15)).TextColor(muted)
		ui.Text(c, a.thousands(r.Read)).FontSize(big(108)).Bold().TextColor(accent).FontFeatures("tnum").LineHeight(1)
		ui.Text(c, a.tx(txtCardVerses)).FontSize(big(26)).Bold()
		ui.Box(c).Height(u(10))
		ui.Row(c).Gap(u(12)).Children(func() {
			stat(fmt.Sprint(r.Minutes), a.tx(txtCardMinutes))
			stat(fmt.Sprint(r.ActiveDays), a.tx(txtCardDays))
			stat(fmt.Sprint(r.LongestStreak), a.tx(txtCardStreak))
		})
	case CardSurahs:
		ui.Text(c, a.tx(txtTopSurahs)).FontSize(big(30)).Bold().LineHeight(1.2)
		most := 1
		if len(r.TopSurahs) > 0 {
			most = r.TopSurahs[0].N
		}
		n := len(r.TopSurahs)
		if post {
			n = min(n, 3)
		}
		for i, sc := range r.TopSurahs[:n] {
			ui.Column(c.Key(sc.Surah.ID)).Gap(u(6)).Children(func() {
				ui.Row(c).Gap(u(10)).AlignItems(ui.Center).Children(func() {
					ui.Text(c, fmt.Sprint(i+1)).FontSize(u(15)).Bold().TextColor(accent).Width(u(18))
					ui.Text(c, sc.Surah.Name).FontSize(big(20)).Bold().Grow(1).SingleLine()
					ui.Text(c, sc.Surah.Arabic).Font(quranFont).FontSize(big(24)).TextColor(ink)
				})
				ui.Row(c).Gap(u(10)).AlignItems(ui.Center).Padding(0, 0, 0, u(28)).Children(func() {
					bar := float32(sc.N) / float32(most)
					ui.Box(c).Height(u(8)).Width(u(300)*bar + u(6)).Radius(u(4)).Background(accent.Alpha(0.35 + 0.65*bar))
					ui.Textf(c, "%d %s", sc.N, a.tx(txtVerses)).FontSize(u(12)).TextColor(muted)
				})
			})
		}
	case CardTime:
		ui.Text(c, a.tx(r.Persona.Name)).FontSize(big(52)).Bold().TextColor(accent).LineHeight(1.1)
		ui.Text(c, fmt.Sprintf(a.tx(txtMostlyAt), r.Persona.From%24, r.Persona.To%24)).FontSize(big(20)).LineHeight(1.35)
		ui.Box(c).Height(u(8))
		most := 1
		for _, n := range r.Hours {
			most = max(most, n)
		}
		ui.Row(c).Gap(u(3)).AlignItems(ui.End).Height(u(140)).Children(func() {
			for hr, n := range r.Hours {
				in := hr >= r.Persona.From && hr < r.Persona.To || hr+24 < r.Persona.To
				col := muted.Alpha(0.4)
				if in {
					col = accent
				}
				if n == 0 {
					col = muted.Alpha(0.18) // an empty hour, a hairline
				}
				ui.Box(c.Key(hr)).Grow(1).Basis(0).Height(max(u(2), u(140)*float32(n)/float32(most))).Radius(u(2)).Background(col)
			}
		})
		ui.Row(c).Children(func() {
			for i, l := range []string{"00", "06", "12", "18", "24"} {
				t := ui.Text(c.Key(i), l).FontSize(u(11)).TextColor(muted).Grow(1).Basis(0)
				if i == 4 {
					t.TextAlign(ui.End).Grow(0)
				}
			}
		})
	case CardHabit:
		ui.Text(c, a.tx(txtLongest)).FontSize(big(20)).TextColor(muted)
		ui.Row(c).Gap(u(12)).AlignItems(ui.End).Children(func() {
			ui.Text(c, fmt.Sprint(r.LongestStreak)).FontSize(big(108)).Bold().TextColor(accent).LineHeight(1)
			ui.Text(c, a.tx(txtDays)).FontSize(big(28)).Bold().Padding(0, 0, u(14), 0)
		})
		if r.Reciter != nil {
			ui.Box(c).Height(u(16))
			ui.Text(c, a.tx(txtFavReciter)).FontSize(big(20)).TextColor(muted)
			ui.Text(c, r.Reciter.label()).FontSize(big(30)).Bold().LineHeight(1.2)
			ui.Text(c, fmt.Sprintf(a.tx(txtVersesHeard), r.ReciterVerses)).FontSize(u(14)).TextColor(muted)
		}
	case CardVerse:
		v := card.Verse
		f := quranFonts[a.set.Font]
		n := len([]rune(v.Pickthall))
		size := min(max(44-float32(n)/14, 24), 42)
		if post {
			size *= 0.8
		}
		ui.Text(c, "سورة "+a.quran.Surahs[v.Surah-1].Arabic).Font(quranFont).FontSize(u(22)).TextColor(accent).AlignSelf(ui.End)
		ui.RichText(c).Font(f.Family).FontSize(u(size)).LineHeight(1.9).TextAlign(ui.Start).TextColor(ink).Children(func() {
			for _, seg := range v.Segments {
				el := ui.Text(c, seg.Text)
				if seg.Rule != nil && a.shareTajweed {
					el.TextColor(seg.Rule.Color(&ui.Theme{Dark: p.Dark}))
				}
			}
			ui.Text(c, " ")
			ui.Text(c, ayahMark(v.Number)).TextColor(accent)
		})
		trans := a.translationText(v)
		tsize := float32(16)
		if len([]rune(trans)) > 260 {
			tsize = 13
		}
		ui.Text(c, trans).FontSize(u(tsize)).LineHeight(1.5).TextColor(muted).MaxLines(map[bool]int{true: 6, false: 12}[post])
	case CardBadge:
		b := card.Badge
		ui.Column(c).Size(big(96), big(96)).Radius(big(48)).Background(accent.Alpha(0.14)).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, b.Icon).FontSize(big(48)).TextColor(accent)
		})
		ui.Text(c, a.tx(b.Name)).FontSize(big(44)).Bold().LineHeight(1.1)
		ui.Text(c, a.tx(b.Desc)).FontSize(big(20)).TextColor(muted).LineHeight(1.4)
		if day, ok := a.reading.stats.Badges[b.ID]; ok {
			if t, err := time.Parse("2006-01-02", day); err == nil {
				ui.Text(c, fmt.Sprintf(a.tx(txtBadgeEarned), a.longDate(t))).FontSize(u(13)).TextColor(muted)
			}
		}
	}
}

// openShare shows the share dialog for a card.
func (a *app) openShare(card ShareCard) {
	a.share = &card
	a.shareOpen = true
}

// shareDialog previews a card, in a style and a format, to save or copy.
func (a *app) shareDialog(c *ui.Context) {
	if a.share == nil {
		return
	}
	t := c.Theme()
	styles := a.cardStyles()
	if a.shareStyle >= len(styles) {
		a.shareStyle = 0
	}
	style := styles[a.shareStyle]
	ui.Modal(c, &a.shareOpen, func() {
		ui.Row(c).Gap(28).Children(func() {
			unit := float32(0.62)
			if a.shareFormat == FormatPost {
				unit = 0.7
			}
			ui.Column(c).Radius(10).Clip().Shadow(0, 8, 28, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
				a.cardView(c, *a.share, style, a.shareFormat, unit)
			})
			ui.Column(c).Width(250).Gap(14).Children(func() {
				ui.Text(c, a.tx(txtShare)).FontSize(18).Bold()
				ui.Text(c, a.tx(txtStyle)).FontSize(12).Bold().TextColor(t.TextMuted)
				ui.Row(c).Gap(8).Wrap().Children(func() {
					for i, p := range styles {
						dot := ui.Column(c.Key(p.ID)).Size(34, 34).Radius(17).Background(p.Page).Cursor(ui.CursorPointer).
							Justify(ui.Center).AlignItems(ui.Center).Label(a.tx(p.Name)).Tooltip(a.tx(p.Name))
						if i == a.shareStyle {
							dot.Border(3, t.Accent)
						} else {
							dot.Border(1, t.Border)
						}
						if dot.Clicked() {
							a.shareStyle = i
						}
						dot.Children(func() { ui.Box(c).Size(10, 10).Radius(5).Background(p.Accent) })
					}
				})
				ui.Text(c, a.tx(txtFormat)).FontSize(12).Bold().TextColor(t.TextMuted)
				ui.Segmented(c, &a.shareFormat, "Story 9:16", "Post 1:1").Label(a.tx(txtFormat))
				if a.share.Kind == CardVerse {
					a.settingRow(c, "share-tajweed", &a.shareTajweed, a.tx(txtCardTajweed))
				}
				ui.Textf(c, a.tx(txtImageSize), cardWidth*2, int(cardHeight(a.shareFormat))*2).FontSize(11).TextColor(t.TextMuted)
				ui.Spacer(c)
				if labeledButton(c, iconDownload, a.tx(txtSaveImage), true).Clicked() {
					a.saveCard(c, *a.share, style, a.shareFormat)
				}
				if labeledButton(c, iconCopy, a.tx(txtCopyImage), false).Clicked() {
					if data, err := a.cardPNG(*a.share, style, a.shareFormat); err == nil {
						c.Services().WriteClipboardData(transfer.New(transfer.NewItem(transfer.Bytes(transfer.PNG, data))))
						c.Toast(a.tx(txtCopied))
					} else {
						c.Toast(err.Error())
					}
				}
				if ui.Button(c, a.tx(txtClose)).Clicked() {
					a.shareOpen = false
				}
			})
		})
	})
}

// cardPNG renders a card to a PNG at twice its size, with the same
// drawing as the preview.
func (a *app) cardPNG(card ShareCard, p Paper, format int) ([]byte, error) {
	w, h := cardWidth, int(cardHeight(format))
	tt := ui.NewTester(func(c *ui.Context) { a.cardView(c, card, p, format, 1) }, w, h)
	tt.SetScale(2)
	tt.Frame()
	var buf bytes.Buffer
	if err := png.Encode(&buf, tt.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// saveCard asks where to save the card, and saves it there.
func (a *app) saveCard(c *ui.Context, card ShareCard, p Paper, format int) {
	data, err := a.cardPNG(card, p, format)
	if err != nil {
		c.Toast(err.Error())
		return
	}
	name := "quranku-" + strings.ToLower(strings.ReplaceAll(a.cardLabel(card), " ", "-"))
	name = strings.NewReplacer(":", "", "[", "", "]", "", "–", "-", "·", "").Replace(name) + ".png"
	saved, failed := a.tx(txtSaved), ""
	go func() {
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Title: a.tx(txtSaveImage), DefaultPath: name, CreateDirectories: true,
			Filters: []mygo.FileFilter{{Name: "PNG", Extensions: []string{"png"}}},
		})
		if err == nil && path != "" {
			err = os.WriteFile(path, data, 0o644)
		}
		a.post(func() {
			switch {
			case err != nil:
				failed = err.Error()
				a.toast = failed
			case path != "":
				a.toast = saved
			}
		})
	}()
}
