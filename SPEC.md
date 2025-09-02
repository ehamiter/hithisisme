# `index.hi` DSL — Spec

## Purpose

A tiny, human-readable format for building a static homepage from a few JSON/markdown sources without writing code each time. The renderer:

1. Resolves **bindings** (fetch/glob/manual),
2. Evaluates **sections** and **loops**, and
3. Emits HTML into a layout (e.g., with MVP.css).

---

## File format

### 0) Comments

* Lines starting with `#` are comments and ignored.

### 1) Bindings (header)

Declare named sources. Whitespace is cosmetic.

```
name = <target> [<< <url_or_template>]
```

* **Eager fetch**: `apps = apps.json << https://...`
  Fetch now and write JSON to `apps.json`. Bind `apps` to its parsed value.
* **Lazy fetch**: `!languages = languages.json << https://.../{repo.name}/languages`
  Don’t fetch yet. Store a **URL template**. When referenced **inside** a loop with the needed context (e.g., `repo.name`), expand + fetch once, cache in `languages.json` under a key (see below).
* **Manual JSON**: `things = things.json` (no fetch; load file as is).
* **Glob**: `posts = ./posts/**` (resolved to a list of files; only `*.md` are used).

#### Posts glob semantics

From `./posts/*.md` (expected filename `YYYY-MM-DD.md`) derive:

* `post.date`   → `YYYY-MM-DD` from filename
* `post.title`  → first Markdown heading `# Title` (fallback: filename)
* `post.preview`→ first paragraph (truncated with `…` if long)
* `post.url`    → `/posts/YYYY-MM-DD/` (or whatever convention the renderer uses)

#### Lazy languages caching shape

When `languages` is resolved for a repo named `X`, cache to `languages.json` as:

```json
{ "X": { "Swift": 12345, "Shell": 234 } }
```

---

### 2) Sections

```
{section_id: Freeform text (Markdown allowed).}
```

* Renders as `<section id="section_id">…</section>`.
* Keep `section_id` to `[a-z0-9_-]` for valid HTML ids.

---

### 3) Loops

```
[for <var> in <source> : <sort_keys>?]
  <lines...>
```

* **Body** = the indented lines following, until the next non-indented block or EOF.
* `<source>` can be:

  * a bound name (e.g., `apps.results`, `repos`, `posts`, `things`)
  * a map (iterate values; use two vars for key/value)
* **Fields** are dotted paths, e.g., `repo.name`, `app.trackViewUrl`.

#### Sorting

* `<sort_keys>` = comma-separated field names.
* **Direction rule:** **DESC by default**; append `^` to a key for **ASC**.

  * Examples:

    * `currentVersionReleaseDate` → newest first
    * `stargazers_count, updated_at` → high-stars first, then latest updated
    * `date_published, category^, title^` → newest first, then category/title A→Z
* If a sort key is missing/invalid, it’s ignored; if all are invalid, original order is kept.
* Nulls sort **last**.

#### Nested loops & lazy sources

Inside a repo loop:

```
[for repo in repos: stargazers_count, updated_at]
  repo.name
  [for name, count in languages: count]
    name
```

* `languages` resolves lazily using the current `repo.name` (template from header), fetches once, caches, and yields a map to iterate.

---

## Data source specifics

### `apps` (iTunes Lookup)

* Renderer should use `apps.results` **filtered** to items where `kind == "software"`.
* Common fields used here:

  * `app.trackName`, `app.trackViewUrl`, `app.version`,
  * `app.description`, `app.genres`, `app.currentVersionReleaseDate`
* Typical sort: `currentVersionReleaseDate` (DESC default).

### `repos` (GitHub REST `/users/{u}/repos`)

* Fields used:

  * `repo.name`, `repo.html_url`, `repo.description`,
  * `repo.updated_at`, `repo.stargazers_count`
* `watchers_count` mirrors stars in REST—omit to avoid confusion.
* Typical sort: `stargazers_count, updated_at` (both DESC).

### `languages` (GitHub REST `/repos/{u}/{repo}/languages`)

* Lazy, per-repo via template; see “lazy languages caching shape” above.
* Inner loop often sorts by value: `[for name, count in languages: count]` (DESC default).

### `things` (manual JSON)

* Expect fields: `thing.title`, `thing.url`, `thing.description`, `thing.date_published` (ISO date), optional `thing.category`.
* Typical sort: `date_published, category^, title^`.

---

## Rendering rules

* Each **section** becomes a `<section>`; section body supports Markdown → HTML.
* Each **loop item** becomes an `<article>` (one line per field as a `<p>` or `<div>`). Keep markup minimal to play nice with MVP.css.
* HTML escaping on by default (after Markdown rendering where applicable).

---

## Errors & resilience

* **Fetch errors** (eager or lazy): reuse last-good JSON on disk and continue; log a warning.
* **Unknown source**: warn and skip the loop.
* **Missing field**: render empty string; warn once per field name per run.
* Footer may include a small “Last updated YYYY-MM-DD” timestamp (optional).

---

## Example (excerpt)

```hi
apps       = apps.json      << https://itunes.apple.com/lookup?id=1482332471&entity=software&country=US
repos      = repos.json     << https://api.github.com/users/ehamiter/repos?sort=pushed&direction=desc
!languages = languages.json << https://api.github.com/repos/ehamiter/{repo.name}/languages
things     = things.json
posts      = ./posts/**

{hi: Hi, welcome to my home page. This is a digital garden of sorts.}

{posts: Random writing.}
[for post in posts: date]
  post.title
  post.url
  post.preview
  post.date

{apps: I've published a few things on the app store.}
[for app in apps.results: currentVersionReleaseDate]
  app.trackName
  app.trackViewUrl
  app.version
  app.description
  app.genres
  app.currentVersionReleaseDate

{repos: Some projects I've made or am working on.}
[for repo in repos: stargazers_count, updated_at]
  repo.name
  repo.html_url
  repo.description
  repo.updated_at
  repo.stargazers_count
  [for name, count in languages: count]
    name

{things: A curated list I recommend.}
[for thing in things: date_published, category^, title^]
  thing.category
  thing.title
  thing.url
  thing.description
  thing.date_published
```

---

## Editor & repo niceties (optional)

* **Syntax highlight**: map `*.hi` → Makefile in your editor; add
  `.gitattributes`: `*.hi linguist-language=Makefile`
