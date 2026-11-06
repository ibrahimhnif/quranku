package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	quranAPI   = "https://api.quran.com/api/v4"
	audioCDN   = "https://verses.quran.com/"
	userAgent  = "Quranku/0.1"
	maxRequest = 50 << 20 // bytes in one response, at most
)

var httpClient = &http.Client{Timeout: 90 * time.Second}

// background runs work on a goroutine, then the function it returns on the
// UI thread, where it may change the app's state.
func (a *app) background(work func() func()) {
	go func() {
		apply := work()
		if apply != nil {
			a.post(apply)
		}
	}()
}

// getJSON fetches url and decodes its JSON into v.
func (a *app) getJSON(url string, v any) error {
	return a.getJSONContext(context.Background(), url, v)
}

func (a *app) getJSONContext(ctx context.Context, url string, v any) error {
	body, err := a.get(ctx, url)
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(io.LimitReader(body, maxRequest)).Decode(v)
}

// get requests url; the caller closes the body.
func (a *app) get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return resp.Body, nil
}

// download saves url to path, through a temporary file.
func (a *app) download(ctx context.Context, url, path string) error {
	body, err := a.get(ctx, url)
	if err != nil {
		return err
	}
	defer body.Close()
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, io.LimitReader(body, maxRequest))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
