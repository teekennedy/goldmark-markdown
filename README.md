# goldmark-markdown

[![GoDoc](https://godoc.org/github.com/teekennedy/goldmark-markdown?status.svg)](https://godoc.org/github.com/teekennedy/goldmark-markdown) ![Go Version](https://img.shields.io/github/go-mod/go-version/teekennedy/goldmark-markdown) [![latest](https://img.shields.io/github/v/tag/teekennedy/goldmark-markdown)](https://github.com/teekennedy/goldmark-markdown/tags) [![test](https://github.com/teekennedy/goldmark-markdown/actions/workflows/test.yml/badge.svg)](https://github.com/teekennedy/goldmark-markdown/actions/workflows/test.yml) [![Coverage Status](https://coveralls.io/repos/github/teekennedy/goldmark-markdown/badge.svg?branch=main)](https://coveralls.io/github/teekennedy/goldmark-markdown?branch=main)

Goldmark-markdown ia a [goldmark] renderer that renders to markdown. It can be used directly as an
auto-formatter for markdown source, or extended via goldmark's powerful [AST] transformers to
programmatically transform markdown files.

This module was created for my [update-a-changelog] GitHub Action, to allow it to intelligently
merge new changelog entries from Pull Requests into CHANGELOG.md, as well as add new versions to
CHANGELOG.md when the corresponding tag is pushed.

## As a formatter

You can use goldmark-markdown to format existing markdown documents. It removes extraneous
whitespace, and enforces consistent style for things like indentation, headings, and lists.

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

	// Insert node for any text after the link, and apply line break flags from original Text
	aText := ast.NewTextSegment(tSegment.WithStart(lSegment.Stop))
	aText.SetSoftLineBreak(node.SoftLineBreak())
	aText.SetHardLineBreak(node.HardLineBreak())
	parent.InsertBefore(parent, node, aText)

	// Linkify after text if not empty
	if aText.Segment.Len() > 0 {
		t.LinkifyText(aText, source)
	}

	// Remove original node
	parent.RemoveChild(parent, node)
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
