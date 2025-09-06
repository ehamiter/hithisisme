---
applyTo: '**'
---
This is a static site generator app written in golang and a DSL called `hi`.

A file `index.hi` is written as a blueprint for a web site; the app generates an html page out of it. It is documented so it serves as a self-documenting file to the language itself.

SPEC.md is the DSL spec for `hi`.

Any time changes are made, the `hi` program needs to be re-built and the html data re-rendered. Running this command will accomplish that:

```
go run dev.go
```

This command is sufficient for you to run to ensure everything works properly. Do not create one-off tests or run any other go commands unless we discuss it explicitly.

This app uses the Bulma CSS framework:
https://r.jina.ai/https://bulma.io/documentation/

For the structure of the components of the app, we have a few dedicated objects:

`things`    = things.json
`apps`      = apps.json
`repos`     = repos.json
`languages` = languages.json
