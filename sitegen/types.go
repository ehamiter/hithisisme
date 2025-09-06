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

// SortKey represents a field and sort direction.
type SortKey struct {
	Path string
	Asc  bool
}

// Binding defines a variable binding from header.
type Binding struct {
	Name   string
	Target string
	URL    string
	Lazy   bool
	Manual bool
}

// Section is a simple markdown section.
type Section struct {
	ID   string
	Text string
}

// Loop represents a for-loop block.
type Loop struct {
	Vars   []string
	Source string
	Sort   []SortKey
	Body   []Node
}

// Node is a body node.
type Node interface{}

// Field line in loop body.
type Field struct {
	Path string
}
