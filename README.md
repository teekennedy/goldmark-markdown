# goldmark-markdown

[![GoDoc](https://godoc.org/github.com/cyphus/goldmark-markdown?status.svg)](https://godoc.org/github.com/cyphus/goldmark-markdown)

Goldmark-markdown ia a [goldmark] extension that renders and formats markdown.
It can be used directly as an auto-formatter for markdown source, or extended
via goldmark's powerful [AST] to programmatically transform markdown files.

This package was created for my [update-a-changelog] GitHub Action, to allow it
to programmatically add changelog entries to CHANGELOG.md.

## As a formatter

Without any formatting options specified, goldmark-markdown will preserve as
much of the original format as it can, effectively returning the source
unchanged. This isn't very useful on its own, but acts as a baseline behavior
to allow users to pick and choose which formatting options to enable.

## As a markdown transformer

TODO

[goldmark]: https://github.com/yuin/goldmark
[AST]: https://pkg.go.dev/github.com/yuin/goldmark/ast
[update-a-changelog]: https://github.com/cyphus/update-a-changelog
