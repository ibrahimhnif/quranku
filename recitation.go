package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egoist/mygo/audio"
	_ "github.com/egoist/mygo/audio/mp3" // the recitations are MP3
	"github.com/egoist/mygo/ui"
)

// Reciter is a recitation quran.com serves, a file a verse.
type Reciter struct {
	ID          int
	Name, Style string
}

var reciters = []Reciter{
	{7, "Mishari Rashid al-Afasy", ""},
	{3, "Abdur-Rahman as-Sudais", ""},
	{10, "Sa'ud ash-Shuraym", ""},
	{4, "Abu Bakr al-Shatri", ""},
	{5, "Hani ar-Rifai", ""},
	{2, "AbdulBaset AbdulSamad", "Murattal"},
	{1, "AbdulBaset AbdulSamad", "Mujawwad"},
	{6, "Mahmoud Khalil Al-Husary", ""},
	{12, "Mahmoud Khalil Al-Husary", "Muallim"},
	{9, "Mohamed Siddiq al-Minshawi", "Murattal"},
	{8, "Mohamed Siddiq al-Minshawi", "Mujawwad"},
	{11, "Mohamed al-Tablawi", ""},
}

func reciterByID(id int) *Reciter {
	for i := range reciters {
		if reciters[i].ID == id {
			return &reciters[i]
		}
	}
	return nil
}

func (r *Reciter) label() string {
	if r.Style != "" {
		return r.Name + " (" + r.Style + ")"
	}
	return r.Name
}

// sound is a playing recitation: an *audio.Player, or a fake in tests.
type sound interface {
	Play()
	Pause()
	Close()
	OnEnd(func()) (off func())
}

func openMP3(path string) (sound, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	p, err := audio.OpenStream(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return p, nil
}

type surahKey struct{ reciter, surah int }

// dlJob is a surah downloading.
type dlJob struct {
	done, total int
	cancel      context.CancelFunc
}

// recitation is the recitations' state: their files, downloads and what
// plays.
type recitation struct {
	dir    string
	jobs   map[surahKey]*dlJob
	queue  []surahKey       // surahs waiting to download, one at a time
	counts map[surahKey]int // files on disk, as last counted
	err    string
	open   func(path string) (sound, error)

	play *playback
	gen  int // which playback a sound's end belongs to
}

// playback is what plays: verses from verse to until of a surah.
type playback struct {
	reciter, surah, verse, from, until int
	snd                                sound
	paused, loading                    bool
	err                                string
	// preview is a reciter's sample, heard while choosing: it plays one
	// verse, counts as nothing, and moves no page.
	preview bool
}

func newRecitation(dataDir string) *recitation {
	r := &recitation{jobs: map[surahKey]*dlJob{}, counts: map[surahKey]int{}, open: openMP3}
	if dataDir != "" {
		r.dir = filepath.Join(dataDir, "audio")
	}
	return r
}

func (r *recitation) path(reciter, surah, verse int) string {
	return filepath.Join(r.dir, fmt.Sprint(reciter), fmt.Sprintf("%03d%03d.mp3", surah, verse))
}

func (r *recitation) has(reciter, surah, verse int) bool {
	_, err := os.Stat(r.path(reciter, surah, verse))
	return r.dir != "" && err == nil
}

// count counts a surah's verses on disk, remembering it until downloads
// change it.
func (a *app) audioCount(reciter int, s *Surah) int {
	k := surahKey{reciter, s.ID}
	if n, ok := a.rec.counts[k]; ok {
		return n
	}
	n := 0
	for v := 1; v <= len(s.Verses); v++ {
		if a.rec.has(reciter, s.ID, v) {
			n++
		}
	}
	a.rec.counts[k] = n
	return n
}

// audioURLs fetches the files of a surah's verses: verse number -> URL.
func (a *app) audioURLs(ctx context.Context, reciter, surah int) (map[int]string, error) {
	urls := map[int]string{}
	for page := 1; page > 0; {
		var resp struct {
			AudioFiles []struct {
				Key string `json:"verse_key"`
				URL string `json:"url"`
			} `json:"audio_files"`
			Pagination struct {
				Next *int `json:"next_page"`
			} `json:"pagination"`
		}
		url := fmt.Sprintf("%s/recitations/%d/by_chapter/%d?per_page=50&page=%d", a.apiBase, reciter, surah, page)
		if err := a.getJSONContext(ctx, url, &resp); err != nil {
			return nil, err
		}
		for _, f := range resp.AudioFiles {
			var s, v int
			if _, err := fmt.Sscanf(f.Key, "%d:%d", &s, &v); err != nil || s != surah {
				continue
			}
			u := f.URL
			switch {
			case strings.HasPrefix(u, "//"):
				u = "https:" + u
			case !strings.HasPrefix(u, "http"):
				u = a.audioBase + u
			}
			urls[v] = u
		}
		page = 0
		if resp.Pagination.Next != nil {
			page = *resp.Pagination.Next
		}
	}
	return urls, nil
}

// downloadSurah downloads a surah's verses that are not on disk.
func (a *app) downloadSurah(reciter int, s *Surah) {
	k := surahKey{reciter, s.ID}
	if a.rec.dir == "" || a.rec.jobs[k] != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	job := &dlJob{total: len(s.Verses), cancel: cancel}
	a.rec.jobs[k] = job
	missing := []int{}
	for v := 1; v <= len(s.Verses); v++ {
		if !a.rec.has(reciter, s.ID, v) {
			missing = append(missing, v)
		}
	}
	job.done = job.total - len(missing)
	a.background(func() func() {
		err := a.fetchVerses(ctx, reciter, s.ID, missing, func() {
			a.post(func() { job.done++ })
		})
		return func() {
			cancel()
			delete(a.rec.jobs, k)
			delete(a.rec.counts, k)
			if err != nil && ctx.Err() == nil {
				a.rec.err = err.Error()
			}
			a.nextQueued()
		}
	})
}

// fetchVerses downloads verses of a surah, four at a time, calling done
// after each.
func (a *app) fetchVerses(ctx context.Context, reciter, surah int, verses []int, done func()) error {
	if len(verses) == 0 {
		return nil
	}
	urls, err := a.audioURLs(ctx, reciter, surah)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.rec.path(reciter, surah, 1)), 0o755); err != nil {
		return err
	}
	work := make(chan int)
	var mu sync.Mutex
	var first error
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range work {
				url, ok := urls[v]
				err := fmt.Errorf("no recitation of %d:%d", surah, v)
				if ok {
					err = a.download(ctx, url, a.rec.path(reciter, surah, v))
				}
				if err != nil {
					mu.Lock()
					first = cmpOr(first, err)
					mu.Unlock()
					continue
				}
				done()
			}
		}()
	}
	for _, v := range verses {
		select {
		case work <- v:
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			break
		}
	}
	close(work)
	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return first
}

