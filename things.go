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
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Thing struct {
	Category      string `json:"category"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	Description   string `json:"description"`
	DatePublished string `json:"date_published"`
}

func main() {
	fmt.Println("Let's add a new thing!")
	
	scanner := bufio.NewScanner(os.Stdin)
	
	// Get Category
	fmt.Print("Category: ")
	scanner.Scan()
	category := strings.TrimSpace(scanner.Text())
	
	// Get Title
	fmt.Print("Title: ")
	scanner.Scan()
	title := strings.TrimSpace(scanner.Text())
	
	// Get URL
	fmt.Print("URL: ")
	scanner.Scan()
	url := strings.TrimSpace(scanner.Text())
	
	// Get Description
	fmt.Print("Description: ")
	scanner.Scan()
	description := strings.TrimSpace(scanner.Text())
	
	// Create new thing with current date
	newThing := Thing{
		Category:      category,
		Title:         title,
		URL:           url,
		Description:   description,
		DatePublished: time.Now().Format("2006-01-02"),
	}
	
	// Read existing things from JSON file
	filePath := "data/things.json"
	var things []Thing
	
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	// Unmarshal existing data
	err = json.Unmarshal(data, &things)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}
	
	// Add new thing to the slice
	things = append(things, newThing)
	
	// Marshal back to JSON with pretty formatting
	updatedData, err := json.MarshalIndent(things, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	
	// Write back to file
	err = ioutil.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
	
	fmt.Println("\nNew thing added successfully!")
	fmt.Printf("Category: %s\n", newThing.Category)
	fmt.Printf("Title: %s\n", newThing.Title)
	fmt.Printf("URL: %s\n", newThing.URL)
	fmt.Printf("Description: %s\n", newThing.Description)
	fmt.Printf("Date Published: %s\n", newThing.DatePublished)
}
