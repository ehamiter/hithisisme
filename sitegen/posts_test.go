package sitegen

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadPosts(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	pattern := filepath.Join(filepath.Dir(file), "..", "posts", "*.md")
	posts, err := LoadPosts(pattern)
	if err != nil {
		t.Fatalf("load posts: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("expected posts")
	}
	found := false
	for _, p := range posts {
		m := p.(map[string]interface{})
		if m["date"] == "2025-08-31" {
			found = true
			if m["title"].(string) != "lumberjack suburbia" {
				t.Fatalf("title mismatch: %v", m["title"])
			}
			if m["url"].(string) != "/posts/2025-08-31/" {
				t.Fatalf("url mismatch: %v", m["url"])
			}
		}
	}
	if !found {
		t.Fatalf("post not found")
	}
}
