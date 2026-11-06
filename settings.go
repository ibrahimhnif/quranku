package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
)

// Reading modes.
const (
	ModeVerses = iota // verse by verse, with translations
	ModeMushaf        // pages of the Madinah Mushaf
)

// Repeat modes of the recitation.
const (
	RepeatOff = iota
	RepeatVerse
	RepeatSurah
)

// Settings are what the app remembers between runs.
type Settings struct {
	Lang  int    `json:"lang"`
	Theme string `json:"themeId"` // a built-in paper's ID, or a custom theme's
	Font  int    `json:"font"`
	Mode  int    `json:"mode"`
	// Size is the Arabic text's size; LineSpacing its line height, in ems.
	Size        float64 `json:"size"`
	LineSpacing float64 `json:"lineSpacing"`
	Page        int     `json:"page"`
	Surah       int     `json:"surah"`
	Colors      bool    `json:"colors"`
	Translation bool    `json:"translation"`
	Guide       bool    `json:"guide"`
	Spread      bool    `json:"spread"` // two pages side by side
	// TranslationID is the quran.com translation shown; 0 follows the
	// app's language.
	TranslationID int     `json:"translationId"`
	TransSize     float64 `json:"translationSize"`
	Reciter       int     `json:"reciter"`
	Repeat        int     `json:"repeat"`
	DailyGoal     float64 `json:"dailyGoal"` // verses a day
	// Themes are the user's own.
	Themes []CustomTheme `json:"themes"`
	// Onboarded is set once the first-run introduction is done or skipped.
	Onboarded bool `json:"onboarded"`
	// LastVerse is where the reader stopped: a verse index, from 0.
	LastVerse int `json:"lastVerse"`
	// MushafZoom scales the Mushaf's page from the size that fits the
	// window (1).
	MushafZoom float64 `json:"mushafZoom"`
}

// CustomTheme is a theme the user made, its colors in hex.
type CustomTheme struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Page    string `json:"page"`
	Surface string `json:"surface"`
	Text    string `json:"text"`
	Accent  string `json:"accent"`
}

func defaultSettings() Settings {
	return Settings{
		Lang: Indonesian, Theme: "auto", Mode: ModeVerses, Size: 34, LineSpacing: 2,
		Page: 1, Surah: 1, Colors: true, Translation: true, Guide: true,
		TransSize: 15, Reciter: 7, DailyGoal: 10, MushafZoom: 1,
	}
}

// clone copies the settings, with their own themes.
func (s Settings) clone() Settings {
	s.Themes = slices.Clone(s.Themes)
	return s
}

func (s Settings) equal(o Settings) bool { return reflect.DeepEqual(s, o) }

// valid keeps settings read from disk within their ranges.
func (s Settings) valid() Settings {
	d := defaultSettings()
	if s.Lang != English && s.Lang != Indonesian {
		s.Lang = d.Lang
	}
	if s.Font < 0 || s.Font >= len(quranFonts) {
		s.Font = 0
	}
	if s.Mode != ModeVerses && s.Mode != ModeMushaf {
		s.Mode = d.Mode
	}
	if s.Size < minSize || s.Size > maxSize {
		s.Size = d.Size
	}
	if s.LineSpacing < 1.4 || s.LineSpacing > 3 {
		s.LineSpacing = d.LineSpacing
	}
	if s.TransSize < 11 || s.TransSize > 26 {
		s.TransSize = d.TransSize
	}
	if s.Page < 1 || s.Page > PageCount {
		s.Page = 1
	}
	if s.Surah < 1 || s.Surah > 114 {
		s.Surah = 1
	}
	if reciterByID(s.Reciter) == nil {
		s.Reciter = d.Reciter
	}
	if s.Repeat < RepeatOff || s.Repeat > RepeatSurah {
		s.Repeat = RepeatOff
	}
	if s.MushafZoom < minMushafZoom || s.MushafZoom > maxMushafZoom {
		s.MushafZoom = d.MushafZoom
	}
	if s.LastVerse < 0 || s.LastVerse >= VerseCount {
		s.LastVerse = 0
	}
	if s.DailyGoal < 1 || s.DailyGoal > 1000 {
		s.DailyGoal = d.DailyGoal
	}
	s.Themes = slices.DeleteFunc(slices.Clone(s.Themes), func(t CustomTheme) bool {
		return t.ID == "" || builtinPaper(t.ID) != nil
	})
	if paperByID(s.Theme, s.Themes) == nil {
		s.Theme = d.Theme
	}
	return s
}

const (
	minSize = 24
	maxSize = 52

	minMushafZoom = 0.5
	maxMushafZoom = 2.5
)

// jsonFile keeps a value in a JSON file. Saves run in the background, the
// latest winning.
type jsonFile struct {
	path    string
	mu      sync.Mutex
	pending []byte
	writing bool
}

func newJSONFile(path string) *jsonFile {
	if path == "" {
		return nil
	}
	return &jsonFile{path: path}
}

// load reads the file into v; a missing file leaves v alone.
func (f *jsonFile) load(v any) error {
	if f == nil {
		return nil
	}
	data, err := os.ReadFile(f.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// save writes v in the background. v is encoded now, so the caller may
// change it after.
func (f *jsonFile) save(v any) {
	if f == nil {
		return
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("%s: %v", f.path, err)
		return
	}
	f.mu.Lock()
	f.pending = data
	if f.writing {
		f.mu.Unlock()
		return
	}
	f.writing = true
	f.mu.Unlock()
	go func() {
		for {
			f.mu.Lock()
			next := f.pending
			f.pending = nil
			if next == nil {
				f.writing = false
				f.mu.Unlock()
				return
			}
			f.mu.Unlock()
			if err := writeFileAtomic(f.path, next); err != nil {
				log.Printf("%s: %v", f.path, err)
			}
		}
	}()
}

// writeFileAtomic writes a file through a temporary one, so a crash leaves
// the old file or the new, never half of one.
func writeFileAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// loadSettings reads the settings, or returns the defaults.
func loadSettings(f *jsonFile) Settings {
	set := defaultSettings()
	if err := f.load(&set); err != nil {
		log.Printf("settings: %v", err)
		return defaultSettings()
	}
	return set.valid()
}
