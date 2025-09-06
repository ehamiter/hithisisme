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
	"strconv"
	"strings"
)

// Resolve returns the value at path from obj using dotted notation.
func Resolve(obj interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	cur := obj
	for _, p := range parts {
		switch c := cur.(type) {
		case map[string]interface{}:
			cur = c[p]
		case []interface{}:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil
			}
			cur = c[idx]
		default:
			return nil
		}
	}
	return cur
}

// GetString gets a string value or empty string.
func GetString(obj interface{}) string {
	if s, ok := obj.(string); ok {
		return s
	}
	return ""
}
