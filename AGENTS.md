# AGENTS.md - hithisisme Project Guide

## Build, Test, and Development Commands
- **Build and render**: `go run dev.go` (default behavior, required after any code changes)
- **Development with server**: `go run dev.go --serve` (serves on localhost:8000)
- **Run tests**: `go test ./sitegen` (tests are in sitegen/*_test.go)
- **Run single test**: `go test ./sitegen -run TestName`
- **Add new thing**: `go run things.go` (interactive CLI to add items to data/things.json)
- **Override theme color**: `THEME_COLOR="#2d5016" go run dev.go`

## Architecture
- **DSL**: `hi` language (see SPEC.md) - custom static site DSL for defining site structure
- **Main binary**: Built as `./hi` from `cmd/sitegen/main.go`, runs with `./hi render`
- **Core package**: `sitegen/` - parser, renderer, color system, data fetching, sorting
- **Data sources**: `data/things.json`, `data/apps.json`, `data/repos.json`, `data/languages.json`
- **Templates**: `templates/layout.html` (main layout), `templates/style.css.template` (CSS with `{{THEME_COLORS}}` placeholder)
- **Output**: Generated site goes to `public/index.html` and `public/style.css`
- **Color system**: Dynamic ROYGBIV theming in `sitegen/colors.go` progresses through spectrum based on date

## Code Style
- **License header**: All files must include GPLv3 copyright header (see existing files)
- **Package naming**: `sitegen` for site generation logic, `main` for entry points
- **Error handling**: Return errors up the stack, log at top level, fallback gracefully (reuse last-good JSON on fetch errors)
- **Dependencies**: Minimal - only `github.com/yuin/goldmark` for Markdown rendering
- **Testing**: Standard Go testing in `sitegen/*_test.go`

## Important Notes from Copilot Rules
- CSS framework: Bulma (https://bulma.io/documentation/)
- After changes, verify output in `public/index.html`
- Do NOT start a server unless explicitly asked
- Do NOT create one-off tests unless discussed
