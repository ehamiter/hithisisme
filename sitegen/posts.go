package sitegen

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

// LoadPosts reads markdown posts matching pattern and returns post objects.
func LoadPosts(pattern string) ([]interface{}, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	posts := make([]interface{}, 0, len(matches))
	for _, f := range matches {
		if filepath.Ext(f) != ".md" {
			continue
		}
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		content := string(b)
		lines := strings.Split(content, "\n")
		var title string
		var preview string
		for _, l := range lines {
			if strings.HasPrefix(l, "# ") && title == "" {
				title = strings.TrimSpace(l[2:])
				continue
			}
			if preview == "" {
				t := strings.TrimSpace(l)
				if t != "" && !strings.HasPrefix(t, "#") {
					preview = stripMarkdown(t)
				}
			}
			if title != "" && preview != "" {
				break
			}
		}
		if title == "" {
			base := filepath.Base(f)
			title = strings.TrimSuffix(base, filepath.Ext(base))
		}
		if len(preview) > 240 {
			preview = preview[:240] + "…"
		}
		date := filepath.Base(f)[:10]
		post := map[string]interface{}{
			"title":   title,
			"preview": preview,
			"date":    date,
			"url":     "/posts/" + date + "/",
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func stripMarkdown(s string) string {
	// very naive stripping
	md := goldmark.New()
	var buf strings.Builder
	if err := md.Convert([]byte(s), &buf); err == nil {
		out := buf.String()
		out = strings.ReplaceAll(out, "<p>", "")
		out = strings.ReplaceAll(out, "</p>", "")
		return strings.TrimSpace(out)
	}
	return s
}
