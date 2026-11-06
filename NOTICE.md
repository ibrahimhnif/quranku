# Third-party content

Quranku's source code is under the MIT License (`LICENSE`). The content and
assets below are not; they keep their own terms, and are included with
attribution.

| What | Where | Source | Terms |
|---|---|---|---|
| Quran text (Uthmani script) | `data/quran-uthmani.txt` | [Tanzil Project](https://tanzil.net) | Creative Commons Attribution 3.0, with Tanzil's terms: verbatim copies only, source credited with a link to tanzil.net. The file is unchanged, its copyright notice included. |
| Surahs, juz, hizb quarters and the start of each Madinah Mushaf page | `data/quran-data.js` | [Tanzil Project](https://tanzil.net/docs/quran_metadata) | Creative Commons Attribution 3.0; the file is unchanged |
| Tajweed rules of each verse | `data/tajweed.json`, built by `scripts/build_tajweed.py` | [quran-tajweed](https://github.com/cpfair/quran-tajweed) by Collin Fair | Creative Commons Attribution 4.0. Changed: the ranges are moved from the 2017 Tanzil text onto the current one, and the rules renamed. |
| English translation by Mohammed Marmaduke Pickthall (1930) | `data/en.pickthall.txt` | Text from [Tanzil](https://tanzil.net/trans/) | The translation is in the public domain (first published 1930; the translator died in 1936). |
| Amiri Quran font | `data/AmiriQuran-Regular.ttf` | [Amiri Project](https://github.com/aliftype/amiri) | SIL Open Font License 1.1 (`data/AmiriQuran-OFL.txt`) |
| Scheherazade New font | `data/ScheherazadeNew-Regular.ttf` | [SIL Global](https://software.sil.org/scheherazade/) | SIL Open Font License 1.1 (`data/ScheherazadeNew-OFL.txt`) |
| Icons | `data/icons/*.svg` | [Lucide](https://lucide.dev) | ISC License (`data/icons/LICENSE`) |

The surahs' names in Latin letters and their Indonesian meanings
(`surahs.go`) are the common ones. The tajweed rule colors follow the palette
quran.com uses. The Mushaf's line breaks are computed by Quranku; they are
not the printed Mushaf's.

## Downloaded at runtime

On the user's request, or for the app's language, Quranku downloads
translations (the Kementerian Agama RI Indonesian translation among them) and
verse-by-verse recitations from the [quran.com](https://quran.com) API and
audio CDN into the user's own data folder. They are not part of this
repository and keep their publishers' terms.

If you hold rights to any of this content and want it changed or removed,
please open an issue.
