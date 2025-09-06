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
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var bindRe = regexp.MustCompile(`^(!?)([A-Za-z0-9_]+)\s*=\s*(.+)$`)
var eagerRe = regexp.MustCompile(`^([^<\s]+)\s*<<\s*(.+)$`)

// Parse reads a .hi file into bindings and body nodes.
func Parse(r io.Reader) ([]Binding, []Node, error) {
	scanner := bufio.NewScanner(r)
	var bindings []Binding
	var bodyLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if len(strings.TrimSpace(line)) == 0 {
			bodyLines = append(bodyLines, line)
			continue
		}
		if m := bindRe.FindStringSubmatch(line); m != nil {
			lazy := m[1] == "!"
			rest := m[3]
			var b Binding
			b.Name = m[2]
			if m2 := eagerRe.FindStringSubmatch(rest); m2 != nil {
				b.Target = strings.TrimSpace(m2[1])
				b.URL = strings.TrimSpace(m2[2])
				b.Lazy = lazy
			} else {
				b.Target = strings.TrimSpace(rest)
				b.Manual = true
			}
			bindings = append(bindings, b)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	nodes, err := parseBody(bodyLines, 0, 0)
	if err != nil {
		return nil, nil, err
	}
	return bindings, nodes, nil
}

// parseBody parses body lines recursively expecting indent spaces.
func parseBody(lines []string, start, indent int) ([]Node, error) {
	var nodes []Node
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		curIndent := countIndent(line)
		if curIndent < indent {
			return nodes, fmt.Errorf("unexpected dedent")
		}
		if curIndent > indent {
			// handled by recursion from parent
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "{"), "}")
			parts := strings.SplitN(inner, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid section: %s", line)
			}
			nodes = append(nodes, Section{ID: strings.TrimSpace(parts[0]), Text: strings.TrimSpace(parts[1])})
			continue
		}
		if strings.HasPrefix(trimmed, "[for ") {
			loop, used, err := parseLoop(lines[i:], indent)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, loop)
			i += used - 1
			continue
		}
		// field line
		nodes = append(nodes, Field{Path: trimmed})
	}
	return nodes, nil
}

func parseLoop(lines []string, indent int) (Loop, int, error) {
	header := strings.TrimSpace(lines[0])
	end := strings.Index(header, "]")
	if end < 0 {
		return Loop{}, 0, fmt.Errorf("missing ] in loop header")
	}
	inner := strings.TrimSpace(header[1:end]) // remove []
	if !strings.HasPrefix(inner, "for ") {
		return Loop{}, 0, fmt.Errorf("invalid loop header")
	}
	inner = strings.TrimPrefix(inner, "for ")
	parts := strings.SplitN(inner, ":", 2)
	left := parts[0]
	sortSpec := ""
	if len(parts) == 2 {
		sortSpec = strings.TrimSpace(parts[1])
		// Handle prefix ^ for ascending sort on all fields
		if strings.HasPrefix(sortSpec, "^") {
			sortSpec = strings.TrimSpace(strings.TrimPrefix(sortSpec, "^"))
			// Add ^ suffix to each field for ascending sort
			fields := strings.Split(sortSpec, ",")
			for i, field := range fields {
				fields[i] = strings.TrimSpace(field) + "^"
			}
			sortSpec = strings.Join(fields, ", ")
		}
	}
	inParts := strings.Split(left, " in ")
	if len(inParts) != 2 {
		return Loop{}, 0, fmt.Errorf("invalid loop header")
	}
	vars := strings.Split(strings.TrimSpace(inParts[0]), ",")
	for i := range vars {
		vars[i] = strings.TrimSpace(vars[i])
	}
	source := strings.TrimSpace(inParts[1])
	loop := Loop{Vars: vars, Source: source, Sort: ParseSort(sortSpec)}
	// parse body
	bodyIndent := indent + 2
	var bodyLines []string
	used := 1
	for ; used < len(lines); used++ {
		line := lines[used]
		if strings.TrimSpace(line) == "" {
			continue
		}
		ind := countIndent(line)
		if ind < bodyIndent {
			break
		}
		bodyLines = append(bodyLines, line)
	}
	nodes, err := parseBody(bodyLines, 0, bodyIndent)
	if err != nil {
		return Loop{}, 0, err
	}
	loop.Body = nodes
	return loop, used, nil
}

func countIndent(s string) int {
	c := 0
	for _, ch := range s {
		if ch == ' ' {
			c++
		} else {
			break
		}
	}
	return c
}

// ParseSort parses sort specification.
func ParseSort(spec string) []SortKey {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}
	parts := strings.Split(spec, ",")
	keys := make([]SortKey, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		asc := false
		if strings.HasSuffix(p, "^") {
			asc = true
			p = strings.TrimSuffix(p, "^")
		}
		keys = append(keys, SortKey{Path: p, Asc: asc})
	}
	return keys
}
