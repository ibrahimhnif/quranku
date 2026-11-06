---
name: Quranku
description: A calm desktop Quran reader on paper themes, where the text is the only ornament.
colors:
  white-page: "#FFFFFF"
  white-desk: "#EEF0F0"
  white-surface: "#F6F7F7"
  white-hover: "#EAECEC"
  white-border: "#E1E4E4"
  white-text: "#1B1D1E"
  white-muted: "#6B7275"
  white-accent: "#0F766E"
  cream-page: "#FBF5E6"
  cream-desk: "#EDE3CB"
  cream-surface: "#F4ECD8"
  cream-hover: "#EADFC4"
  cream-border: "#E0D3B3"
  cream-text: "#2B2418"
  cream-muted: "#7A6A50"
  cream-accent: "#8A5A14"
  green-page: "#F1F7EF"
  green-desk: "#DCE9D8"
  green-surface: "#E7F0E4"
  green-hover: "#D8E6D3"
  green-border: "#CBDCC5"
  green-text: "#1A2A1E"
  green-muted: "#5C705F"
  green-accent: "#1F6F43"
  night-page: "#16191D"
  night-desk: "#0E1013"
  night-surface: "#1C2026"
  night-hover: "#262B32"
  night-border: "#2C323A"
  night-text: "#E8E6E3"
  night-muted: "#9AA0A6"
  night-accent: "#2DD4BF"
  auto-dark-accent-text: "#062925"
  accent-text-light: "#FFFFFF"
  accent-text-dark: "#111111"
  tajweed-madd-natural: "#537FFF"
  tajweed-madd-natural-dark: "#7FA0FF"
  tajweed-madd-permissible: "#4050FF"
  tajweed-madd-permissible-dark: "#8A96FF"
  tajweed-madd-obligatory: "#2144C1"
  tajweed-madd-obligatory-dark: "#6C8CFF"
  tajweed-madd-necessary: "#000EBC"
  tajweed-madd-necessary-dark: "#5A6BFF"
  tajweed-ghunnah: "#FF7E1E"
  tajweed-ghunnah-dark: "#FF9A4D"
  tajweed-ikhfa: "#9400A8"
  tajweed-ikhfa-dark: "#D45AE6"
  tajweed-idgham-ghunnah: "#169777"
  tajweed-idgham-ghunnah-dark: "#3CC9A2"
  tajweed-idgham-no-ghunnah: "#169200"
  tajweed-idgham-no-ghunnah-dark: "#4FCB35"
  tajweed-iqlab: "#26BFFD"
  tajweed-iqlab-dark: "#5CD0FF"
  tajweed-ikhfa-shafawi: "#D500B7"
  tajweed-ikhfa-shafawi-dark: "#F05CDA"
  tajweed-idgham-shafawi: "#58B800"
  tajweed-idgham-shafawi-dark: "#7FD936"
  tajweed-qalqalah: "#DD0008"
  tajweed-qalqalah-dark: "#FF5A60"
  tajweed-idgham-mutajanisayn: "#A1A1A1"
  tajweed-idgham-mutajanisayn-dark: "#8E8E8E"
  tajweed-silent: "#AAAAAA"
  tajweed-silent-dark: "#7A7A7A"
typography:
  wordmark:
    fontFamily: "system-ui"
    fontSize: "40px"
    fontWeight: 700
  headline:
    fontFamily: "system-ui"
    fontSize: "26px"
    fontWeight: 700
  section-title:
    fontFamily: "system-ui"
    fontSize: "24px"
    fontWeight: 700
  title:
    fontFamily: "system-ui"
    fontSize: "22px"
    fontWeight: 700
  title-small:
    fontFamily: "system-ui"
    fontSize: "16px"
    fontWeight: 700
  body:
    fontFamily: "system-ui"
    fontSize: "14px"
    fontWeight: 400
  body-small:
    fontFamily: "system-ui"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
  label:
    fontFamily: "system-ui"
    fontSize: "12px"
    fontWeight: 400
    fontFeature: "\"tnum\""
  translation:
    fontFamily: "system-ui"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.6
  quran-reading:
    fontFamily: "Amiri Quran, Scheherazade New"
    fontSize: "34px"
    fontWeight: 400
    lineHeight: 2
  quran-continue:
    fontFamily: "Amiri Quran, Scheherazade New"
    fontSize: "30px"
    fontWeight: 400
    lineHeight: 1.9
  quran-surah-title:
    fontFamily: "Amiri Quran"
    fontSize: "40px"
    fontWeight: 400
  quran-surah-name:
    fontFamily: "Amiri Quran"
    fontSize: "19px"
    fontWeight: 400
