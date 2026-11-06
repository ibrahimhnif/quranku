package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TranslationInfo describes a translation quran.com offers.
type TranslationInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author_name"`
	Lang   string `json:"language_name"`
}

// Translations by their quran.com IDs. Pickthall's English (1930), in the
// public domain, is built into the app; the Indonesian translation of the
// Kementerian Agama RI is downloaded when the app is in Indonesian.
const (
	pickthallID = 19
	kemenagID   = 33
)

var builtinTranslations = []TranslationInfo{
	{ID: pickthallID, Name: "Pickthall", Author: "Mohammed Marmaduke Pickthall", Lang: "english"},
}

// translations holds the translations downloaded, and the catalog.
type translations struct {
	dir     string
	loaded  map[int][]string // by ID: one text a verse
	catalog []TranslationInfo
	// fetching is the ID being downloaded, or -1 for the catalog.
	fetching map[int]bool
	// tried holds the translations downloaded for the language once
	// already, not to try them again on every frame.
	tried map[int]bool
	err   string
}

func newTranslations(dataDir string) *translations {
	tr := &translations{loaded: map[int][]string{}, fetching: map[int]bool{}, tried: map[int]bool{}}
	if dataDir != "" {
		tr.dir = filepath.Join(dataDir, "translations")
	}
	return tr
}

func (tr *translations) path(id int) string {
	return filepath.Join(tr.dir, fmt.Sprintf("%d.json", id))
}

// installed reports whether translation id can be shown.
func (tr *translations) installed(id int) bool {
	if id == pickthallID || tr.loaded[id] != nil {
		return true
	}
	if tr.dir == "" {
		return false
	}
	_, err := os.Stat(tr.path(id))
	return err == nil
}

// text returns a verse's text in translation id, loading it from disk the
// first time; ok is false when it is not installed.
func (tr *translations) text(id int, v *Verse) (string, bool) {
	if id == pickthallID {
		return v.Pickthall, true
	}
	texts, ok := tr.loaded[id]
	if !ok && tr.dir != "" {
		data, err := os.ReadFile(tr.path(id))
		if err == nil && json.Unmarshal(data, &texts) == nil && len(texts) == VerseCount {
			tr.loaded[id] = texts
			ok = true
		}
	}
	if !ok || v.Index >= len(texts) {
		return "", false
	}
	return texts[v.Index], true
}

// translationID is the translation the settings choose: the one picked,
// or the language's.
func (a *app) translationID() int {
	if a.set.TranslationID != 0 {
		return a.set.TranslationID
	}
	if a.set.Lang == Indonesian {
		return kemenagID
	}
	return pickthallID
}

// translationText is a verse's translation as the settings choose it, or
// Pickthall's while that one is not downloaded.
func (a *app) translationText(v *Verse) string {
	if s, ok := a.trans.text(a.translationID(), v); ok {
		return s
	}
	return v.Pickthall
}

// ensureTranslation downloads the translation of the app's language, once,
// when it is not there.
func (a *app) ensureTranslation() {
	tr := a.trans
	id := a.translationID()
	if tr.dir == "" || tr.tried[id] || tr.fetching[id] || tr.installed(id) {
		return
	}
	tr.tried[id] = true
	a.fetchTranslation(id, false)
}

// translationName names the translation shown.
func (a *app) translationName() string {
	id := a.translationID()
	if !a.trans.installed(id) {
		id = pickthallID
	}
	for _, t := range append(builtinTranslations, a.trans.catalog...) {
		if t.ID == id {
			return t.Name
		}
	}
	return fmt.Sprint(id)
}

// loadCatalog fetches the translations quran.com offers, in the
// background.
func (a *app) loadCatalog() {
	tr := a.trans
	if tr.catalog != nil || tr.fetching[-1] {
		return
	}
	tr.fetching[-1] = true
	a.background(func() func() {
		var resp struct {
			Translations []TranslationInfo `json:"translations"`
		}
		err := a.getJSON(a.apiBase+"/resources/translations", &resp)
		return func() {
			delete(tr.fetching, -1)
			if err != nil {
				tr.err = err.Error()
				return
			}
			list := resp.Translations
			sort.SliceStable(list, func(i, j int) bool {
				if list[i].Lang != list[j].Lang {
					return langRank(list[i].Lang) < langRank(list[j].Lang)
				}
				return list[i].Name < list[j].Name
			})
			tr.catalog = list
			tr.err = ""
		}
	})
}

// langRank puts Indonesian, Malay and English first in the catalog.
func langRank(lang string) string {
	switch strings.ToLower(lang) {
	case "indonesian":
		return "0"
	case "malay":
		return "1"
	case "english":
		return "2"
	}
	return "3" + lang
}

// downloadTranslation fetches translation id, saves it, and shows it.
func (a *app) downloadTranslation(id int) { a.fetchTranslation(id, true) }

// fetchTranslation fetches translation id and saves it; choose shows it.
func (a *app) fetchTranslation(id int, choose bool) {
	tr := a.trans
	if tr.dir == "" || tr.fetching[id] {
		return
	}
	tr.fetching[id] = true
	a.background(func() func() {
		var resp struct {
			Translations []struct {
				Text string `json:"text"`
			} `json:"translations"`
		}
		err := a.getJSON(fmt.Sprintf("%s/quran/translations/%d", a.apiBase, id), &resp)
		var texts []string
		if err == nil && len(resp.Translations) != VerseCount {
			err = fmt.Errorf("the translation has %d verses, want %d", len(resp.Translations), VerseCount)
		}
		if err == nil {
			for _, t := range resp.Translations {
				texts = append(texts, cleanTranslation(t.Text))
			}
			var data []byte
			if data, err = json.Marshal(texts); err == nil {
				err = writeFileAtomic(tr.path(id), data)
			}
		}
		return func() {
			delete(tr.fetching, id)
			if err != nil {
				tr.err = err.Error()
				return
			}
			tr.err = ""
			tr.loaded[id] = texts
			if choose {
				a.set.TranslationID = id
			}
		}
	})
}

// deleteTranslation removes a downloaded translation.
func (a *app) deleteTranslation(id int) {
	tr := a.trans
	if id == pickthallID || tr.dir == "" {
		return
	}
	if err := os.Remove(tr.path(id)); err != nil && !os.IsNotExist(err) {
		tr.err = err.Error()
		return
	}
	delete(tr.loaded, id)
	if a.set.TranslationID == id {
		a.set.TranslationID = 0
	}
}
