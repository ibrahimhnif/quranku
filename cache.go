package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Quran Foundation's terms keep the data of its API a week at most:
// translations, and the recitations the user played without downloading
// them. Recitations the user downloaded stay until they delete them.
const (
	maxCacheAge = 7 * 24 * time.Hour
	// refreshAge is when a translation in use is fetched again, a day
	// before it would be deleted.
	refreshAge = 6 * 24 * time.Hour
)

// pruneOlder deletes the files under dir last written before t.
func pruneOlder(dir string, t time.Time) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if fi, err := d.Info(); err == nil && fi.ModTime().Before(t) {
			os.Remove(path)
		}
		return nil
	})
}
