# goldmark-markdown

[![GoDoc](https://godoc.org/github.com/teekennedy/goldmark-markdown?status.svg)](https://godoc.org/github.com/teekennedy/goldmark-markdown) ![Go Version](https://img.shields.io/github/go-mod/go-version/teekennedy/goldmark-markdown) [![latest](https://img.shields.io/github/v/tag/teekennedy/goldmark-markdown)](https://github.com/teekennedy/goldmark-markdown/tags) [![test](https://github.com/teekennedy/goldmark-markdown/actions/workflows/test.yml/badge.svg)](https://github.com/teekennedy/goldmark-markdown/actions/workflows/test.yml) [![Coverage Status](https://coveralls.io/repos/github/teekennedy/goldmark-markdown/badge.svg?branch=main)](https://coveralls.io/github/teekennedy/goldmark-markdown?branch=main)

Goldmark-markdown is a [goldmark] renderer that renders to markdown. It can be used directly as an
auto-formatter for markdown source, or extended via goldmark's powerful [AST] transformers to
programmatically transform markdown files.

This module was created for my [update-a-changelog] GitHub Action, to allow it to intelligently
merge new changelog entries from Pull Requests into CHANGELOG.md, as well as add new versions to
CHANGELOG.md when the corresponding tag is pushed.

## As a formatter

You can use goldmark-markdown to format existing markdown documents. It removes extraneous
whitespace, and enforces consistent style for things like indentation, headings, and lists.

```go
// Create goldmark converter with markdown renderer object
// Can pass functional Options as arguments. This example converts headings to ATX style.
renderer := markdown.NewRenderer(markdown.WithHeadingStyle(markdown.HeadingStyleATX))
md := goldmark.New(goldmark.WithRenderer(renderer))

// "Convert" markdown to formatted markdown
source := `
My Document Title
=================
`
buf := bytes.Buffer{}
err := md.Convert([]byte(source), &buf)
if err != nil {
  log.Fatal(err)
}
log.Print(buf.String()) // # My Document Title
```

### Options

You can control the style of various markdown elements via functional options that are passed to
the renderer.

| Functional Option | Type | Description |
| ----------------- | ---- | ----------- |
| WithIndentStyle | markdown.IndentStyle | Indent nested blocks with spaces or tabs. |
| WithHeadingStyle | markdown.HeadingStyle | Render markdown headings as ATX (`#`-based), Setext (underlined with `===` or `---`), or variants thereof. |
| WithThematicBreakStyle | markdown.ThematicBreakStyle | Render thematic breaks with `-`, `*`, or `_`. |
| WithThematicBreakLength | markdown.ThematicBreakLength | Number of characters to use in a thematic break (minimum 3). |

## As a markdown transformer

Goldmark supports writing transformers that can inspect and modify the parsed markdown [AST] before
it gets sent to the renderer for output. You can use transformers in conjunction with
goldmark-markdown's renderer to make changes to markdown sources while preserving valid syntax.

For example, you can scan the AST for text that matches a pattern for an external resource, and
transform that text into a link to the resource, similar to GitHub's [custom autolinks] feature.
Start by adding a struct that holds a regexp pattern to scan text for and a URL replacement for the
pattern:


```go
// RegexpLinkTransformer is an AST Transformer that transforms markdown text that matches a regex
// pattern into a link.
type RegexpLinkTransformer struct {
	LinkPattern *regexp.Regexp
	ReplUrl     []byte
}
```

Next, implement a `Transform` function that walks the AST and calls a `LinkifyText` function on any
Text nodes encountered:

```go
// Transform implements goldmark.parser.ASTTransformer
func (t *RegexpLinkTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	// Walk the AST in depth-first fashion and apply transformations
	err := ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		// Each node will be visited twice, once when it is first encountered (entering), and again
		// after all the node's children have been visited (if any). Skip the latter.
		if !entering {
			return ast.WalkContinue, nil
		}
		// Skip the children of existing links to prevent double-transformation.
		if node.Kind() == ast.KindLink || node.Kind() == ast.KindAutoLink {
			return ast.WalkSkipChildren, nil
		}
		// Linkify any Text nodes encountered
		if node.Kind() == ast.KindText {
			textNode := node.(*ast.Text)
			t.LinkifyText(textNode, source)
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		log.Fatal("Error encountered while transforming AST:", err)
	}
}
```

The function passed to ast.Walk will be called for every node visited. The function controls which
node is visited next via its return value. `ast.WalkContinue` causes the walk to continue to the
next child, sibling, or parent node, depending on whether such nodes exist. `ast.WalkSkipChildren`
continues only to sibling or parent nodes. Both `ast.WalkStop` and returning a non-nil error cause
the walk to end.

For this example, we are only interested in linkifying Text nodes that are not already part of a
link, and continue past everything else.

In order to "linkify" the Text nodes, the transformer will replace the original Text node with
nodes for before and after the link, as well as the node for the link itself:

```go
// LinkifyText finds all LinkPattern matches in the given Text node and replaces them with Link
// nodes that point to ReplUrl.
func (t *RegexpLinkTransformer) LinkifyText(node *ast.Text, source []byte) {
	parent := node.Parent()
	tSegment := node.Segment
	match := t.LinkPattern.FindIndex(tSegment.Value(source))
	if match == nil {
		return
	}
	// Create a text.Segment for the link text.
	lSegment := text.NewSegment(tSegment.Start+match[0], tSegment.Start+match[1])

	// Insert node for any text before the link
	if lSegment.Start != tSegment.Start {
		bText := ast.NewTextSegment(tSegment.WithStop(lSegment.Start))
		parent.InsertBefore(parent, node, bText)
	}

	// Insert Link node
	link := ast.NewLink()
	link.AppendChild(link, ast.NewTextSegment(lSegment))
	link.Destination = t.LinkPattern.ReplaceAll(lSegment.Value(source), t.ReplUrl)
	parent.InsertBefore(parent, node, link)

	// Update original node to represent the text after the link (may be empty)
	node.Segment = tSegment.WithStart(lSegment.Stop)

	// Linkify remaining text if not empty
	if node.Segment.Len() > 0 {
		t.LinkifyText(node, source)
	}
}
```

To use this transformer, we'll need to instantiate one or more RegexpLinkTransformer structs, then
prioritize them and add them to the parser configuration of the goldmark object. The
transformation(s) will then be automatically applied to all markdown documents converted by the
goldmark object.

```go
transformer := RegexpLinkTransformer{
  LinkPattern: regexp.MustCompile(`TICKET-\d+`),
  ReplUrl:     []byte("https://example.com/TICKET?query=$0"),
}
// Goldmark supports multiple AST transformers and runs them sequentially in order of priority.
prioritizedTransformer := util.Prioritized(&transformer, 0)
// Setup goldmark with the markdown renderer and our transformer
gm := goldmark.New(
  goldmark.WithRenderer(markdown.NewRenderer()),
  goldmark.WithParserOptions(parser.WithASTTransformers(prioritizedTransformer)),
)
```

The complete example can be found in [autolink_example_test.go], or in the go doc for this package.

[AST]: https://pkg.go.dev/github.com/yuin/goldmark/ast
[autolink_example_test.go]: /autolink_example_test.go
[custom autolinks]: https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/autolinked-references-and-urls#custom-autolinks-to-external-resources
[goldmark]: https://github.com/yuin/goldmark
[update-a-changelog]: https://github.com/teekennedy/update-a-changelog