rounded:
  frame: "4px"
  page: "6px"
  control: "8px"
  tile: "10px"
  card: "12px"
  panel: "14px"
  sample: "16px"
  hero: "18px"
  pill: "999px"
spacing:
  hair: "4px"
  xs: "8px"
  sm: "12px"
  md: "14px"
  lg: "20px"
  xl: "28px"
  page-top: "36px"
  page-side: "40px"
  page-bottom: "48px"
components:
  button-primary:
    backgroundColor: "{colors.white-accent}"
    textColor: "{colors.accent-text-light}"
    rounded: "{rounded.control}"
    padding: "12px 20px"
  button-secondary:
    backgroundColor: "{colors.white-surface}"
    textColor: "{colors.white-text}"
    rounded: "{rounded.control}"
  nav-item:
    textColor: "{colors.white-muted}"
    rounded: "{rounded.control}"
    padding: "8px 10px"
  nav-item-hover:
    backgroundColor: "{colors.white-hover}"
  sidebar:
    backgroundColor: "{colors.white-surface}"
    width: "224px"
    padding: "20px 14px"
  continue-panel:
    backgroundColor: "{colors.white-surface}"
    rounded: "{rounded.hero}"
    padding: "28px 28px 24px 28px"
  surah-tile:
    rounded: "{rounded.tile}"
    padding: "10px 12px"
  surah-tile-hover:
    backgroundColor: "{colors.white-hover}"
  surah-number-current:
    backgroundColor: "{colors.white-accent}"
    textColor: "{colors.accent-text-light}"
    rounded: "{rounded.pill}"
    size: "30px"
  settings-section:
    backgroundColor: "{colors.white-surface}"
    rounded: "{rounded.panel}"
    padding: "18px 20px"
  stat-card:
    backgroundColor: "{colors.white-surface}"
    rounded: "{rounded.card}"
    padding: "16px"
  choice-card:
    backgroundColor: "{colors.white-surface}"
    rounded: "{rounded.panel}"
    padding: "18px"
  verse-card:
    padding: "20px 16px"
  verse-card-playing:
    rounded: "{rounded.panel}"
    padding: "20px 16px"
  reciting-pill:
    backgroundColor: "{colors.white-accent}"
    textColor: "{colors.accent-text-light}"
    rounded: "{rounded.pill}"
    padding: "3px 10px"
  rule-card:
    backgroundColor: "{colors.white-page}"
    textColor: "{colors.white-text}"
    rounded: "{rounded.tile}"
    padding: "12px 14px"
    width: "280px"
  player-bar:
    backgroundColor: "{colors.white-surface}"
    padding: "10px 20px"
  reader-progress-edge:
    backgroundColor: "{colors.white-accent}"
    height: "3px"
  mushaf-page:
    backgroundColor: "{colors.white-page}"
    rounded: "{rounded.page}"
  theme-swatch:
    rounded: "{rounded.pill}"
    size: "36px"
---

# Design System: Quranku

## Overview

**Creative North Star: "The Quiet Paper"**

Quranku is a sheet of paper on a desk. Each theme is a paper (White, Cream, Green, Night, or one the reader mixes), with one ink for text, one muted ink for everything secondary, and one accent. Nothing is drawn on the paper for its own sake: the Quran text, its ayah marks and its surah titles are the only ornament, and the chrome around them is flat, thin-lined and quiet. The app opens on today's reading (the verse where the reader stopped, in Arabic, at reading size) and reading itself is a full-window view entered with Continue and left with Esc.

Density is relaxed. Pages are centered columns with generous outer margins (36 top, 40 sides, 48 bottom) and a 28px rhythm between blocks; lists are airy rows that only show a fill on hover. Depth comes from stepping between the paper's tones (Page, Surface, Desk) and from 1px hairlines, not from shadows. Tajweed help, translation and recitation sit in the same quiet register as the chrome so they never compete with the text.

This is a native desktop UI drawn by MyGo in Go. There is no CSS: every value below lives in Go code (`themes.go`, `tajweed.go`, and the views). The theme radius is 8.

**Key Characteristics:**
- Paper themes with a fixed role set: Page, Desk, Surface, Hover, Border, Text, Muted, Accent.
- One accent per paper, spent on the primary action, progress, the current selection and the text's own ornaments.
- Emphasis by fading what surrounds, never by drawing boxes.
- One platform sans for the UI; Amiri Quran (or Scheherazade New) for Arabic.
- Lucide line icons at 1.75 stroke.
- Flat surfaces, hairline borders, no gradients.

