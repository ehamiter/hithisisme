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
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	// Parse command line flags
	serve := flag.Bool("serve", false, "serve the site on http://localhost:8000 after building and rendering")
	flag.Parse()

	// Build the sitegen binary
	log.Println("Building sitegen...")
	buildCmd := exec.Command("go", "build", "-o", "hi", "./cmd/sitegen")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		log.Fatal("Failed to build sitegen:", err)
	}

	// Render the site
	log.Println("Rendering site...")
	renderCmd := exec.Command("./hi", "render")
	renderCmd.Stdout = os.Stdout
	renderCmd.Stderr = os.Stderr
	if err := renderCmd.Run(); err != nil {
		log.Fatal("Failed to render site:", err)
	}

	// Only serve if --serve flag is provided
	if *serve {
		// Serve files from ./public directory
		fs := http.FileServer(http.Dir("public"))

		// Mount it at root
		http.Handle("/", fs)

		// Listen on port 8000
		log.Println("Serving on http://localhost:8000")
		err := http.ListenAndServe(":8000", nil)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Build and render complete. Use --serve to start the development server.")
	}
}