func cmpOr(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

// downloadAll downloads every surah not complete, one after another.
func (a *app) downloadAll(reciter int) {
	for _, s := range a.quran.Surahs {
		k := surahKey{reciter, s.ID}
		if a.rec.jobs[k] == nil && a.audioCount(reciter, s) < len(s.Verses) && !containsKey(a.rec.queue, k) {
			a.rec.queue = append(a.rec.queue, k)
		}
	}
	if len(a.rec.jobs) == 0 {
		a.nextQueued()
	}
}

func containsKey(ks []surahKey, k surahKey) bool {
	for _, x := range ks {
		if x == k {
			return true
		}
	}
	return false
}

// nextQueued starts the next surah waiting.
func (a *app) nextQueued() {
	for len(a.rec.queue) > 0 && len(a.rec.jobs) == 0 {
		k := a.rec.queue[0]
		a.rec.queue = a.rec.queue[1:]
		a.downloadSurah(k.reciter, a.quran.Surahs[k.surah-1])
	}
}

// cancelDownloads stops every download, and what waits.
func (a *app) cancelDownloads() {
	a.rec.queue = nil
	for _, j := range a.rec.jobs {
		j.cancel()
	}
}

// deleteSurahAudio removes a surah's files.
func (a *app) deleteSurahAudio(reciter int, s *Surah) {
	for v := 1; v <= len(s.Verses); v++ {
		if err := os.Remove(a.rec.path(reciter, s.ID, v)); err != nil && !os.IsNotExist(err) {
			a.rec.err = err.Error()
		}
	}
	delete(a.rec.counts, surahKey{reciter, s.ID})
}

// playFrom plays a surah's verses from verse to until (0: its end).
func (a *app) playFrom(surah, verse, until int) {
	a.stopPlayback()
	s := a.quran.Surahs[surah-1]
	if until <= 0 || until > len(s.Verses) {
		until = len(s.Verses)
	}
	a.rec.play = &playback{reciter: a.set.Reciter, surah: surah, verse: verse, from: verse, until: until}
	a.startVerse()
}

// previewReciter plays a reciter's sample: the first verse of Al-Fatihah.
func (a *app) previewReciter(id int) {
	a.stopPlayback()
	a.rec.play = &playback{reciter: id, surah: 1, verse: 1, from: 1, until: 1, preview: true}
	a.startVerse()
}

// startVerse plays the verse the playback is on, downloading it first when
// it is not on disk.
func (a *app) startVerse() {
	pb := a.rec.play
	if pb == nil {
		return
	}
	a.rec.gen++
	gen := a.rec.gen
	if pb.snd != nil {
		pb.snd.Close()
		pb.snd = nil
	}
	pb.paused, pb.err = false, ""
	a.followPlayback()
	defer a.syncMedia()
	if !a.rec.has(pb.reciter, pb.surah, pb.verse) {
		if a.rec.dir == "" {
			pb.err = "no place to keep recitations"
			return
		}
		pb.loading = true
		reciter, surah, verse := pb.reciter, pb.surah, pb.verse
		a.background(func() func() {
			err := a.fetchVerses(context.Background(), reciter, surah, []int{verse}, func() {})
			return func() {
				delete(a.rec.counts, surahKey{reciter, surah})
				if a.rec.play != pb || a.rec.gen != gen {
					return // something else plays now
				}
				pb.loading = false
				if err != nil {
					pb.err = err.Error()
					return
				}
				a.startVerse()
			}
		})
		return
	}
	pb.loading = false
	snd, err := a.rec.open(a.rec.path(pb.reciter, pb.surah, pb.verse))
	if err != nil {
		pb.err = err.Error()
		return
	}
	pb.snd = snd
	snd.OnEnd(func() {
		a.post(func() {
			if a.rec.play == pb && a.rec.gen == gen {
				a.verseEnded()
			}
		})
	})
	snd.Play()
}

// verseEnded goes on after a verse: again, to the next, or back to the
// start, as the repeat mode says.
func (a *app) verseEnded() {
	pb := a.rec.play
	if pb.preview {
		a.stopPlayback()
		return
	}
	a.reading.markListened(pb.reciter, a.now())
	switch {
	case a.set.Repeat == RepeatVerse:
	case pb.verse < pb.until:
		pb.verse++
	case a.set.Repeat == RepeatSurah:
		pb.verse = pb.from
	default:
		a.stopPlayback()
		return
	}
	a.startVerse()
}

// stepVerse plays the verse d after the one playing.
func (a *app) stepVerse(d int) {
	pb := a.rec.play
	if pb == nil {
		return
	}
	n := len(a.quran.Surahs[pb.surah-1].Verses)
	pb.verse = min(max(pb.verse+d, 1), n)
	pb.until = max(pb.until, pb.verse)
	a.startVerse()
}

func (a *app) togglePause() {
	pb := a.rec.play
	if pb == nil || pb.snd == nil {
		return
	}
	if pb.paused {
		pb.snd.Play()
	} else {
		pb.snd.Pause()
	}
	pb.paused = !pb.paused
	a.syncMedia()
}

func (a *app) stopPlayback() {
	if pb := a.rec.play; pb != nil && pb.snd != nil {
		pb.snd.Close()
	}
	a.rec.play = nil
	a.rec.gen++
	a.syncMedia()
}

// followPlayback keeps the verse playing in view.
func (a *app) followPlayback() {
	pb := a.rec.play
	if pb == nil || pb.preview || a.section != SectionRead {
		return
	}
	v := a.quran.Verse(pb.surah, pb.verse)
	if a.set.Mode == ModeMushaf {
		for _, p := range a.shownPages() {
			if p.Number == v.Page {
				return
			}
		}
		a.goPage(v.Page)
		return
	}
	if a.surah.ID == pb.surah {
		a.verses.ScrollIntoView(pb.verse)
	}
}

// recitedOn reports whether the verse playing is on page p.
func (a *app) recitedOn(p *Page) bool {
	pb := a.rec.play
	if pb == nil || pb.preview {
		return false
	}
	v := a.quran.Verse(pb.surah, pb.verse)
	return v != nil && v.Page == p.Number
}

// playing reports whether verse v is the one playing.
func (a *app) playing(v *Verse) bool {
	pb := a.rec.play
	return pb != nil && !pb.preview && pb.surah == v.Surah && pb.verse == v.Number
}

// Texts of the recitations.
var (
	txtRecitations  = T{"Recitations", "Murattal"}
	txtRecIntro     = T{"Choose a reciter. Download surahs to listen offline; a verse not downloaded is fetched as it plays.", "Pilih qari. Unduh surah untuk didengar tanpa internet; ayat yang belum diunduh diambil saat diputar."}
	txtReciter      = T{"Reciter", "Qari"}
	txtDownload     = T{"Download", "Unduh"}
	txtDownloadAll  = T{"Download all", "Unduh semua"}
	txtCancelAll    = T{"Cancel downloads", "Batalkan unduhan"}
	txtDownloaded   = T{"Downloaded", "Terunduh"}
	txtWaiting      = T{"Waiting", "Menunggu"}
	txtSurahsReady  = T{"surahs downloaded", "surah terunduh"}
	txtRepeat       = T{"Repeat", "Ulangi"}
	txtRepeatOff    = T{"Off", "Mati"}
	txtRepeatVerse  = T{"Verse", "Ayat"}
	txtRepeatSurah  = T{"All", "Semua"}
	txtLoading      = T{"Loading…", "Memuat…"}
	txtPlaySurah    = T{"Play", "Putar"}
	txtPlayFromHere = T{"Play from this verse", "Putar dari ayat ini"}
	txtNowReciting  = T{"Now reciting", "Sedang dibacakan"}
	txtPaused       = T{"Paused", "Dijeda"}
	txtPause        = T{"Pause", "Jeda"}
	txtPrevVerse    = T{"Previous verse", "Ayat sebelumnya"}
	txtNextVerse    = T{"Next verse", "Ayat berikutnya"}
	txtStop         = T{"Stop", "Berhenti"}
	txtResume       = T{"Resume", "Lanjutkan"}
)

// recitationPage lets the user choose a reciter, and download surahs.
func (a *app) recitationPage(c *ui.Context) {
	t := c.Theme()
	rec := reciterByID(a.set.Reciter)
	ready := 0
	for _, s := range a.quran.Surahs {
		if a.audioCount(rec.ID, s) == len(s.Verses) {
			ready++
		}
	}
	ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
		// Reciters.
		ui.Scroll(c).Width(300).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
			ui.Column(c).Padding(24, 16).Gap(4).Children(func() {
				ui.Text(c, a.tx(txtRecitations)).FontSize(24).Bold().Padding(0, 8, 8, 8)
				ui.Text(c, a.tx(txtRecIntro)).FontSize(12).TextColor(t.TextMuted).Padding(0, 8, 12, 8)
				ui.Text(c, a.tx(txtReciter)).FontSize(11).Bold().TextColor(t.TextMuted).Padding(4, 8)
				for _, r := range reciters {
					row := ui.Row(c.Key(r.ID)).Padding(8, 10).Gap(8).Radius(8).Cursor(ui.CursorPointer)
					chosen := r.ID == a.set.Reciter
					switch {
					case chosen:
						row.Background(t.Accent.Alpha(0.14))
					case row.Hovered():
						row.Background(t.SurfaceHover)
					}
					if row.Clicked() {
						a.set.Reciter = r.ID
					}
					row.Children(func() {
						ui.Column(c).Grow(1).Children(func() {
							name := ui.Text(c, r.Name).FontSize(13)
							if chosen {
								name.Bold().TextColor(t.Accent)
							}
							if r.Style != "" {
								ui.Text(c, r.Style).FontSize(11).TextColor(t.TextMuted)
							}
						})
					})
				}
			})
		})
		// Surahs.
		ui.Column(c).Grow(1).Children(func() {
			ui.Row(c).Padding(20, 24, 12, 24).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Grow(1).Gap(2).Children(func() {
					ui.Text(c, rec.label()).FontSize(17).Bold()
					ui.Textf(c, "%d / 114 %s", ready, a.tx(txtSurahsReady)).FontSize(12).TextColor(t.TextMuted)
				})
				if len(a.rec.jobs) > 0 || len(a.rec.queue) > 0 {
					if ui.Button(c, a.tx(txtCancelAll)).Clicked() {
						a.cancelDownloads()
					}
				} else if ready < 114 && labeledButton(c, iconDownload, a.tx(txtDownloadAll), true).Disabled(a.rec.dir == "").Clicked() {
					a.downloadAll(rec.ID)
				}
			})
			if a.rec.err != "" {
				ui.Text(c, a.rec.err).FontSize(12).TextColor(t.Danger).Padding(0, 24, 8, 24)
			}
			ui.List(c, &a.audioList, len(a.quran.Surahs), func(i int) {
				a.audioRow(c, rec.ID, a.quran.Surahs[i])
			}).Grow(1).Padding(0, 16, 24, 16)
		})
	})
}

