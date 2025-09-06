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
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HSL represents a color in HSL (Hue, Saturation, Lightness) color space
type HSL struct {
	H, S, L float64
}

// HexToHSL converts hex color to HSL
func HexToHSL(hex string) HSL {
	if hex[0] == '#' {
		hex = hex[1:]
	}
	
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)
	
	rF := float64(r) / 255.0
	gF := float64(g) / 255.0
	bF := float64(b) / 255.0
	
	max := math.Max(math.Max(rF, gF), bF)
	min := math.Min(math.Min(rF, gF), bF)
	
	h, s, l := 0.0, 0.0, (max+min)/2.0
	
	if max == min {
		h, s = 0.0, 0.0 // achromatic
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}
		
		switch max {
		case rF:
			h = (gF-bF)/d + func() float64 { if gF < bF { return 6 } else { return 0 } }()
		case gF:
			h = (bF-rF)/d + 2
		case bF:
			h = (rF-gF)/d + 4
		}
		h /= 6
	}
	
	return HSL{H: h * 360, S: s * 100, L: l * 100}
}

// HSLToHex converts HSL back to hex
func HSLToHex(h, s, l float64) string {
	s /= 100
	l /= 100
	
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	
	var r, g, b float64
	
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	
	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255
	
	return fmt.Sprintf("#%02x%02x%02x", int(r), int(g), int(b))
}

// GenerateThemeFromBase creates CSS custom properties from a base color
func GenerateThemeFromBase(baseHex string) string {
	hsl := HexToHSL(baseHex)
	
	return fmt.Sprintf(`  /* Base color - generated from %s */
  --base-color: %s;
  --base-hue: %.0fdeg;
  --base-saturation: %.0f%%;
  --base-lightness: %.0f%%;`, baseHex, baseHex, hsl.H, hsl.S, hsl.L)
}

// GenerateColorFromTimestamp creates a color based on timestamp
// Progresses through ROYGBIV spectrum over the year, perfect for daily builds
func GenerateColorFromTimestamp() string {
	now := time.Now()
	month := int(now.Month())
	dayOfMonth := now.Day()
	
	// Map months to ROYGBIV color ranges
	var baseHue float64
	var colorName string
	
	switch month {
	case 1: // January - Red
		baseHue = 0
		colorName = "Red"
	case 2, 3: // February, March - Orange  
		baseHue = 25
		colorName = "Orange"
	case 4, 5: // April, May - Yellow
		baseHue = 50
		colorName = "Yellow"
	case 6, 7: // June, July - Green
		baseHue = 120
		colorName = "Green"
	case 8, 9: // August, September - Blue
		baseHue = 240
		colorName = "Blue"
	case 10, 11: // October, November - Indigo
		baseHue = 270
		colorName = "Indigo"
	case 12: // December - Violet
		baseHue = 300
		colorName = "Violet"
	}
	
	// Add subtle daily variation within the color family (±10 degrees)
	dailyVariation := (float64(dayOfMonth-1) / 30.0) * 20.0 - 10.0 // -10 to +10
	hue := math.Mod(baseHue + dailyVariation, 360.0)
	if hue < 0 {
		hue += 360
	}
	
	// Seasonal saturation: higher in summer, lower in winter for natural feel
	// Range: 45-70% to keep colors rich but not overwhelming
	seasonalSat := 45.0
	if month >= 4 && month <= 9 { // Spring/Summer: more vibrant
		seasonalSat = 60.0 + float64(dayOfMonth%10) // 60-69%
	} else { // Fall/Winter: more muted
		seasonalSat = 45.0 + float64(dayOfMonth%10) // 45-54%
	}
	
	// Lightness varies by day for subtle daily difference while staying dark
	// Range: 18-30% to ensure good contrast for both light and dark themes
	lightness := 18.0 + float64(dayOfMonth%12) // 18-29%
	
	// Convert HSL to hex
	result := HSLToHex(hue, seasonalSat, lightness)
	
	// Log which color family we're in (helpful for debugging/interest)
	fmt.Printf("Theme: %s family (month %d, day %d) ", colorName, month, dayOfMonth)
	
	return result
}

// GetDefaultColor returns the current default color for fallback
func GetDefaultColor() string {
	return "#2d5016"
}

// generateCSS creates the dynamic CSS file with timestamp-based colors
func generateCSS(outputDir string) error {
	// Generate color from timestamp, with optional override
	var baseColor string
	if envColor := os.Getenv("THEME_COLOR"); envColor != "" {
		baseColor = envColor
		fmt.Printf("Using override theme color from THEME_COLOR: %s\n", baseColor)
	} else {
		baseColor = GenerateColorFromTimestamp()
	}
	
	// Read CSS template
	templatePath := "templates/style.css.template"
	template, err := os.ReadFile(templatePath)
	if err != nil {
		// Fallback: if template doesn't exist, use default color and log
		fmt.Printf("Warning: CSS template not found at %s, using default color\n", templatePath)
		baseColor = GetDefaultColor()
		
		// Try to read the existing CSS and update it (fallback mode)
		existingCSS := filepath.Join(outputDir, "style.css")
		if existing, err := os.ReadFile(existingCSS); err == nil {
			template = existing
		} else {
			return fmt.Errorf("no CSS template or existing CSS file found")
		}
	}
	
	// Generate theme colors from base
	themeColors := GenerateThemeFromBase(baseColor)
	
	// Replace placeholder with actual colors
	cssContent := strings.Replace(string(template), "{{THEME_COLORS}}", themeColors, 1)
	
	// Write CSS file
	cssPath := filepath.Join(outputDir, "style.css")
	if err := os.WriteFile(cssPath, []byte(cssContent), 0o644); err != nil {
		return fmt.Errorf("failed to write CSS file: %w", err)
	}
	
	fmt.Printf("Generated CSS with theme color: %s\n", baseColor)
	return nil
}
