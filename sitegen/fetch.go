package sitegen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Fetcher handles HTTP fetching with ETag caching.
type Fetcher struct {
	DataDir string
	ETags   map[string]string
	client  *http.Client
}

func NewFetcher(dataDir string) *Fetcher {
	return &Fetcher{
		DataDir: dataDir,
		ETags:   make(map[string]string),
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (f *Fetcher) LoadETags() {
	path := filepath.Join(f.DataDir, ".etag.json")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(b, &f.ETags)
}

func (f *Fetcher) SaveETags() {
	path := filepath.Join(f.DataDir, ".etag.json")
	b, _ := json.MarshalIndent(f.ETags, "", "  ")
	_ = ioutil.WriteFile(path, b, 0o644)
}

func (f *Fetcher) Fetch(target, url string) ([]byte, error) {
	if f.client == nil {
		f.client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if et, ok := f.ETags[url]; ok {
		req.Header.Set("If-None-Match", et)
	}
	req.Header.Set("User-Agent", "sitegen/0.1")
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if et := resp.Header.Get("ETag"); et != "" {
			f.ETags[url] = et
		}
		path := filepath.Join(f.DataDir, target)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
			ioutil.WriteFile(path, body, 0o644)
		}
		return body, nil
	case http.StatusNotModified:
		path := filepath.Join(f.DataDir, target)
		return ioutil.ReadFile(path)
	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}