// audioRow is a surah's row: its downloaded verses, and its buttons.
func (a *app) audioRow(c *ui.Context, reciter int, s *Surah) {
	t := c.Theme()
	k := surahKey{reciter, s.ID}
	job := a.rec.jobs[k]
	n := a.audioCount(reciter, s)
	ui.Row(c.Key(s.ID)).Padding(8).Gap(12).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.Text(c, fmt.Sprint(s.ID)).FontSize(12).TextColor(t.TextMuted).Width(28).TextAlign(ui.Center)
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, s.Name).FontSize(14)
			switch {
			case job != nil:
				ui.Progress(c, float64(job.done)/float64(max(job.total, 1))).Label(s.Name)
			case containsKey(a.rec.queue, k):
				ui.Text(c, a.tx(txtWaiting)).FontSize(11).TextColor(t.TextMuted)
			case n == len(s.Verses):
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconCheck).FontSize(12).TextColor(t.Success)
					ui.Text(c, a.tx(txtDownloaded)).FontSize(11).TextColor(t.Success)
				})
			default:
				ui.Textf(c, "%d / %d %s", n, len(s.Verses), a.tx(txtVerses)).FontSize(11).TextColor(t.TextMuted)
			}
		})
		ui.Text(c, s.Arabic).Font(quranFont).FontSize(18)
		if iconButton(c, iconPlay, a.tx(txtPlay)+" "+s.Name).Clicked() {
			a.playFrom(s.ID, 1, 0)
		}
		switch {
		case job != nil:
			if ui.Button(c, a.tx(txtCancel)).Clicked() {
				job.cancel()
			}
		case n == len(s.Verses):
			if ui.Button(c, a.tx(txtDelete)).Clicked() {
				a.deleteSurahAudio(reciter, s)
			}
		default:
			if labeledButton(c, iconDownload, a.tx(txtDownload), false).Disabled(a.rec.dir == "").Clicked() {
				a.downloadSurah(reciter, s)
			}
		}
	})
}

