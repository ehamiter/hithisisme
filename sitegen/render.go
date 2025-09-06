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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

// RenderOptions holds CLI options.
type RenderOptions struct {
	Input   string
	Out     string
	DataDir string
	Layout  string
}

// Render performs full render pipeline.
func Render(opts RenderOptions) error {
	in, err := os.ReadFile(opts.Input)
	if err != nil {
		return err
	}
	bindings, nodes, err := Parse(bytes.NewReader(in))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(opts.DataDir, 0o755); err != nil {
		return err
	}
	fetcher := NewFetcher(opts.DataDir)
	fetcher.LoadETags()
	ctx := &context{
		bindings: make(map[string]interface{}),
		lazy:     make(map[string]*lazyBinding),
		fetcher:  fetcher,
	}
	// resolve bindings
	for _, b := range bindings {
		if b.URL != "" && !b.Lazy {
			body, err := fetcher.Fetch(b.Target, b.URL)
			if err != nil {
				// if file exists use cache
				path := filepath.Join(opts.DataDir, b.Target)
				if body, err = os.ReadFile(path); err != nil {
					return err
				}
			}
			var v interface{}
			if err := json.Unmarshal(body, &v); err != nil {
				return err
			}
			ctx.bindings[b.Name] = v
			continue
		}
		if b.URL != "" && b.Lazy {
			lb := &lazyBinding{Target: b.Target, Template: b.URL, Data: make(map[string]interface{}), Fetched: make(map[string]bool)}
			// load cache
			path := filepath.Join(opts.DataDir, b.Target)
			if data, err := os.ReadFile(path); err == nil {
				json.Unmarshal(data, &lb.Data)
			}
			ctx.lazy[b.Name] = lb
			continue
		}
		if b.Manual {
			path := filepath.Join(opts.DataDir, b.Target)
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var v interface{}
			if err := json.Unmarshal(body, &v); err != nil {
				return err
			}
			ctx.bindings[b.Name] = v
		}
	}
	var buf strings.Builder
	if err := ctx.renderNodes(nodes, make(map[string]interface{}), &buf); err != nil {
		return err
	}
	// load layout
	layout, err := os.ReadFile(opts.Layout)
	if err != nil {
		return err
	}
	bodyHTML := buf.String()
	outHTML := strings.Replace(string(layout), "<!--CONTENT-->", bodyHTML, 1)
	// last updated: use now with readable format
	outHTML = strings.Replace(outHTML, "<!--LAST_UPDATED-->", time.Now().Format("January 2, 2006"), 1)
	if err := os.MkdirAll(filepath.Dir(opts.Out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(opts.Out, []byte(outHTML), 0o644); err != nil {
		return err
	}
	
	// Generate dynamic CSS with timestamp-based color
	if err := generateCSS(filepath.Dir(opts.Out)); err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}
	
	fetcher.SaveETags()
	// save lazy caches
	for _, lb := range ctx.lazy {
		path := filepath.Join(opts.DataDir, lb.Target)
		b, _ := json.MarshalIndent(lb.Data, "", "  ")
		os.WriteFile(path, b, 0o644)
	}
	return nil
}

type context struct {
	bindings map[string]interface{}
	lazy     map[string]*lazyBinding
	fetcher  *Fetcher
}

type lazyBinding struct {
	Target   string
	Template string
	Data     map[string]interface{}
	Fetched  map[string]bool
}

func (c *context) renderNodes(nodes []Node, vars map[string]interface{}, buf *strings.Builder) error {
	var heroRendered bool
	var tabsAdded bool
	
	for _, n := range nodes {
		switch t := n.(type) {
		case Section:
			md := goldmark.New()
			var h bytes.Buffer
			if err := md.Convert([]byte(t.Text), &h); err != nil {
				return err
			}
			
			// Special handling for hero section
			if t.ID == "hero" {
				buf.WriteString(fmt.Sprintf(`<section class="hero is-dark is-medium">
  <div class="hero-body">
    <div class="container">
      <h1 class="title is-1 has-text-white">%s</h1>
    </div>
  </div>
</section>`, strings.TrimPrefix(strings.TrimSuffix(h.String(), "</p>\n"), "<p>")))
				heroRendered = true
			} else {
				// Add tabs after hero but before other sections
				if heroRendered && !tabsAdded {
					buf.WriteString(`
<div class="container">
  <div class="tabs is-centered is-large is-boxed has-text-weight-semibold is-family-code is-lowercase">
    <ul>
      <li class="is-active" data-tab="things">
        <a>
          <span>Things</span>
        </a>
      </li>
      <li data-tab="apps">
        <a>
          <span>Apps</span>
        </a>
      </li>
      <li data-tab="repos">
        <a>
          <span>Repos</span>
        </a>
      </li>
    </ul>
  </div>
</div>`)
					tabsAdded = true
				}
				
				// Check if this section should be wrapped in a tab content div
				if t.ID == "things" || t.ID == "apps" || t.ID == "repos" {
					titleText := strings.TrimPrefix(strings.TrimSuffix(h.String(), "</p>\n"), "<p>")
					buf.WriteString(fmt.Sprintf(`<div id="%s-content" class="tab-content">
<section class="section" id="%s">
  <div class="container">
    <h2 class="subtitle has-text-weight-semibold">%s</h2>
  </div>
</section>`, t.ID, t.ID, titleText))
				} else {
					// Regular section rendering for non-tab sections
					titleText := strings.TrimPrefix(strings.TrimSuffix(h.String(), "</p>\n"), "<p>")
					buf.WriteString(fmt.Sprintf(`<section class="section" id="%s">
  <div class="container">
    <h2 class="subtitle has-text-weight-semibold">%s</h2>
  </div>
</section>`, t.ID, titleText))
				}
			}
		case Field:
			v := c.resolvePath(t.Path, vars)
			buf.WriteString(fmt.Sprintf("<p>%s</p>", htmlEscape(fmt.Sprint(v))))
		case Loop:
			if err := c.renderLoop(t, vars, buf); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *context) renderLoop(l Loop, vars map[string]interface{}, buf *strings.Builder) error {
	src := c.resolvePath(l.Source, vars)
	switch arr := src.(type) {
	case []interface{}:
		items := append([]interface{}{}, arr...)
		
		// Check if this is an apps loop and filter for software items only
		// Based on SPEC.md: "Renderer should use apps.results filtered to items where kind == 'software'"
		// Only apply filtering if we detect this is apps data (has wrapperType or kind fields)
		isAppsData := false
		if len(items) > 0 {
			if itemMap, ok := items[0].(map[string]interface{}); ok {
				if _, hasKind := itemMap["kind"]; hasKind {
					isAppsData = true
				} else if _, hasWrapper := itemMap["wrapperType"]; hasWrapper {
					isAppsData = true
				}
			}
		}
		
		if isAppsData {
			filteredItems := make([]interface{}, 0, len(items))
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Check if this is a software app (not an artist entry)
					if kind, hasKind := itemMap["kind"]; hasKind && kind == "software" {
						filteredItems = append(filteredItems, item)
					} else if wrapperType, hasWrapper := itemMap["wrapperType"]; hasWrapper && wrapperType == "software" {
						// Also check wrapperType for additional safety
						filteredItems = append(filteredItems, item)
					}
				}
			}
			items = filteredItems
		}
		
		SortSlice(items, l.Sort)
		
		// Check loop type by examining the first item
		isAppsLoop := false
		isThingsLoop := false
		isReposLoop := false
		if len(items) > 0 {
			if item, ok := items[0].(map[string]interface{}); ok {
				// Check if this is an apps loop by looking for app-specific fields
				if _, hasTrackName := item["trackName"]; hasTrackName {
					if _, hasTrackViewUrl := item["trackViewUrl"]; hasTrackViewUrl {
						if _, hasGenres := item["genres"]; hasGenres {
							isAppsLoop = true
						}
					}
				}
				// Check if this is a things loop by looking for thing-specific fields
				if _, hasTitle := item["title"]; hasTitle {
					if _, hasUrl := item["url"]; hasUrl {
						if _, hasCategory := item["category"]; hasCategory {
							if _, hasDatePublished := item["date_published"]; hasDatePublished {
								isThingsLoop = true
							}
						}
					}
				}
				// Check if this is a repos loop by looking for repo-specific fields
				if _, hasName := item["name"]; hasName {
					if _, hasHtmlUrl := item["html_url"]; hasHtmlUrl {
						if _, hasStargazersCount := item["stargazers_count"]; hasStargazersCount {
							if _, hasUpdatedAt := item["updated_at"]; hasUpdatedAt {
								isReposLoop = true
							}
						}
					}
				}
			}
		}
		
		if isAppsLoop {
			buf.WriteString(`<section class="section">`)
			buf.WriteString(`<div class="container">`)
			buf.WriteString(`<div class="grid is-col-min-16">`)
			for _, it := range items {
				nv := map[string]interface{}{}
				if len(l.Vars) > 0 {
					nv[l.Vars[0]] = it
				}
				buf.WriteString(`<div class="cell">`)
				if err := c.renderAppCard(l.Body, merge(vars, nv), buf); err != nil {
					return err
				}
				buf.WriteString(`</div>`)
			}
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			buf.WriteString(`</section>`)
			buf.WriteString(`</div>`) // Close tab-content div for apps
		} else if isThingsLoop {
			// Group things by category and collect unique categories
			categoryMap := make(map[string][]interface{})
			categoryOrder := []string{}
			
			for _, it := range items {
				if itemMap, ok := it.(map[string]interface{}); ok {
					category := fmt.Sprintf("%v", itemMap["category"])
					if _, exists := categoryMap[category]; !exists {
						categoryOrder = append(categoryOrder, category)
						categoryMap[category] = []interface{}{}
					}
					categoryMap[category] = append(categoryMap[category], it)
				}
			}
			
			// Generate category filter menu
			buf.WriteString(`<div class="category-filter-wrapper">`)
			buf.WriteString(`<div class="category-filter-scroll">`)
			buf.WriteString(`<div class="level is-mobile category-filter-level">`)
			buf.WriteString(`<div class="level-item">`)
			buf.WriteString(`<div class="buttons has-addons category-filter-buttons">`)
			buf.WriteString(`<button class="button is-info is-selected" onclick="filterThings('all')">All</button>`)
			
			for _, category := range categoryOrder {
				capitalizedCategory := category
				if len(category) > 0 {
					capitalizedCategory = strings.ToUpper(category[:1]) + category[1:]
				}
				buf.WriteString(`<button class="button" onclick="filterThings('` + category + `')">`)
				buf.WriteString(htmlEscape(capitalizedCategory))
				buf.WriteString(`</button>`)
			}
			
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			
			// Render all things in one section with category data attributes
			buf.WriteString(`<section class="section">`)
			buf.WriteString(`<div class="container">`)
			buf.WriteString(`<div class="grid is-col-min-16" id="things-grid">`)
			
			for _, category := range categoryOrder {
				for _, it := range categoryMap[category] {
					nv := map[string]interface{}{}
					if len(l.Vars) > 0 {
						nv[l.Vars[0]] = it
					}
					buf.WriteString(`<div class="cell thing-item" data-category="` + category + `">`)
					if err := c.renderThingCard(l.Body, merge(vars, nv), buf); err != nil {
						return err
					}
					buf.WriteString(`</div>`)
				}
			}
			
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			buf.WriteString(`</section>`)
			buf.WriteString(`</div>`) // Close tab-content div for things
		} else if isReposLoop {
			buf.WriteString(`<section class="section">`)
			buf.WriteString(`<div class="container">`)
			buf.WriteString(`<div class="grid is-col-min-16">`)
			for _, it := range items {
				nv := map[string]interface{}{}
				if len(l.Vars) > 0 {
					nv[l.Vars[0]] = it
				}
				buf.WriteString(`<div class="cell">`)
				if err := c.renderRepoCard(l.Body, merge(vars, nv), buf); err != nil {
					return err
				}
				buf.WriteString(`</div>`)
			}
			buf.WriteString(`</div>`)
			buf.WriteString(`</div>`)
			buf.WriteString(`</section>`)
			buf.WriteString(`</div>`) // Close tab-content div for repos
		} else {
			buf.WriteString(`<section class="section">`)
			for _, it := range items {
				nv := map[string]interface{}{}
				if len(l.Vars) > 0 {
					nv[l.Vars[0]] = it
				}
				buf.WriteString(`<div class="box">`)
				if err := c.renderNodes(l.Body, merge(vars, nv), buf); err != nil {
					return err
				}
				buf.WriteString(`</div>`)
			}
			buf.WriteString(`</section>`)
		}
	case map[string]interface{}:
		keys := make([]interface{}, 0, len(arr))
		for k, v := range arr {
			keys = append(keys, map[string]interface{}{"key": k, "value": v})
		}
		// for map, sort by provided sort keys applied on struct with key/value
		SortSlice(keys, l.Sort)
		buf.WriteString(`<section class="section">`)
		for _, kv := range keys {
			m := kv.(map[string]interface{})
			nv := map[string]interface{}{}
			if len(l.Vars) == 2 {
				nv[l.Vars[0]] = m["key"]
				nv[l.Vars[1]] = m["value"]
			} else if len(l.Vars) == 1 {
				nv[l.Vars[0]] = m["value"]
			}
			buf.WriteString(`<div class="box">`)
			if err := c.renderNodes(l.Body, merge(vars, nv), buf); err != nil {
				return err
			}
			buf.WriteString(`</div>`)
		}
		buf.WriteString(`</section>`)
	}
	return nil
}

func merge(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func (c *context) resolvePath(path string, vars map[string]interface{}) interface{} {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}
	head := parts[0]
	var obj interface{}
	if v, ok := vars[head]; ok {
		obj = v
	} else if v, ok := c.bindings[head]; ok {
		obj = v
	} else if lb, ok := c.lazy[head]; ok {
		obj = c.resolveLazy(lb, head, vars)
	}
	if obj == nil {
		return nil
	}
	if len(parts) == 1 {
		return obj
	}
	return Resolve(obj, strings.Join(parts[1:], "."))
}

var tmplRe = regexp.MustCompile(`\{([^}]+)\}`)

func (c *context) resolveLazy(lb *lazyBinding, name string, vars map[string]interface{}) interface{} {
	// need key from template
	url := lb.Template
	var key string
	matches := tmplRe.FindAllStringSubmatch(lb.Template, -1)
	for _, m := range matches {
		expr := m[1]
		parts := strings.SplitN(expr, ".", 2)
		if len(parts) != 2 {
			continue
		}
		v := vars[parts[0]]
		if rest := parts[1]; rest != "" {
			v = Resolve(v, rest)
		}
		s := fmt.Sprintf("%v", v)
		url = strings.ReplaceAll(url, m[0], s)
		if key == "" {
			key = s
		}
	}
	if key == "" {
		key = url
	}
	if val, ok := lb.Data[key]; ok {
		return val
	}
	if lb.Fetched[key] {
		return nil
	}
	lb.Fetched[key] = true
	body, err := c.fetcher.Fetch(lb.Target, url)
	if err != nil {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil
	}
	lb.Data[key] = v
	return v
}

func htmlEscape(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func (c *context) renderAppCard(nodes []Node, vars map[string]interface{}, buf *strings.Builder) error {
	// Extract app data from variables
	app := vars["app"]
	if app == nil {
		return c.renderNodes(nodes, vars, buf)
	}
	
	appMap, ok := app.(map[string]interface{})
	if !ok {
		return c.renderNodes(nodes, vars, buf)
	}
	
	trackName := fmt.Sprintf("%v", appMap["trackName"])
	trackViewUrl := fmt.Sprintf("%v", appMap["trackViewUrl"])
	description := fmt.Sprintf("%v", appMap["description"])
	artworkUrl100 := fmt.Sprintf("%v", appMap["artworkUrl100"])
	genres := appMap["genres"]
	
	// Truncate description to around 200 characters with ellipsis, breaking on word boundaries
	truncatedDesc := description
	if len(description) > 200 {
		// Find the last space within the first 200 characters
		cutoff := 200
		lastSpace := -1
		for i := cutoff - 1; i >= 0; i-- {
			if description[i] == ' ' {
				lastSpace = i
				break
			}
		}
		
		// If we found a space within reasonable distance, use it; otherwise fallback to character cutoff
		if lastSpace > cutoff - 50 { // Don't go back more than 50 characters to find a space
			truncatedDesc = description[:lastSpace] + "..."
		} else {
			truncatedDesc = description[:200] + "..."
		}
	}
	
	buf.WriteString(`<div class="card">
  <div class="card-content">
    <div class="content">`)
	
	// Add app icon and title in a flex layout
	if artworkUrl100 != "" && artworkUrl100 != "<no value>" {
		buf.WriteString(`
      <div class="level is-mobile" style="margin-bottom: 1rem;">
        <div class="level-left">
          <div class="level-item">
            <figure class="image is-48x48" style="margin-right: 0.75rem;">
              <img src="`)
		buf.WriteString(htmlEscape(artworkUrl100))
		buf.WriteString(`" alt="`)
		buf.WriteString(htmlEscape(trackName))
		buf.WriteString(` icon" style="border-radius: 12px;">
            </figure>
          </div>
          <div class="level-item">
            <h3 class="title is-5" style="margin-bottom: 0;">
              <a href="`)
		buf.WriteString(htmlEscape(trackViewUrl))
		buf.WriteString(`" target="_blank" rel="noopener noreferrer">`)
		buf.WriteString(htmlEscape(trackName))
		buf.WriteString(`</a>
            </h3>
          </div>
        </div>
      </div>`)
	} else {
		buf.WriteString(`
      <h3 class="title is-5">
        <a href="`)
		buf.WriteString(htmlEscape(trackViewUrl))
		buf.WriteString(`" target="_blank" rel="noopener noreferrer">`)
		buf.WriteString(htmlEscape(trackName))
		buf.WriteString(`</a>
      </h3>`)
	}
	
	buf.WriteString(`
      <p>`)
	buf.WriteString(htmlEscape(truncatedDesc))
	buf.WriteString(`</p>
    </div>
  </div>`)
	
	// Add categories section if genres exist
	if genres != nil {
		if genreSlice, ok := genres.([]interface{}); ok && len(genreSlice) > 0 {
			buf.WriteString(`
  <div class="card-footer">
    <div class="card-footer-item">
      <div class="tags">`)
			for _, genre := range genreSlice {
				genreStr := fmt.Sprintf("%v", genre)
				buf.WriteString(`
        <span class="tag is-info is-light">`)
				buf.WriteString(htmlEscape(genreStr))
				buf.WriteString(`</span>`)
			}
			buf.WriteString(`
      </div>
    </div>
  </div>`)
		}
	}
	
	buf.WriteString(`
</div>`)
	
	return nil
}

func (c *context) renderThingCard(nodes []Node, vars map[string]interface{}, buf *strings.Builder) error {
	// Extract thing data from variables
	thing := vars["thing"]
	if thing == nil {
		return c.renderNodes(nodes, vars, buf)
	}
	
	thingMap, ok := thing.(map[string]interface{})
	if !ok {
		return c.renderNodes(nodes, vars, buf)
	}
	
	title := fmt.Sprintf("%v", thingMap["title"])
	url := fmt.Sprintf("%v", thingMap["url"])
	description := fmt.Sprintf("%v", thingMap["description"])
	category := fmt.Sprintf("%v", thingMap["category"])
	
	// Use full description without truncation for things
	truncatedDesc := description
	
	buf.WriteString(`<div class="card">
  <div class="card-content">
    <div class="content">
      <h3 class="title is-5">
        <a href="`)
	buf.WriteString(htmlEscape(url))
	buf.WriteString(`" target="_blank" rel="noopener noreferrer">`)
	buf.WriteString(htmlEscape(title))
	buf.WriteString(`</a>
      </h3>
      <p>`)
	buf.WriteString(htmlEscape(truncatedDesc))
	buf.WriteString(`</p>
    </div>
  </div>`)
	
	// Add category section
	if category != "" {
		buf.WriteString(`
  <div class="card-footer">
    <div class="card-footer-item">
      <div class="tags">
        <span class="tag is-info is-light">`)
		buf.WriteString(htmlEscape(category))
		buf.WriteString(`</span>
      </div>
    </div>
  </div>`)
	}
	
	buf.WriteString(`
</div>`)
	
	return nil
}

func (c *context) renderRepoCard(nodes []Node, vars map[string]interface{}, buf *strings.Builder) error {
	// Extract repo data from variables
	repo := vars["repo"]
	if repo == nil {
		return c.renderNodes(nodes, vars, buf)
	}
	
	repoMap, ok := repo.(map[string]interface{})
	if !ok {
		return c.renderNodes(nodes, vars, buf)
	}
	
	name := fmt.Sprintf("%v", repoMap["name"])
	htmlUrl := fmt.Sprintf("%v", repoMap["html_url"])
	description := fmt.Sprintf("%v", repoMap["description"])
	stargazersCount := fmt.Sprintf("%v", repoMap["stargazers_count"])
	updatedAt := fmt.Sprintf("%v", repoMap["updated_at"])
	language := fmt.Sprintf("%v", repoMap["language"])
	
	// Parse and format the date
	var formattedDate string
	if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		formattedDate = t.Format("January 2, 2006")
	} else {
		formattedDate = updatedAt
	}
	
	buf.WriteString(`<div class="card">
  <div class="card-content">
    <div class="content">
      <h3 class="title is-5">
        <a href="`)
	buf.WriteString(htmlEscape(htmlUrl))
	buf.WriteString(`" target="_blank" rel="noopener noreferrer">`)
	buf.WriteString(htmlEscape(name))
	buf.WriteString(`</a>
      </h3>
      <p>`)
	buf.WriteString(htmlEscape(description))
	buf.WriteString(`</p>
      <div class="level">
        <div class="level-left">
          <div class="level-item">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" style="margin-right: 4px;">
              <path d="M8 .2l4.9 15.2L0 6h16L3.1 15.4z"/>
            </svg>
            <span>`)
	buf.WriteString(htmlEscape(stargazersCount))
	buf.WriteString(`</span>
          </div>`)
	
	// Add language if it exists and is not "<no value>"
	if language != "" && language != "<no value>" {
		buf.WriteString(`
          <div class="level-item">
            <span class="tag is-info is-light">`)
		buf.WriteString(htmlEscape(language))
		buf.WriteString(`</span>
          </div>`)
	}
	
	buf.WriteString(`
        </div>
      </div>
    </div>
  </div>
  <footer class="card-footer">
    <p class="card-footer-item has-text-grey-light">
      Last updated `)
	buf.WriteString(htmlEscape(formattedDate))
	buf.WriteString(`
    </p>
  </footer>
</div>`)
	
	return nil
}
