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
