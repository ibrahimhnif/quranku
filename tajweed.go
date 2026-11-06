package main

import "github.com/egoist/mygo/ui"

// Rule describes one tajweed rule as the text marks it.
type Rule struct {
	Key     string // the class name in the source data
	Name    T
	Arabic  string
	Group   *T // family of rules, for the guide
	Summary T  // one line, shown on hover
	Detail  T  // how to recite it, shown in the guide
	Light   ui.Color
	Dark    ui.Color
}

// Color returns the rule's color for the current theme.
func (r *Rule) Color(t *ui.Theme) ui.Color {
	if t.Dark {
		return r.Dark
	}
	return r.Light
}

// Rule groups, in the order the guide shows them.
var (
	groupMadd   = &T{"Madd (prolongation)", "Mad (panjang bacaan)"}
	groupNoon   = &T{"Noon sakinah & tanween", "Nun sukun & tanwin"}
	groupMeem   = &T{"Meem sakinah", "Mim sukun"}
	groupOther  = &T{"Other rules", "Hukum lainnya"}
	groupSilent = &T{"Silent letters", "Huruf yang tidak dibaca"}
	ruleGroups  = []*T{groupMadd, groupNoon, groupMeem, groupOther, groupSilent}
)

// rules lists every rule in the data, with the colors quran.com uses
// (lightened for dark mode so they stay readable).
var rules = []*Rule{
	{
		Key: "madda_normal", Arabic: "مد طبيعي", Group: groupMadd,
		Name:    T{"Natural Madd", "Mad Thabi'i"},
		Summary: T{"Stretch the vowel for 2 counts.", "Dibaca panjang 2 harakat."},
		Detail: T{
			"A madd letter (ا و ي) follows its matching vowel with no hamzah or sukoon after it. Hold the sound for 2 counts — the natural length of the vowel.",
			"Huruf mad (ا و ي) didahului harakat yang sesuai, dan sesudahnya tidak ada hamzah atau sukun. Dibaca panjang 2 harakat, yaitu panjang alami suatu bacaan.",
		},
		Light: ui.Hex("#537FFF"), Dark: ui.Hex("#7FA0FF"),
	},
	{
		Key: "madda_permissible", Arabic: "مد جائز", Group: groupMadd,
		Name:    T{"Permissible Madd", "Mad Jaiz"},
		Summary: T{"Stretch for 2, 4 or 6 counts.", "Boleh dibaca 2, 4 atau 6 harakat."},
		Detail: T{
			"Includes Madd Munfasil (madd letter at the end of a word, hamzah at the start of the next) and Madd 'Arid (stopping on a word after a madd letter). Stretch 2, 4 or 6 counts — be consistent.",
			"Termasuk Mad Jaiz Munfashil (huruf mad di akhir kata, hamzah di awal kata berikutnya) dan Mad 'Aridh Lissukun (berhenti pada kata setelah huruf mad). Dibaca 2, 4 atau 6 harakat — usahakan konsisten.",
		},
		Light: ui.Hex("#4050FF"), Dark: ui.Hex("#8A96FF"),
	},
	{
		Key: "madda_obligatory", Arabic: "مد واجب متصل", Group: groupMadd,
		Name:    T{"Obligatory Madd", "Mad Wajib Muttashil"},
		Summary: T{"Stretch for 4–5 counts.", "Dibaca panjang 4–5 harakat."},
		Detail: T{
			"Madd Muttasil: a madd letter is followed by a hamzah in the same word, e.g. جَاءَ, السَّمَاءِ. Stretch 4 or 5 counts.",
			"Huruf mad bertemu hamzah dalam satu kata, misalnya جَاءَ dan السَّمَاءِ. Wajib dibaca panjang 4 atau 5 harakat.",
		},
		Light: ui.Hex("#2144C1"), Dark: ui.Hex("#6C8CFF"),
	},
	{
		Key: "madda_necessary", Arabic: "مد لازم", Group: groupMadd,
		Name:    T{"Necessary Madd", "Mad Lazim"},
		Summary: T{"Always stretch for 6 counts.", "Selalu dibaca panjang 6 harakat."},
		Detail: T{
			"Madd Lazim: a madd letter is followed by a permanent sukoon or shaddah, e.g. الضَّالِّينَ, and the opening letters like الٓمٓ. Always 6 counts.",
			"Huruf mad bertemu sukun asli atau tasydid, misalnya الضَّالِّينَ, juga huruf-huruf pembuka surah seperti الٓمٓ. Selalu dibaca 6 harakat.",
		},
		Light: ui.Hex("#000EBC"), Dark: ui.Hex("#5A6BFF"),
	},
	{
		Key: "ghunnah", Arabic: "غنة", Group: groupOther,
		Name:    T{"Ghunnah", "Ghunnah"},
		Summary: T{"Nasal sound for 2 counts on نّ / مّ.", "Dengung 2 harakat pada نّ / مّ."},
		Detail: T{
			"Noon or meem with shaddah (نّ مّ). Hold a nasal sound from the nose for 2 counts.",
			"Nun atau mim bertasydid (نّ مّ). Dibaca dengan dengung dari rongga hidung selama 2 harakat.",
		},
		Light: ui.Hex("#FF7E1E"), Dark: ui.Hex("#FF9A4D"),
	},
	{
		Key: "ikhafa", Arabic: "إخفاء", Group: groupNoon,
		Name:    T{"Ikhfa'", "Ikhfa' Haqiqi"},
		Summary: T{"Hide the noon with a light nasal sound, 2 counts.", "Nun dibaca samar dengan dengung, 2 harakat."},
		Detail: T{
			"Noon sakinah or tanween followed by one of 15 letters (ت ث ج د ذ ز س ش ص ض ط ظ ف ق ك). Pronounce the noon partly hidden, between clear and merged, with ghunnah for 2 counts.",
			"Nun sukun atau tanwin bertemu salah satu dari 15 huruf (ت ث ج د ذ ز س ش ص ض ط ظ ف ق ك). Nun dibaca samar, antara jelas dan lebur, disertai dengung 2 harakat.",
		},
		Light: ui.Hex("#9400A8"), Dark: ui.Hex("#D45AE6"),
	},
	{
		Key: "idgham_ghunnah", Arabic: "إدغام بغنة", Group: groupNoon,
		Name:    T{"Idgham with Ghunnah", "Idgham Bighunnah"},
		Summary: T{"Merge the noon into ي ن م و with nasal sound.", "Nun dileburkan ke ي ن م و dengan dengung."},
		Detail: T{
			"Noon sakinah or tanween followed by ي ن م و (in a different word). Merge the noon into the next letter and hold the ghunnah for 2 counts.",
			"Nun sukun atau tanwin bertemu huruf ي ن م و (pada kata yang berbeda). Nun dileburkan ke huruf berikutnya disertai dengung 2 harakat.",
		},
		Light: ui.Hex("#169777"), Dark: ui.Hex("#3CC9A2"),
	},
	{
		Key: "idgham_wo_ghunnah", Arabic: "إدغام بلا غنة", Group: groupNoon,
		Name:    T{"Idgham without Ghunnah", "Idgham Bilaghunnah"},
		Summary: T{"Merge the noon into ل / ر with no nasal sound.", "Nun dileburkan ke ل / ر tanpa dengung."},
		Detail: T{
			"Noon sakinah or tanween followed by ل or ر. Merge it completely into the next letter, without ghunnah.",
			"Nun sukun atau tanwin bertemu huruf ل atau ر. Nun dileburkan sepenuhnya ke huruf berikutnya tanpa dengung.",
		},
		Light: ui.Hex("#169200"), Dark: ui.Hex("#4FCB35"),
	},
	{
		Key: "iqlab", Arabic: "إقلاب", Group: groupNoon,
		Name:    T{"Iqlab", "Iqlab"},
		Summary: T{"Turn the noon into a meem before ب.", "Nun berubah menjadi mim sebelum ب."},
		Detail: T{
			"Noon sakinah or tanween followed by ب. Change the noon sound into a hidden meem with ghunnah for 2 counts. Marked with a small meem (ۢ).",
			"Nun sukun atau tanwin bertemu huruf ب. Bunyi nun diganti menjadi mim yang samar, disertai dengung 2 harakat. Ditandai dengan mim kecil (ۢ).",
		},
		Light: ui.Hex("#26BFFD"), Dark: ui.Hex("#5CD0FF"),
	},
	{
		Key: "ikhafa_shafawi", Arabic: "إخفاء شفوي", Group: groupMeem,
		Name:    T{"Ikhfa' Shafawi", "Ikhfa' Syafawi"},
		Summary: T{"Hide the meem before ب, with ghunnah.", "Mim dibaca samar sebelum ب, dengan dengung."},
		Detail: T{
			"Meem sakinah followed by ب. Close the lips lightly and hide the meem with ghunnah for 2 counts.",
			"Mim sukun bertemu huruf ب. Bibir dirapatkan dengan ringan, mim dibaca samar disertai dengung 2 harakat.",
		},
		Light: ui.Hex("#D500B7"), Dark: ui.Hex("#F05CDA"),
	},
	{
		Key: "idgham_shafawi", Arabic: "إدغام شفوي", Group: groupMeem,
		Name:    T{"Idgham Shafawi", "Idgham Mimi"},
		Summary: T{"Merge meem into the next meem, with ghunnah.", "Mim dileburkan ke mim berikutnya, dengan dengung."},
		Detail: T{
			"Meem sakinah followed by another meem. Merge them into one meem with shaddah and hold the ghunnah for 2 counts.",
			"Mim sukun bertemu mim. Keduanya dilebur menjadi satu mim bertasydid, disertai dengung 2 harakat. Disebut juga Idgham Mitslain.",
		},
		Light: ui.Hex("#58B800"), Dark: ui.Hex("#7FD936"),
	},
	{
		Key: "qalaqah", Arabic: "قلقلة", Group: groupOther,
		Name:    T{"Qalqalah", "Qalqalah"},
		Summary: T{"Bounce the letter ق ط ب ج د when it has sukoon.", "Huruf ق ط ب ج د bersukun dibaca memantul."},
		Detail: T{
			"The letters ق ط ب ج د (قطب جد) with sukoon. Release them with a slight echoing bounce — stronger when stopping at the end of a verse.",
			"Huruf ق ط ب ج د (قطب جد) yang bersukun. Dibaca memantul ringan — lebih kuat (qalqalah kubra) ketika berhenti di akhir ayat.",
		},
		Light: ui.Hex("#DD0008"), Dark: ui.Hex("#FF5A60"),
	},
	{
		Key: "idgham_mutajanisayn", Arabic: "إدغام متجانسين", Group: groupOther,
		Name:    T{"Idgham Mutajanisayn", "Idgham Mutajanisain"},
		Summary: T{"Merge two letters from the same articulation point.", "Dua huruf satu makhraj dileburkan."},
		Detail: T{
			"Two letters share an articulation point but differ in qualities, e.g. ت into ط, د into ت. The first is merged into the second.",
			"Dua huruf yang makhrajnya sama tetapi sifatnya berbeda, misalnya ت dengan ط, atau د dengan ت. Huruf pertama dileburkan ke huruf kedua.",
		},
		Light: ui.Hex("#A1A1A1"), Dark: ui.Hex("#8E8E8E"),
	},
	{
		Key: "idgham_mutaqaribayn", Arabic: "إدغام متقاربين", Group: groupOther,
		Name:    T{"Idgham Mutaqaribayn", "Idgham Mutaqaribain"},
		Summary: T{"Merge two letters from close articulation points.", "Dua huruf yang makhrajnya berdekatan dileburkan."},
		Detail: T{
			"Two letters with close articulation points, e.g. ل into ر, ق into ك (أَلَمْ نَخْلُقكُّم). The first is merged into the second.",
			"Dua huruf yang makhraj dan sifatnya berdekatan, misalnya ل dengan ر, atau ق dengan ك (أَلَمْ نَخْلُقكُّم). Huruf pertama dileburkan ke huruf kedua.",
		},
		Light: ui.Hex("#A1A1A1"), Dark: ui.Hex("#8E8E8E"),
	},
	{
		Key: "ham_wasl", Arabic: "همزة الوصل", Group: groupSilent,
		Name:    T{"Hamzat al-Wasl", "Hamzah Washal"},
		Summary: T{"Connecting hamzah: silent when joining from the previous word.", "Hamzah penghubung: tidak dibaca bila disambung dari kata sebelumnya."},
		Detail: T{
			"The alif with a small ṣād (ٱ). Pronounced only when you start reading from it; when continuing from the previous word it is skipped.",
			"Alif dengan tanda shad kecil (ٱ). Hanya dibaca bila memulai bacaan darinya; bila disambung dengan kata sebelumnya, tidak dibaca.",
		},
		Light: ui.Hex("#AAAAAA"), Dark: ui.Hex("#7A7A7A"),
	},
	{
		Key: "laam_shamsiyah", Arabic: "لام شمسية", Group: groupSilent,
		Name:    T{"Lam Shamsiyyah", "Alif Lam Syamsiyah"},
		Summary: T{"Silent lam: merged into the following sun letter.", "Lam tidak dibaca, lebur ke huruf syamsiyah."},
		Detail: T{
			"The lam of ال before a sun letter (ت ث د ذ ر ز س ش ص ض ط ظ ل ن). The lam is not pronounced; the next letter gets a shaddah instead.",
			"Lam pada ال yang bertemu huruf syamsiyah (ت ث د ذ ر ز س ش ص ض ط ظ ل ن). Lam tidak dibaca, dan huruf sesudahnya dibaca bertasydid.",
		},
		Light: ui.Hex("#AAAAAA"), Dark: ui.Hex("#7A7A7A"),
	},
	{
		Key: "slnt", Arabic: "حرف لا يُنطق", Group: groupSilent,
		Name:    T{"Silent letter", "Huruf tidak dibaca"},
		Summary: T{"Written but not pronounced.", "Ditulis tetapi tidak dibaca."},
		Detail: T{
			"Letters written in the Uthmani script but not read, such as the alif after واو الجماعة (قَالُوا) or extra letters with a small circle above.",
			"Huruf yang tertulis dalam rasm Utsmani tetapi tidak dibaca, seperti alif setelah wau jama'ah (قَالُوا) atau huruf dengan tanda bulatan kecil di atasnya.",
		},
		Light: ui.Hex("#AAAAAA"), Dark: ui.Hex("#7A7A7A"),
	},
}

var rulesByKey = func() map[string]*Rule {
	m := make(map[string]*Rule, len(rules))
	for _, r := range rules {
		m[r.Key] = r
	}
	return m
}()
