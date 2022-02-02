package renderer

import (
	"bytes"
	"testing"

	"github.com/rhysd/go-fakeio"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	transformer = testHelperASTTransformer{}
	md          = goldmark.New(
		goldmark.WithRenderer(NewRenderer()),
		goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(&transformer, 0))),
	)
)

// testHelperASTTransformer is a goldmark AST transformer that helps with debugging failed tests.
type testHelperASTTransformer struct {
	lastDocument *ast.Document
}

// Transform implements goldmark.parser.ASTTransformer
func (t *testHelperASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	t.lastDocument = node
}

// DumpLastAST calls Node.Dump() on the most recently parsed AST and returns its output.
func (t *testHelperASTTransformer) DumpLastAST(source []byte) string {
	// Node.Dump() is hardcoded to write to stdout. Fake stdout to get the output as a string.
	result, err := fakeio.Stdout().Do(func() {
		transformer.lastDocument.Dump(source, 0)
	})
	if err != nil {
		panic(err)
	}
	return result
}

var testCases = []struct {
	name     string
	options  []Option
	source   string
	expected string
}{
	// Headings
	{
		"Setext to ATX heading",
		[]Option{WithHeadingStyle(HeadingStyleATX)},
		"Foo\n---",
		"## Foo",
	},
	{
		"ATX to setext heading",
		[]Option{WithHeadingStyle(HeadingStyleSetext)},
		"## FooBar",
		"FooBar\n---",
	},
	{
		"ATX to setext heading",
		[]Option{WithHeadingStyle(HeadingStyleFullWidthSetext)},
		"Foo Bar\n---",
		"Foo Bar\n-------",
	},
	{
		"ATX heading with closing sequence",
		[]Option{WithHeadingStyle(HeadingStyleATXSurround)},
		"## Foo",
		"## Foo ##",
	},
	{
		// Setext headings cannot be empty, will always be ATX
		"Empty setext heading",
		[]Option{WithHeadingStyle(HeadingStyleSetext)},
		"##",
		"##",
	},
	{
		// ATX headings cannot be multiline, must be setext
		"Multiline ATX heading",
		[]Option{WithHeadingStyle(HeadingStyleATX)},
		"Foo\nBar\n---",
		"Foo\nBar\n---",
	},
}

// TestRenderedOutput tests that the renderer produces the expected output for all test cases
func TestRenderedOutput(t *testing.T) {

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			buf := bytes.Buffer{}

			renderer := NewRenderer(tc.options...)
			md.SetRenderer(renderer)
			err := md.Convert([]byte(tc.source), &buf)
			assert.NoError(err)
			assert.Equal(tc.expected, buf.String())
			t.Logf("Markdown source: %q\n", tc.source)
			t.Log(transformer.DumpLastAST([]byte(tc.source)))
		})
	}
}
