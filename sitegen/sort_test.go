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

import "testing"

func TestParseSort(t *testing.T) {
	keys := ParseSort("stargazers_count, updated_at^")
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Path != "stargazers_count" || keys[0].Asc {
		t.Fatalf("first key wrong: %+v", keys[0])
	}
	if keys[1].Path != "updated_at" || !keys[1].Asc {
		t.Fatalf("second key wrong: %+v", keys[1])
	}
}

func TestSortSlice(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"a": 1, "b": 2},
		map[string]interface{}{"a": 2, "b": 1},
		map[string]interface{}{"b": 3}, // missing a
	}
	SortSlice(items, ParseSort("a^"))
	if Resolve(items[0], "a").(int) != 1 {
		t.Fatalf("expected first item a=1")
	}
	if Resolve(items[2], "a") != nil {
		t.Fatalf("expected last item missing a")
	}
}
