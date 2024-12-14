package markdown

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rhysd/go-fakeio"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var transformer = testHelperASTTransformer{}

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

// TestRenderError tests the renderer's behavior when an error is encountered
func TestRenderError(t *testing.T) {
	err := fmt.Errorf("TestRenderError")
	ew := &errorWriter{err: err}
	renderer := NewRenderer()
	source := []byte("foo")
	paragraph := ast.NewParagraph()
	paragraph.SetBlankPreviousLines(true)
	text := ast.NewTextSegment(text.NewSegment(0, len(source)))
	paragraph.AppendChild(paragraph, text)

	result := renderer.Render(ew, source, paragraph)
	assert.Equal(t, err, result)
}

// TestCustomRenderers tests that the renderer uses any config.NodeRenderers defined by the user
func TestCustomRenderers(t *testing.T) {
	md := goldmark.New(
		goldmark.WithRenderer(NewRenderer()),
		goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(&transformer, 0))),
	)
	buf := bytes.Buffer{}
	source := `# My Tasks
- [x] Add support for custom renderers
	`

	extension.TaskList.Extend(md)
	err := md.Convert([]byte(source), &buf)
	assert.NoError(t, err)
	t.Log(buf.String())
}

// TestRenderedOutput tests that the renderer produces the expected output for all test cases
func TestRenderedOutput(t *testing.T) {
	md := goldmark.New(
		goldmark.WithRenderer(NewRenderer()),
		goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(&transformer, 0))),
	)
	testCases := []struct {
		name     string
		options  []Option
		source   string
		expected string
	}{
		// Document
		{
			"Empty doc",
			[]Option{},
			"",
			"",
		},
		{
			"Non-empty doc trailing newline",
			[]Option{},
			"x",
			"x\n",
		},
		// Headings
		{
			"Setext to ATX heading",
			[]Option{WithHeadingStyle(HeadingStyleATX)},
			"Foo\n---",
			"## Foo\n",
		},
		{
			"ATX to setext heading",
			[]Option{WithHeadingStyle(HeadingStyleSetext)},
			"## FooBar",
			"FooBar\n---\n",
		},
		{
			"Full width setext heading",
			[]Option{WithHeadingStyle(HeadingStyleFullWidthSetext)},
			"Foo Bar\n---",
			"Foo Bar\n-------\n",
		},
		{
			"ATX heading with closing sequence",
			[]Option{WithHeadingStyle(HeadingStyleATXSurround)},
			"## Foo",
			"## Foo ##\n",
		},
		{
			"Empty ATX heading with closing sequence",
			[]Option{WithHeadingStyle(HeadingStyleATXSurround)},
			"##",
			"## ##\n",
		},
		{
			// Setext headings cannot be empty, will always be ATX
			"Empty setext heading",
			[]Option{WithHeadingStyle(HeadingStyleSetext)},
			"##",
			"##\n",
		},
		{
			// ATX headings cannot be multiline, must be setext
			"Multiline ATX heading",
			[]Option{WithHeadingStyle(HeadingStyleATX)},
			"Foo\nBar\n---",
			"Foo\nBar\n---\n",
		},
		// Autolink
		{
			"Url autolink",
			[]Option{},
			"<https://github.com/teekennedy/github-markdown>",
			"<https://github.com/teekennedy/github-markdown>\n",
		},
		{
			"Mailto autolink",
			[]Option{},
			"<foo@bar.com>",
			"<foo@bar.com>\n",
		},
		// Blockquote
		{
			"Blockquote",
			[]Option{},
			"> You will speak\n> an infinite deal\n> of nothing\n\n\\- William Shakespeare",
			"> You will speak\n> an infinite deal\n> of nothing\n\n\\- William Shakespeare\n",
		},
		{
			"Nested blockquote",
			[]Option{},
			"> one\n> > two\n> > > three\n\n> one again",
			"> one\n> > two\n> > > three\n\n> one again\n",
		},
		// Code Block
		{
			"Space indented code block",
			[]Option{},
			"    foo",
			"    foo\n",
		},
		{
			"Tab indented code block",
			[]Option{WithIndentStyle(IndentStyleTabs)},
			"    foo",
			"\tfoo\n",
		},
		{
			"Multiline code block",
			[]Option{WithIndentStyle(IndentStyleSpaces)},
			"\tfoo\n\tbar\n\tbaz",
			"    foo\n    bar\n    baz\n",
		},
		// Code Span
		{
			"Simple code span",
			[]Option{},
			"`foo`",
			"`foo`\n",
		},
		{
			"Multiline code span",
			[]Option{},
			"`foo\nbar`",
			"`foo\nbar`\n",
		},
		{
			"Two-backtick code span",
			[]Option{},
			"``foo ` bar``",
			"``foo ` bar``\n",
		},
		{
			"Reduced backtick code span",
			[]Option{},
			"``foo bar``",
			"`foo bar`\n",
		},
		{
			"Code span preserving leading and trailing spaces",
			[]Option{},
			"` `` `",
			"` `` `\n",
		},
		{
			"Code span preserving surrounding spaces",
			[]Option{},
			"`  ``  `",
			"`  ``  `\n",
		},
		{
			"Unstrippable left space only",
			[]Option{},
			"` a`",
			"` a`\n",
		},
		{
			"Unstrippable only spaces",
			[]Option{},
			"` `\n`  `",
			"` `\n`  `\n",
		},
		{
			"Line-ending treated as space",
			[]Option{},
			"``\nfoo \n``",
			"`foo `\n",
		},
		{
			"Backlashes are treated literally",
			[]Option{},
			"`foo\\`bar`",
			"`foo\\`bar`\n",
		},
		{
			"Two backticks act as delimiters",
			[]Option{},
			"``foo`bar``",
			"``foo`bar``\n",
		},
		{
			"Two backtics inside single ones with spaces trimmed",
			[]Option{},
			"` foo `` bar `",
			"`foo `` bar`\n",
		},
		{
			"Codespan backticks have precedence over emphasis",
			[]Option{},
			"*foo`*`",
			"*foo`*`\n",
		},
		{
			"Codespan backticks have equal precedence with HTML",
			[]Option{},
			"`<a href=\"`\">`",
			"`<a href=\"`\">`\n",
		},
		{
			"HTML tag with backtick",
			[]Option{},
			"<a href=\"`\">`",
			"<a href=\"`\">`\n",
		},
		{
			"Autolink split by a backtick",
			[]Option{},
			"`<http://foo.bar.`baz>`",
			"`<http://foo.bar.`baz>`\n",
		},
		{
			"Unbalanced 3-2 backticks remain intact",
			[]Option{},
			"```foo``",
			"```foo``\n",
		},
		{
			"Unbalanced 1-0 backticks remain intact",
			[]Option{},
			"`foo",
			"`foo\n",
		},
		{
			"Unbalanced double backticks",
			[]Option{},
			"`foo``bar``",
			"`foo`bar`\n",
		},
		// Emphasis
		{
			"Emphasis",
			[]Option{},
			"*emph*",
			"*emph*\n",
		},
		{
			"Strong",
			[]Option{},
			"**strong**",
			"**strong**\n",
		},
		{
			"Strong emphasis",
			[]Option{},
			"***strong emph***",
			"***strong emph***\n",
		},
		{
			"Strong in emphasis",
			[]Option{},
			"***strong** in emph*",
			"***strong** in emph*\n",
		},
		{
			"Emphasis in strong",
			[]Option{},
			"***emph* in strong**",
			"***emph* in strong**\n",
		},
		{
			"Escaped emphasis",
			[]Option{},
			"*escaped\\*emphasis*",
			"*escaped\\*emphasis*\n",
		},
		{
			"In emphasis strong",
			[]Option{},
			"*in emph **strong***",
			"*in emph **strong***\n",
		},
		// Paragraph
		{
			"Simple paragraph",
			[]Option{},
			"foo",
			"foo\n",
		},
		{
			"Paragraph with escaped characters",
			[]Option{},
			"\\# foo \\*bar\\* \\__baz\\_\\_",
			"\\# foo \\*bar\\* \\__baz\\_\\_\n",
		},
		// Thematic Break
		{
			"Thematic break default style",
			[]Option{},
			"---",
			"---\n",
		},
		{
			"Thematic break underline style",
			[]Option{WithThematicBreakStyle(ThematicBreakStyleUnderlined)},
			"---",
			"___\n",
		},
		{
			"Thematic break starred style",
			[]Option{WithThematicBreakStyle(ThematicBreakStyleStarred)},
			"---",
			"***\n",
		},
		{
			// Thematic breaks are a minimum of three characters
			"Thematic break zero value",
			[]Option{WithThematicBreakLength(ThematicBreakLength(0))},
			"---",
			"---\n",
		},
		{
			"Thematic break longer length",
			[]Option{WithThematicBreakLength(ThematicBreakLength(10))},
			"---",
			"----------\n",
		},
		// Fenced Code Block
		{
			"Fenced Code Block",
			[]Option{},
			"```\nfoo\nbar\nbaz\n```",
			"```\nfoo\nbar\nbaz\n```\n",
		},
		{
			"Fenced Code Block with info",
			[]Option{},
			"```ruby startline=3\ndef foo(x)\n  return 3\nend\n```",
			"```ruby startline=3\ndef foo(x)\n  return 3\nend\n```\n",
		},
		{
			"Fenced Code Block with special chars",
			[]Option{},
			"```\n!@#$%^&*\\[],./;'()\n```",
			"```\n!@#$%^&*\\[],./;'()\n```\n",
		},
		// Raw HTML
		{
			"Raw HTML open tags",
			[]Option{},
			"<a><bab><c2c>",
			"<a><bab><c2c>\n",
		},
		{
			"Raw HTML empty elements",
			[]Option{},
			"<a/><b2/>",
			"<a/><b2/>\n",
		},
		{
			"Raw HTML with attributes",
			[]Option{},
			"<a foo=\"bar\" bam = 'baz <em>\"</em>'\n_boolean zoop:33=zoop:33 />",
			"<a foo=\"bar\" bam = 'baz <em>\"</em>'\n_boolean zoop:33=zoop:33 />\n",
		},
		// HTML blocks
		{
			"HTML Block Type 1",
			[]Option{},
			"<pre>\nfoo\n</pre>",
			"<pre>\nfoo\n</pre>\n",
		},
		{
			"HTML Block Type 2",
			[]Option{},
			"<!--\ncomment\n-->",
			"<!--\ncomment\n-->\n",
		},
		{
			"HTML Block Type 3",
			[]Option{},
			"<?\nfoo\n?>",
			"<?\nfoo\n?>\n",
		},
		{
			"HTML Block Type 4",
			[]Option{},
			"<!FOO\n!>",
			"<!FOO\n!>\n",
		},
		{
			"HTML Block Type 5",
			[]Option{},
			"<![CDATA[\nfoo\n]]>",
			"<![CDATA[\nfoo\n]]>\n",
		},
		{
			"HTML Block Type 6",
			[]Option{},
			"<hr />",
			"<hr />\n",
		},
		{
			"HTML Block Type 7",
			[]Option{},
			"</a>",
			"</a>\n",
		},
		// Lists
		{
			"Unordered list",
			[]Option{},
			"- A1\n- B1\n  - C2\n    - D3\n- E1",
			"- A1\n- B1\n  - C2\n    - D3\n- E1\n",
		},
		{
			"Ordered list",
			[]Option{},
			"1. X1\n2. B1\n   1. C2\n      1. D3\n3. E1\n",
			"1. X1\n2. B1\n   1. C2\n      1. D3\n3. E1\n",
		},
		{
			"Mixed list",
			[]Option{},
			"1. A1\n2. B1\n   - C2\n     1. D3\n     2. E3\n   - F2\n   - G2\n3. H1\n",
			"1. A1\n2. B1\n   - C2\n     1. D3\n     2. E3\n   - F2\n   - G2\n3. H1\n",
		},
		{
			"Nested list length",
			[]Option{WithNestedListLength(2)},
			"1. A1\n2. B1\n   - C2\n     1. D3\n     2. E3\n   - F2\n   - G2\n3. H1\n",
			"1. A1\n2. B1\n      - C2\n          1. D3\n          2. E3\n      - F2\n      - G2\n3. H1\n",
		},
		// Block separators
		{
			"ATX heading block separator",
			[]Option{},
			"# Foo\n# Bar\n\n# Baz",
			"# Foo\n# Bar\n\n# Baz\n",
		},
		{
			"Setext heading block separator",
			[]Option{WithHeadingStyle(HeadingStyleSetext)},
			"Foo\n---\nBar\n---\n\nBaz\n---",
			"Foo\n---\nBar\n---\n\nBaz\n---\n",
		},
		{
			"Code block separator",
			[]Option{WithIndentStyle(IndentStyleTabs)},
			"\tcode 1\n---\n\tcode 2\n---\n\n\tcode 3",
			"\tcode 1\n---\n\tcode 2\n---\n\n\tcode 3\n",
		},
		{
			"Fenced code block separator",
			[]Option{},
			"```\ncode 1\n```\n```\ncode 2\n```\n\n```\ncode 3\n```",
			"```\ncode 1\n```\n```\ncode 2\n```\n\n```\ncode 3\n```\n",
		},
		{
			"HTML block separator",
			[]Option{},
			"<?foo?>\n<?bar?>\n\n<?baz?>",
			"<?foo?>\n<?bar?>\n\n<?baz?>\n",
		},
		{
			"List block separator",
			[]Option{},
			"- foo\n+ bar\n\n* baz",
			"- foo\n+ bar\n\n* baz\n",
		},
		{
			"List item block separator",
			[]Option{},
			"- foo\n- bar\n\n- baz",
			"- foo\n- bar\n\n- baz\n",
		},
		{
			"Text block separator",
			[]Option{},
			"- foo\n- bar\n\n- baz",
			"- foo\n- bar\n\n- baz\n",
		},

		// Tight and "loose" lists
		{
			"Tight list",
			[]Option{},
			"Paragraph\n- A1\n- B1",
			"Paragraph\n- A1\n- B1\n",
		},
		{
			"Loose list",
			[]Option{},
			"Paragraph\n\n- A1\n- B1",
			"Paragraph\n\n- A1\n- B1\n",
		},
		// Links
		{
			"Empty Link",
			[]Option{},
			"[]()",
			"[]()\n",
		},
		{
			"Link",
			[]Option{},
			"[link](/uri)",
			"[link](/uri)\n",
		},
		{
			"Link with title",
			[]Option{},
			"[link](/uri \"title\")",
			"[link](/uri \"title\")\n",
		},
		// Images
		{
			"Empty image",
			[]Option{},
			"![]()",
			"![]()\n",
		},
		{
			"Image",
			[]Option{},
			"![image](/uri)",
			"![image](/uri)\n",
		},
		{
			"Image with title",
			[]Option{},
			"![image](/uri \"title\")",
			"![image](/uri \"title\")\n",
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
