package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/example/hithisisme/sitegen"
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
