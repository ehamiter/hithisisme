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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ehamiter/hithisisme/sitegen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "expected subcommand")
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "render":
		renderCmd(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %s\n", cmd)
		os.Exit(1)
	}
}

func renderCmd(args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	input := fs.String("input", "index.hi", "input .hi file")
	out := fs.String("out", "public/index.html", "output HTML file")
	dataDir := fs.String("data-dir", "data", "data directory")
	layout := fs.String("layout", "templates/layout.html", "layout HTML file")
	fs.Parse(args)

	err := sitegen.Render(sitegen.RenderOptions{
		Input:   *input,
		Out:     *out,
		DataDir: *dataDir,
		Layout:  *layout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
