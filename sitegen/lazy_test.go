package sitegen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLazyFetchCaching(t *testing.T) {
	dir := t.TempDir()
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"value":1}`)
	}))
	defer srv.Close()

	lb := &lazyBinding{Target: "lang.json", Template: srv.URL + "/{repo.name}", Data: map[string]interface{}{}, Fetched: map[string]bool{}}
	ctx := &context{lazy: map[string]*lazyBinding{"languages": lb}, fetcher: NewFetcher(dir)}
	vars := map[string]interface{}{"repo": map[string]interface{}{"name": "foo"}}
	if v := ctx.resolvePath("languages", vars); v == nil {
		t.Fatalf("expected value")
	}
	if hits != 1 {
		t.Fatalf("expected 1 hit, got %d", hits)
	}
	if v := ctx.resolvePath("languages", vars); v == nil {
		t.Fatalf("expected cached value")
	}
	if hits != 1 {
		t.Fatalf("expected no additional hit")
	}
	// save to disk
	b, _ := json.Marshal(lb.Data)
	os.WriteFile(filepath.Join(dir, "lang.json"), b, 0o644)
	// new context
	lb2 := &lazyBinding{Target: "lang.json", Template: srv.URL + "/{repo.name}", Data: map[string]interface{}{}, Fetched: map[string]bool{}}
	data, _ := os.ReadFile(filepath.Join(dir, "lang.json"))
	json.Unmarshal(data, &lb2.Data)
	ctx2 := &context{lazy: map[string]*lazyBinding{"languages": lb2}, fetcher: NewFetcher(dir)}
	if v := ctx2.resolvePath("languages", vars); v == nil {
		t.Fatalf("expected cached value in second run")
	}
	if hits != 1 {
		t.Fatalf("expected cache reuse across runs")
	}
}
