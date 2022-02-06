# goldmark-markdown

[![GoDoc](https://godoc.org/github.com/teekennedy/goldmark-markdown?status.svg)](https://godoc.org/github.com/teekennedy/goldmark-markdown) [![latest](https://img.shields.io/github/v/tag/teekennedy/goldmark-markdown)](https://github.com/teekennedy/goldmark-markdown/tags) [![test](https://github.com/teekennedy/goldmark-markdown/actions/workflows/test.yml/badge.svg)](https://github.com/teekennedy/goldmark-markdown/actions/workflows/test.yml) [![Coverage Status](https://coveralls.io/repos/github/teekennedy/goldmark-markdown/badge.svg?branch=main)](https://coveralls.io/github/teekennedy/goldmark-markdown?branch=main)

Goldmark-markdown ia a [goldmark] renderer renders to markdown. It can be used directly as an
auto-formatter for markdown source, or extended via goldmark's powerful [AST] transformers to
programmatically transform markdown files.

This module was created for my [update-a-changelog] GitHub Action, to allow it to intelligently
merge new changelog entries from Pull Requests into CHANGELOG.md, as well as add new versions to
CHANGELOG.md when the corresponding tag is pushed.

## As a formatter

You can use goldmark-markdown to format existing markdown documents. It removes extraneous
whitespace, and enforces consistent style for things like indentation, headings, and lists.

## As a markdown transformer

TODO Add an example table of contents generator as a document transformer.

[goldmark]: https://github.com/yuin/goldmark
[AST]: https://pkg.go.dev/github.com/yuin/goldmark/ast
[update-a-changelog]: https://github.com/teekennedy/update-a-changelog
