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
// Uses a combination of time-based factors to create interesting, seasonal colors
func GenerateColorFromTimestamp() string {
	now := time.Now()
	
	// Create a more sophisticated color generation algorithm
	// Use hour and day to create a hue that cycles through the spectrum
	hourWeight := float64(now.Hour()) * 15.0 // 0-345 degrees
	dayWeight := float64(now.YearDay()) * 0.986 // Almost full cycle through year
	hue := math.Mod(hourWeight + dayWeight, 360.0)
	
	// Adjust for pleasing color ranges - prefer greens, blues, purples, warm earth tones
	// Avoid harsh yellows and magentas by mapping to more pleasing ranges
	if hue >= 45 && hue < 75 { // yellow range -> shift to green
		hue = 45 + (hue-45)*0.5 // compress yellow range
	} else if hue >= 285 && hue < 315 { // magenta range -> shift to purple/blue
		hue = 270 + (hue-285)*0.5 // compress magenta range
	}
	
	// Use month for saturation variation (45-75% for rich but not overwhelming colors)
	monthSat := 45 + (int(now.Month())%30)
	
	// Use week of year for lightness variation (18-32% for darker, more sophisticated colors)
	weekOfYear := int(now.YearDay() / 7)
	lightness := 18 + (weekOfYear%14)
	
	// Convert HSL to hex
	return HSLToHex(hue, float64(monthSat), float64(lightness))
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
