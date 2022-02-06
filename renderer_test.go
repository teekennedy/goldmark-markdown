package markdown

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

// TestRenderedOutput tests that the renderer produces the expected output for all test cases
func TestRenderedOutput(t *testing.T) {
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
			"Full width setext heading",
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
			"Empty ATX heading with closing sequence",
			[]Option{WithHeadingStyle(HeadingStyleATXSurround)},
			"##",
			"## ##",
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
		// Code Block
		{
			"Space indented code block",
			[]Option{},
			"    foo",
			"    foo",
		},
		{
			"Tab indented code block",
			[]Option{WithIndentStyle(IndentStyleTabs)},
			"    foo",
			"\tfoo",
		},
		{
			"Multiline code block",
			[]Option{WithIndentStyle(IndentStyleSpaces)},
			"\tfoo\n\tbar\n\tbaz",
			"    foo\n    bar\n    baz",
		},
		// Paragraph
		{
			"Simple paragraph",
			[]Option{},
			"foo",
			"foo",
		},
		{
			"Paragraph with escaped characters",
			[]Option{},
			"\\# foo \\*bar\\* \\__baz\\_\\_",
			"\\# foo \\*bar\\* \\__baz\\_\\_",
		},
		// Thematic Break
		{
			"Thematic break default style",
			[]Option{},
			"---",
			"---",
		},
		{
			"Thematic break underline style",
			[]Option{WithThematicBreakStyle(ThematicBreakStyleUnderlined)},
			"---",
			"___",
		},
		{
			"Thematic break starred style",
			[]Option{WithThematicBreakStyle(ThematicBreakStyleStarred)},
			"---",
			"***",
		},
		{
			// Thematic breaks are a minimum of three characters
			"Thematic break zero value",
			[]Option{WithThematicBreakLength(ThematicBreakLength(0))},
			"---",
			"---",
		},
		{
			"Thematic break longer length",
			[]Option{WithThematicBreakLength(ThematicBreakLength(10))},
			"---",
			"----------",
		},
		// Fenced Code Block
		{
			"Fenced Code Block",
			[]Option{},
			"```\nfoo\nbar\nbaz\n```",
			"```\nfoo\nbar\nbaz\n```",
		},
		{
			"Fenced Code Block with language",
			[]Option{},
			"```ruby\ndef foo(x)\n  return 3\nend\n```",
			"```ruby\ndef foo(x)\n  return 3\nend\n```",
		},
		{
			"Fenced Code Block with special chars",
			[]Option{},
			"```\n!@#$%^&*\\[],./;'()\n```",
			"```\n!@#$%^&*\\[],./;'()\n```",
		},
		// HTML blocks
		// Trailing newline is necessary to avoid empty blockquote in AST
		// https://github.com/yuin/goldmark/issues/274
		{
			"HTML Block Type 1",
			[]Option{},
			"<pre>\nfoo\n</pre>\n",
			"<pre>\nfoo\n</pre>\n",
		},
		{
			"HTML Block Type 2",
			[]Option{},
			"<!--\ncomment\n-->\n",
			"<!--\ncomment\n-->\n",
		},
		{
			"HTML Block Type 3",
			[]Option{},
			"<?\nfoo\n?>\n",
			"<?\nfoo\n?>\n",
		},
		{
			"HTML Block Type 4",
			[]Option{},
			"<!FOO\n!>\n",
			"<!FOO\n!>\n",
		},
		{
			"HTML Block Type 5",
			[]Option{},
			"<![CDATA[\nfoo\n]]>\n",
			"<![CDATA[\nfoo\n]]>\n",
		},
		{
			"HTML Block Type 6",
			[]Option{},
			"<hr />",
			"<hr />",
		},
		{
			"HTML Block Type 7",
			[]Option{},
			"</a>",
			"</a>",
		},
		// Block separator
		{
			"Block separator",
			[]Option{},
			"## ATX Heading\nSetext Heading\n---\nparagraph\n\n--- thematic break\n",
			"## ATX Heading\n\n## Setext Heading\n\nparagraph\n\n--- thematic break",
		},
	}

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