## Colors

Every paper is a quiet tonal set of one hue family plus a single deeper accent; the tajweed palette is the only multi-hue color in the app, and it belongs to the text.

### Primary
- **Mosque Teal** (`white-accent`): the accent of White, and of Auto in light appearance. The Continue button, progress bars, the selected nav item and list row, the current surah's number disc, ayah marks and surah titles.
- **Lamp Teal** (`night-accent`): the accent of Night, and of Auto in dark appearance (with `auto-dark-accent-text` on it).
- **Old Ink Brown** (`cream-accent`): Cream's accent, a deep amber-brown like aged ink.
- **Leaf Green** (`green-accent`): Green's accent.

### Neutral
Each paper defines the same eight roles. Read them as one family per theme:
- **Page** (`*-page`): the reading surface and the window background; also the Mushaf page itself.
- **Desk** (`*-desk`): what lies behind the Mushaf's pages, one step darker than the paper.
- **Surface** (`*-surface`): sidebars, the player bar, cards and section panels.
- **Hover** (`*-hover`): the fill a row or link gains under the pointer; also the pressed state.
- **Border** (`*-border`): every hairline: card edges, verse dividers, the sidebar's edge.
- **Text** (`*-text`): body and headings.
- **Muted** (`*-muted`): dates, metadata, secondary lines, inactive icons, translation text.

### Tajweed
The 17 tajweed rules use quran.com's palette (`tajweed-*`), each with a `-dark` variant lightened to stay readable on Night. Rules within a family share a hue (the four madd rules are a blue ramp from light to near-navy; silent letters and the two mutajanis/mutaqarib idghams are greys). These colors are applied to the letters themselves and to small 10-12px dots in the guide and rule card; never to backgrounds larger than a guide row's 0.16 wash.

### Named Rules
**The One Accent Rule.** Each paper has one accent. It marks the primary action (Continue, Begin, Next, the player's play button), progress (goal bars, the reader's edge line, juz progress, the 14-day chart), and current selection (nav item, picked surah, picked reciter, picked theme, selected choice card). The only other place it appears is the text's own ornament: ayah marks, surah titles in Arabic, the Mushaf frame and surah header lozenge. Selection is shown as an accent wash at 0.12-0.14 alpha with accent bold label text, never a solid accent block.

**The Accent Ink Rule.** Text on a solid accent is white (`accent-text-light`) unless the accent's luminance exceeds 0.6, then near-black (`accent-text-dark`). Auto's dark appearance uses `auto-dark-accent-text`.

**The Four Colors Rule.** A custom theme is four chosen colors (Page, Surface, Text, Accent); the rest are mixed, never chosen: Desk is Surface mixed 8% toward Text, Hover 7%, Border 14%, Muted is Text mixed 42% toward Page. A theme is dark when its Page luminance is under 0.5.

## Typography

