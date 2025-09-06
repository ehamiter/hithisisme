## hithisisme

Minimal static site generator with dynamic ROYGBIV color theming that evolves throughout the year.

### Usage

```
# Build and render only (default behavior)
go run dev.go

# Development mode (builds, renders, and serves on localhost:8000)
go run dev.go --serve

# Use a specific theme color (overrides automatic color generation)
THEME_COLOR="#2d5016" go run dev.go
```

### Dynamic Color System

The site automatically generates a new color theme each day based on the current date, progressing through the ROYGBIV spectrum over the year:

- **January**: Red 🔴
- **February/March**: Orange 🟠
- **April/May**: Yellow 🟡
- **June/July**: Green 🟢
- **August/September**: Blue 🔵
- **October/November**: Indigo 🟣
- **December**: Violet 🟪

Each day within a color family gets subtle variations in hue, saturation, and lightness.

### Technical Details

The color system uses a CSS custom properties approach with a template system:

- **CSS Template**: `templates/style.css.template` contains the design with `{{THEME_COLORS}}` placeholder
- **Color Generation**: HSL-based algorithm in `sitegen/colors.go` creates mathematically related light/dark theme variants
- **Build Integration**: Colors are generated and CSS is written during the render process
- **Override Support**: Use `THEME_COLOR` environment variable to manually set colors during development

The system generates complementary colors for buttons, links, gradients, and background tints, ensuring consistent design across both light and dark themes.

### Adding Things

Use the things CLI to add new items to your `things` collection:

```
go run things.go
```

This will prompt you to enter:

  - **Category**: The category for your thing (e.g., "software", "running", "lifestyle")  
  - **Title**: The name/title of the thing  
  - **URL**: A link to the thing  
  - **Description**: A description of the thing  

The CLI will automatically timestamp the entry with the current date and add it to `data/things.json`.

---

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0) 

You are free to use, modify, and distribute this code (including commercially), 
but any modified versions you distribute must also be open source under the GPLv3.  

The software is provided *as is*, without warranty of any kind.