// playerBar controls what plays.
func (a *app) playerBar(c *ui.Context) {
	pb := a.rec.play
	if pb == nil || pb.preview {
		return // a sample shows its controls where it was asked for
	}
	t := c.Theme()
	s := a.quran.Surahs[pb.surah-1]
	ui.Row(c).Padding(10, 20).Gap(10).AlignItems(ui.Center).Background(t.Surface).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
		if iconButton(c, iconPrev, a.tx(txtPrevVerse)).Clicked() {
			a.stepVerse(-1)
		}
		svg, label := iconPause, a.tx(txtPause)
		if pb.paused {
			svg, label = iconPlay, a.tx(txtResume)
		}
		if ui.PrimaryButton(c, "").Label(label).Tooltip(label).Width(44).Children(func() { ui.Icon(c, svg) }).Clicked() {
			a.togglePause()
		}
		if iconButton(c, iconNext, a.tx(txtNextVerse)).Clicked() {
			a.stepVerse(1)
		}
		if iconButton(c, iconStop, a.tx(txtStop)).Clicked() {
			a.stopPlayback()
			return
		}
		ui.Column(c).Grow(1).Gap(2).Children(func() {
			ui.Textf(c, "%s %d:%d", s.Name, s.ID, pb.verse).FontSize(14).Bold()
			status := reciterByID(pb.reciter).label()
			switch {
			case pb.err != "":
				status = pb.err
			case pb.loading:
				status = a.tx(txtLoading)
			}
			st := ui.Text(c, status).FontSize(12).TextColor(t.TextMuted).SingleLine()
			if pb.err != "" {
				st.TextColor(t.Danger)
			}
		})
		ui.Icon(c, iconRepeat).FontSize(14).TextColor(t.TextMuted).Label(a.tx(txtRepeat))
		ui.Segmented(c, &a.set.Repeat, a.tx(txtRepeatOff), a.tx(txtRepeatVerse), a.tx(txtRepeatSurah)).Label(a.tx(txtRepeat))
	})
}
