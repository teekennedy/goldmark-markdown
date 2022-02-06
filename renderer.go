// Package markdown is a goldmark renderer that outputs markdown.
package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// NewNodeRenderer returns a new renderer.NodeRenderer that is configured by default values.
func NewNodeRenderer(options ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config: NewConfig(),
	}
	for _, opt := range options {
		opt.SetMarkdownOption(&r.Config)
	}
	return r
}

// NewRenderer returns a new renderer.Renderer that is configured by default values.
func NewRenderer(options ...Option) renderer.Renderer {
	r := NewNodeRenderer(options...)
	return renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(r, 1000)))
}

// The Renderer struct is an implementation of renderer that renders nodes
// as Markdown
type Renderer struct {
	Config
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	/* TODO
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
	*/

	// inlines
	reg.Register(ast.KindText, r.renderText)
	/* TODO
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	*/
}

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	// Empty headings or headings above level 2 can only be ATX
	if !n.HasChildren() || n.Level > 2 {
		return r.renderATXHeading(w, source, n, entering)
	}
	// Multiline headings can only be Setext
	if n.Lines().Len() > 1 {
		return r.renderSetextHeading(w, source, n, entering)
	}
	// Otherwise it's up to the configuration
	if r.HeadingStyle.IsSetext() {
		return r.renderSetextHeading(w, source, n, entering)
	}
	return r.renderATXHeading(w, source, n, entering)
}

func (r *Renderer) renderATXHeading(w util.BufWriter, source []byte, node *ast.Heading, entering bool) (ast.WalkStatus, error) {
	if entering {
		atxHeadingChars := strings.Repeat("#", node.Level)
		fmt.Fprint(w, atxHeadingChars)
		// Only print space after heading if non-empty
		if node.HasChildren() {
			fmt.Fprint(w, " ")
		}
	} else if r.HeadingStyle == HeadingStyleATXSurround {
		atxHeadingChars := strings.Repeat("#", node.Level)
		fmt.Fprintf(w, " %v", atxHeadingChars)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderSetextHeading(w util.BufWriter, source []byte, node *ast.Heading, entering bool) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}
	underlineChar := [...]string{"", "=", "-"}[node.Level]
	underlineWidth := 3
	if r.HeadingStyle == HeadingStyleFullWidthSetext {
		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			lineWidth := line.Len()

			if lineWidth > underlineWidth {
				underlineWidth = lineWidth
			}
		}
	}
	fmt.Fprintf(w, "\n%v", strings.Repeat(underlineChar, underlineWidth))
	return ast.WalkContinue, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Text)
	if entering {
		_, _ = w.Write(n.Text(source))
		if n.SoftLineBreak() {
			_, _ = w.Write([]byte("\n"))
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.CodeBlock)
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			_, _ = w.Write(r.IndentStyle.bytes())
			_, _ = w.Write(line.Value(source))
		}
	}
	return ast.WalkContinue, nil
}
