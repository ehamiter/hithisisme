## hithisisme

Minimal static site generator.

### Usage

```
# Build and render only (default behavior)
go run dev.go

# Development mode (builds, renders, and serves on localhost:8000)
go run dev.go --serve
```


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
