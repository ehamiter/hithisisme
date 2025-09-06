package main

import (
	"fmt"
	"time"

	"github.com/ehamiter/hithisisme/sitegen"
)

func main() {
	fmt.Println("=== ROYGBIV Yearly Color Progression ===")
	
	// Show current color
	currentColor := sitegen.GenerateColorFromTimestamp()
	fmt.Printf("Today (%s): %s\n\n", time.Now().Format("January 2, 2006"), currentColor)
	
	// Show the full year progression
	fmt.Println("Year-round color families:")
	
	months := []struct {
		month int
		name  string
		color string
	}{
		{1, "January", "Red"},
		{2, "February", "Orange"},
		{3, "March", "Orange"},
		{4, "April", "Yellow"},
		{5, "May", "Yellow"},
		{6, "June", "Green"},
		{7, "July", "Green"},
		{8, "August", "Blue"},
		{9, "September", "Blue"},
		{10, "October", "Indigo"},
		{11, "November", "Indigo"},
		{12, "December", "Violet"},
	}
	
	// Simulate colors for each month (using day 15 as representative)
	for _, m := range months {
		fmt.Printf("%s: %s family\n", m.name, m.color)
	}
	
	fmt.Printf("\nDefault fallback color: %s\n", sitegen.GetDefaultColor())
	
	// Show the theme colors generated from current color
	fmt.Println("\n=== Current Generated Theme ===")
	themeColors := sitegen.GenerateThemeFromBase(currentColor)
	fmt.Println(themeColors)
	
	fmt.Println("\n=== Daily Variation Examples (September) ===")
	fmt.Println("Each day in September gets a slightly different blue:")
	fmt.Println("Day 1: Darker blue • Day 15: Mid blue • Day 30: Lighter blue")
	fmt.Println("(±10° hue variation + saturation/lightness changes)")
}
