package main

import (
	_ "embed"
	"log"

	"github.com/egoist/mygo/ui"
)

// The Quran's fonts, both under the SIL Open Font License (data/*-OFL.txt).
//
// The fonts of the printed Madinah Mushaf (KFGQPC Uthmanic Hafs, Digital
// Khatt) need HarfBuzz: macOS's Core Text, which MyGo's text uses, draws
// them without most of their vowel marks.
var (
	//go:embed data/AmiriQuran-Regular.ttf
	amiriQuran []byte
	//go:embed data/ScheherazadeNew-Regular.ttf
	scheherazade []byte
)

// QuranFont is a font the Quran can be read in.
type QuranFont struct {
	Name   string // as the settings show it
	Family string
	// Scale sizes it to look as large as the others at the same size.
	Scale float32
	// Width is the width of the Mushaf's page in ems of the font: wide
	// enough for nearly every line, whose words then spread to fill it.
	Width float32
}

// The widths are about the 95th percentile of the lines' natural widths
// (TestMushafLineWidths).
var quranFonts = []QuranFont{
	{Name: "Amiri Quran", Family: "Amiri Quran", Scale: 1, Width: 20.5},
	{Name: "Scheherazade", Family: "Scheherazade New", Scale: 1, Width: 21.6},
}

// quranFont is the family of the app's own Arabic texts (names of surahs).
const quranFont = "Amiri Quran"

func init() {
	if err := ui.RegisterFont(amiriQuran, "Amiri Quran"); err != nil {
		log.Fatal(err)
	}
	if err := ui.RegisterFont(scheherazade, "Scheherazade New"); err != nil {
		log.Fatal(err)
	}
}

// ayahMark is the ornament ending verse n: both fonts draw the end of ayah
// sign (U+06DD) around the digits after it.
func ayahMark(n int) string { return "۝" + arabicDigits(n) }
