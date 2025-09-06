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

func TestResolveNested(t *testing.T) {
	data := map[string]interface{}{
		"repo": map[string]interface{}{
			"languages": []interface{}{
				map[string]interface{}{"name": "Go"},
			},
		},
	}
	if got := Resolve(data, "repo.languages.0.name"); got != "Go" {
		t.Fatalf("expected Go, got %v", got)
	}
	if Resolve(data, "repo.missing") != nil {
		t.Fatalf("expected nil for missing field")
	}
}
