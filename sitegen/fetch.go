// hithisisme - A simple static site generator in Go
// Copyright (C) 2025  Eric Hamiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package sitegen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	
	if strings.Contains(url, "api.github.com") && strings.Contains(url, "/repos") {
		return f.fetchGitHubRepos(target, url)
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if et, ok := f.ETags[url]; ok {
		req.Header.Set("If-None-Match", et)
	}
	req.Header.Set("User-Agent", "hithisisme/0.1")
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

func (f *Fetcher) fetchGitHubRepos(target, baseURL string) ([]byte, error) {
	var allRepos []map[string]interface{}
	page := 1
	perPage := 100
	
	for {
		url := fmt.Sprintf("%s&per_page=%d&page=%d", baseURL, perPage, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "hithisisme/0.1")
		
		resp, err := f.client.Do(req)
		if err != nil {
			return nil, err
		}
		
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			return nil, err
		}
		
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		
		var repos []map[string]interface{}
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, err
		}
		
		if len(repos) == 0 {
			break
		}
		
		for _, repo := range repos {
			if archived, ok := repo["archived"].(bool); ok && archived {
				continue
			}
			allRepos = append(allRepos, repo)
		}
		
		if len(repos) < perPage {
			break
		}
		
		page++
	}
	
	result, err := json.MarshalIndent(allRepos, "", "  ")
	if err != nil {
		return nil, err
	}
	
	path := filepath.Join(f.DataDir, target)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		ioutil.WriteFile(path, result, 0o644)
	}
	
	return result, nil
}