**UI Font:** the platform sans (no family set; MyGo's default system face).
**Quran Font:** Amiri Quran (default) or Scheherazade New, the reader's choice; Amiri Quran always for the app's own Arabic (surah names, the wordmark's Arabic, page headers).

**Character:** an unobtrusive system sans that steps aside for a calligraphic naskh. Hierarchy in the UI comes from size and bold only; there is no second Latin face.

### Hierarchy
- **Wordmark** (bold, 40px): "Quranku" with "قرآنكو" in Muted Amiri beside it, on the onboarding welcome only; 19px in the sidebar.
- **Headline** (bold, 26px): the home greeting and each onboarding step's title; the date or step note sits under it in Muted 14px.
- **Section title** (bold, 24px): the top of Stats, Settings, Shuffle, Recitations.
- **Title** (bold, 22px): the continue panel's surah and verse reference. 16-18px bold for section headings, card titles and the index heading.
- **Body** (14px): list rows, nav labels, strip items. 13px for setting labels, notes and summaries (line height 1.4-1.5 for prose).
- **Label** (12px, Muted, tabular figures): metadata lines, counts, verse keys, page numbers. 11px for the smallest notes.
- **Translation** (15px default, 11-26 adjustable, line height 1.6, Muted): under each verse.
- **Quran reading** (34px default, 24-52 adjustable, line height 2 default, 1.5-3 adjustable): verse text in the reader. The Mushaf sizes from 0.8 of this and shrinks to fit the window.

### Named Rules
**The One Sans Rule.** The UI uses the platform sans alone. Do not introduce a display face; weight and size carry the hierarchy.

**The Core Text Rule.** Arabic is set only in Amiri Quran or Scheherazade New. Faces whose marks Core Text drops (KFGQPC, Digital Khatt) are not used.

**The Tabular Figures Rule.** Numbers that line up or tick (verse keys, surah numbers, page counts, slider values, stats) use `tnum`.

## Layout

Two window structures. **Sections** (Home, Shuffle, Stats, Recitations, Settings) put a 224px Surface sidebar on the left (brand, nav items, Settings pinned to the bottom) and the page to its right, with the player bar along the bottom when audio plays. **The reader** takes the full window: the reader bar on top, the text, the player bar, and the tajweed guide as a 300px right panel when on.

**The Centered Column Rule.** Section content sits in a centered column with a maximum width: Home, Stats and wide Settings 1280; Settings 780 when the content area is under 1100; Shuffle 960; onboarding 720; verse-by-verse reading 880 (with 32px side padding). The column has 36 / 40 / 48 outer padding and a 28px gap between blocks.

Responsive behavior follows the measured content width, not the window: the home index takes one column per 320px, clamped to two through four; at 1100px or more, Settings splits into two equal columns (28 gap) and Stats puts its chart and juz grid side by side. The window opens at 1320x880 and stops at 960x600.

Spacing rhythm: 4 inside tight stacks, 8-14 within cards and rows, 20-28 between blocks. Card padding is 16-20; the continue panel uses 28.

The Mushaf view fits its page to the window: the reader's size is the most it takes, and the page shrinks so all 15 lines are in view, as a printed page is. Two pages form a spread right to left with a 20px gap.

## Elevation & Depth

Flat by default. Depth is tonal (Desk under Page, Surface for panels) and drawn with 1px Border hairlines. Shadows are rare, soft and ambient, used where a physical sheet or a floating card is implied: the Mushaf page (0 2 14, 10% black), the continue panel (0 6 24, 5% black), and the tajweed rule card that floats over text (0 8 24, 18% black). No gradients anywhere.

### Named Rules
**The Flat Paper Rule.** Surfaces are flat paper with hairline edges. Separate by tone and line, not by lift; no gradients, no glass, no colored shadows.

## Shapes

Gently rounded, with radius growing with the size of the thing: 4 on the Mushaf's inner frame, 6 on the Mushaf page and guide rows, 8 for buttons, nav items, strip links and the reader title (the theme radius), 10 for surah tiles and the rule card, 12 for stat and shuffle cards, 14 for settings sections, choice cards and the playing verse, 18 for the continue panel. Full pills for the surah number disc, theme swatches, the reciting status and the reader's mode switch. The reader's 3px progress line is the one square edge (radius 0), because it is a rule, not a bar.

## Components

### Buttons
- **Shape:** gently rounded (8px).
- **Primary:** solid accent with accent ink; used once per view for the main act (Continue with a book icon, Begin, Next, Start reading, Shuffle). The continue button pads 12/20.
- **Secondary:** MyGo's default button on Surface, for the alternate path (open in the other mode, All settings).
- **Icon buttons:** an icon alone, always named and tooltipped (back, Aa display, prev/next verse, stop).
- **Labeled buttons:** a Lucide icon before the label.

### Navigation
- **Sidebar:** Surface, 224 wide, right hairline. Items pad 8/10, 17px icon in Muted with a 14px label. Hover fills Hover; the current item gets an accent wash (0.13) with its icon and bold label in accent.
- **Reader bar:** back (Esc), the surah title as a button with a chevron that opens the surah picker popover (340 wide, search field, list), then the play button, the Verses / Mushaf segmented switch, and the Aa display popover (400 wide: theme swatches, size and spacing sliders, translation / tajweed / guide switches).

**The Edge Progress Rule.** Today's goal progress runs as a 3px, square-ended line along the reader bar's bottom edge. The reader shows progress there and in the bar's subtitle, nowhere else.

**The Esc Rule.** Esc leaves the reader for Home; the back button says so in its label. In the Mushaf, the left arrow and Page Down go to the next page (right to left, as the book reads).

### Continue panel (signature)
The first thing on Home: a Surface panel, 18 radius, hairline border, faint ambient shadow, 28px padding. On the left the surah and verse reference (22 bold), its meaning, juz and page in Muted, the verse in Arabic at 30px (two lines at most), and a 220px goal bar with "n of goal verses today". On the right a 210px column with the primary Continue button and the other-mode button under it.

### Home strip
A quiet row under the panel: icon and line pairs (streak, verses read, shuffle, listen from here). Items with an action are links: accent icon, Text label, Hover fill on pointer; items without one have a Muted icon.

### Surah and juz tiles
Rows padding 10/12, radius 10, no fill until hovered. A 30px number disc (Surface with Muted figures; accent with accent ink for the surah being read), the name and a Muted meta line, and the Arabic name in Amiri at 19px on the end. Juz rows end in a 48px progress bar.

### Verse cards
Padding 20/16, separated by a bottom hairline. The verse key (12 bold Muted, tabular), a play-from-here icon button, a Muted check once read; then the Arabic, then the translation. While recited, the verse becomes a rounded panel (14) with an accent wash at 0.06, an accent border at 0.45 and an accent "Now reciting" pill.

### Tajweed text (signature)
**The Fading Emphasis Rule.** Tajweed is taught by fading the surroundings, never by boxes, outlines or backgrounds behind letters. Hovering a colored letter keeps it at full color and fades the rest of that verse (or Mushaf page) to 0.3 alpha. While a verse plays on a Mushaf page, the other verses rest at 0.38. When the guide focuses one rule, other rules' letters lose their color and return to Text. The hover tooltip is the rule card: Page background, hairline, radius 10, 280 wide, a rule-color dot, name, Arabic name in the rule color, and a one-line summary.

### Mushaf page (signature)
The Mushaf sits on Desk. Each page is Page-colored, radius 6, hairline, soft shadow; surah name and juz in Muted Amiri above, page number below in Arabic digits. The 15 lines sit inside a 1.5px accent frame (0.55 alpha) separated from the page edge by a 3px inset ring of page color. Lines are justified edge to edge with kashida, as the printed Mushaf sets them; short lines and the opening pages stay centered. Surah headers are a pill lozenge with an accent hairline (0.7) and an 0.08 wash, the name in accent.

### Choice cards (onboarding)
Equal-width Surface cards, radius 14, padding 18. Border 1 Border at rest, 1 Muted on hover, 2 accent when chosen with an accent check. Each step shows a live preview of its choice.

### Settings sections, stat cards, sliders, swatches
- **Settings section:** a 16 bold heading over a Surface panel (radius 14, padding 18/20, hairline). Rows: a 150px label then the control.
- **Stat card:** Surface, radius 12, padding 16; title 12 bold Muted, value 30 bold tabular, unit 12 Muted.
- **Slider row:** 130px label, slider, 36px right-aligned tabular value.
- **Theme swatch:** a 36px circle of the paper with a 10px accent dot; 3px accent ring when chosen, hairline otherwise; Auto shows half Cream, half Night.

### Player bar
Surface strip with a top hairline, padding 10/20: prev, a 44-wide primary play/pause, next, stop, the verse reference and reciter, and the repeat switch.

### Icons
Lucide, 1.75 stroke, sized by font size (12-18); Muted at rest, accent when current or when the icon is a link.

## Do's and Don'ts

### Do:
- **Do** take every color from the current paper's roles (Page, Desk, Surface, Hover, Border, Text, Muted, Accent) so all five built-in papers and every custom theme work unchanged.
- **Do** keep the accent to the primary action, progress, current selection and the text's ornaments (ayah marks, surah titles, Mushaf frame).
- **Do** show selection as an accent wash (0.12-0.14) with accent bold text.
- **Do** emphasize a tajweed rule by fading its surroundings (0.3 on hover, 0.38 for quiet verses).
- **Do** center section content in a column capped at its width (1280, 960, 880, 720) with 36 / 40 / 48 padding and a 28 gap.
- **Do** name every icon-only button and give it a tooltip.
- **Do** use tabular figures for numbers that align or change.

### Don't:
- **Don't** put boxes, outlines or backgrounds around tajweed letters.
- **Don't** use gradients, glass, or heavy and colored shadows; the three soft ambient shadows (Mushaf page, continue panel, rule card) are the whole vocabulary.
- **Don't** add a second UI typeface or set the UI in a display face.
- **Don't** set Arabic in fonts Core Text cannot mark correctly.
- **Don't** spend the accent on headings, decoration or large fills.
- **Don't** add chrome to the reader beyond back, surah, play, mode and Aa; tajweed and guide switches live in the Aa menu.
