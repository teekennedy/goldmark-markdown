package markdown_test

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// RegexpLinkTransformer is an AST Transformer that transforms markdown text that matches a regex
// pattern into a link.
type RegexpLinkTransformer struct {
	LinkPattern *regexp.Regexp
	ReplUrl     []byte
}

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

var source = `
Standup notes:

- Previous day:
    - Gave feedback on TICKET-123.
    - Outlined presentation on syntax-aware markdown transformations.
    - Finished my part of TICKET-456 and assigned to Emily.
- Today:
    - Add integration tests for TICKET-789.
    - Create slides for presentation.
`

func Example() {
	// Instantiate our transformer
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
	// Output buffer
	buf := bytes.Buffer{}

	// Convert parses the source, applies transformers, and renders output to the given io.Writer
	err := gm.Convert([]byte(source), &buf)
	if err != nil {
		log.Fatalf("Encountered Markdown conversion error: %v", err)
	}
	fmt.Print(buf.String())

	// Output:
	// Standup notes:
	//
	// - Previous day:
	//     - Gave feedback on [TICKET-123](https://example.com/TICKET?query=TICKET-123).
	//     - Outlined presentation on syntax-aware markdown transformations.
	//     - Finished my part of [TICKET-456](https://example.com/TICKET?query=TICKET-456) and assigned to Emily.
	// - Today:
	//     - Add integration tests for [TICKET-789](https://example.com/TICKET?query=TICKET-789).
	//     - Create slides for presentation.
}
