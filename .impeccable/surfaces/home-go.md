---
version: 1
slug: "home-go"
primary_target: "home.go"
related_targets: ["onboarding.go","settings_page.go","reader.go"]
---

Scope: the app's flow: first-run onboarding, the "Hari ini" home, the focused reader, and the Settings page. Mode: operate (the reader itself reads like Read mode).

Audience and job: daily readers (mostly Indonesian) opening the app at a set moment to continue tilawah and keep a daily goal. Task: continue where they stopped in one click; set the app up once in about a minute.

Decisions made while building:
- The FIRST VIEWPORT's "juz chips" became a Surah / Juz tab: 30 chips above 114 rows crowded the index, and the juz view needs no search (30 entries, each showing where it starts and how much of it was read).
- Tajweed and guide switches moved from the reader bar into the Aa display menu, so the bar keeps only back, surah, play, mode and Aa.

Constraints: native MyGo UI (no web); Arabic only in Amiri Quran / Scheherazade New; existing themes (Auto, White, Cream, Green, Night, custom) and tajweed rule colors stay; tajweed emphasis by fading, never boxes.

## Direction contract

THESIS: The app opens on today's reading, not on a tool. A "Hari ini" home puts the continue point (the verse itself, in Arabic) and today's goal first; reading is a separate, full-window focus view you enter and leave with Esc. Refuses the category default of a permanent three-pane browser (surah list, text, panel) as the first and only screen.

OWN-WORLD: The incumbent paper themes (Cream default), one accent per theme, the user's own text as the only ornament. Drawn line icons (Lucide, 1.75 stroke), one sans for UI, Amiri for Arabic. Restrained colour: accent only for the primary action, progress and current selection.

STORY: The reader sees where they stopped and how much is left today, presses Continue, reads in a quiet full window with progress at the edge, returns home with Esc; first run asks only language, reading style, look, reciter and goal, each with a live preview.

FIRST VIEWPORT: Home: greeting line with the date; a wide continue panel holding the surah and verse reference, the verse's Arabic at reading size, today's progress line (n of goal verses) and the Continue button at its right; below, a quiet strip of streak / verses read / shuffle / recitation links; then the surah browser (search, juz chips, list). Sidebar holds sections only.

FORM: Beranda "Hari ini" with a focused reader, position 3 on the ordered list. seed:b9304adec10c1f7d

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
