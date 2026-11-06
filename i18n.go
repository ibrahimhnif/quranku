package main

// Language indices into T.
const (
	English    = 0
	Indonesian = 1
)

// T is a text in English and Indonesian.
type T [2]string

// Interface texts.
var (
	txtSearch      = T{"Search surah or number", "Cari surah atau nomor"}
	txtNoMatch     = T{"No surah matches", "Surah tidak ditemukan"}
	txtVerses      = T{"verses", "ayat"}
	txtMeccan      = T{"Meccan", "Makkiyah"}
	txtMedinan     = T{"Medinan", "Madaniyah"}
	txtTajweed     = T{"Tajweed", "Tajwid"}
	txtTranslation = T{"Translation", "Terjemahan"}
	txtGuide       = T{"Guide", "Panduan"}
	txtTextSize    = T{"Text size", "Ukuran teks"}
	txtLanguage    = T{"Language", "Bahasa"}
	txtGuideTitle  = T{"Tajweed guide", "Panduan Tajwid"}
	txtShowAll     = T{"Show all", "Tampilkan semua"}
	txtHint        = T{
		"Hover a colored letter to see its rule. Click a rule to study it on its own.",
		"Arahkan kursor ke huruf berwarna untuk melihat hukumnya. Klik sebuah hukum untuk mempelajarinya secara khusus.",
	}
	txtLearnMore = T{"Click to learn more", "Klik untuk mempelajari lebih lanjut"}
	txtInSurah   = T{"%d in this surah", "%d kali di surah ini"}
	txtCredits   = T{
		"Quran text and Mushaf pages: Tanzil (tanzil.net). Tajweed: quran-tajweed by Collin Fair (CC BY 4.0). Translations and recitations: quran.com. Fonts: Amiri Quran, Scheherazade New.",
		"Teks Al-Qur'an dan halaman mushaf: Tanzil (tanzil.net). Tajwid: quran-tajweed oleh Collin Fair (CC BY 4.0). Terjemahan dan murattal: quran.com. Font: Amiri Quran, Scheherazade New.",
	}
)

// Texts of the Mushaf and the display settings.
var (
	txtOnPage     = T{"%d on this page", "%d kali di halaman ini"}
	txtVersesMode = T{"Verses", "Ayat"}
	txtMushafMode = T{"Mushaf", "Mushaf"}
	txtMode       = T{"Reading mode", "Mode baca"}
	txtDisplay    = T{"Display", "Tampilan"}
	txtTheme      = T{"Theme", "Tema"}
	txtFont       = T{"Font", "Font"}
	txtSpread     = T{"Two pages", "Dua halaman"}
	txtPage       = T{"Page", "Halaman"}
	txtJuz        = T{"Juz", "Juz"}
	txtNextPage   = T{"Next page", "Halaman berikutnya"}
	txtPrevPage   = T{"Previous page", "Halaman sebelumnya"}
	txtNextShort  = T{"Next", "Berikutnya"}
	txtPrevShort  = T{"Previous", "Sebelumnya"}
)
