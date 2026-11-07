package main

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/egoist/mygo"
)

// appIcon is the artwork of what plays, where the system shows media.
//
//go:embed resources/icon.png
var appIcon []byte

// listenMedia lets the media keys and the system's media controls (Now
// Playing on macOS, the media flyout on Windows, MPRIS on Linux) drive the
// recitation, which syncMedia shows there.
func (a *app) listenMedia() {
	a.media = true
	mygo.NowPlaying.OnCommand(func(cmd mygo.MediaCommand, _ time.Duration) {
		a.post(func() {
			pb := a.rec.play
			if pb == nil {
				return
			}
			switch cmd {
			case mygo.MediaPlay, mygo.MediaPause:
				if pb.paused == (cmd == mygo.MediaPlay) {
					a.togglePause()
				}
			case mygo.MediaStop:
				a.stopPlayback()
			case mygo.MediaNext:
				a.stepVerse(1)
			case mygo.MediaPrevious:
				a.stepVerse(-1)
			}
		})
	})
}

// syncMedia shows the verse playing where the system shows media, or
// nothing when none plays. A reciter's sample is not shown.
func (a *app) syncMedia() {
	if !a.media {
		return
	}
	pb := a.rec.play
	if pb == nil || pb.preview {
		mygo.NowPlaying.Clear()
		return
	}
	s := a.quran.Surahs[pb.surah-1]
	info := mygo.NowPlayingInfo{
		Title:    fmt.Sprintf("%s %d:%d", s.Name, s.ID, pb.verse),
		Artist:   reciterByID(pb.reciter).label(),
		Album:    "Quranku",
		Artwork:  appIcon,
		State:    mygo.PlaybackPlaying,
		Commands: mygo.MediaPlay | mygo.MediaPause | mygo.MediaStop | mygo.MediaNext | mygo.MediaPrevious,
	}
	if pb.paused {
		info.State = mygo.PlaybackPaused
	}
	if t, ok := pb.snd.(interface {
		Position() time.Duration
		Duration() time.Duration
	}); ok {
		info.Position, info.Duration = t.Position(), t.Duration()
	}
	mygo.NowPlaying.Set(info)
}
