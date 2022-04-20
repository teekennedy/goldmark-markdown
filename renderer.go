// Package markdown is a goldmark renderer that outputs markdown.
package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// NewNodeRenderer returns a new markdown Renderer that is configured by default values.
func NewNodeRenderer(options ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config: NewConfig(),
		rc:     renderContext{},
		writer: renderWriter{},
	}
	for _, opt := range options {
		opt.SetMarkdownOption(&r.Config)
	}
	return r
}

// NewRenderer returns a new renderer.Renderer containing a markdown NodeRenderer with defaults.
func NewRenderer(options ...Option) renderer.Renderer {
	r := NewNodeRenderer(options...)
	return renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(r, 1000)))
}

// Renderer is an implementation of renderer.Renderer that renders nodes as Markdown
type Renderer struct {
	Config
	rc     renderContext
	writer renderWriter
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.chainRenderers(r.renderBlockSeparator, r.renderNothing))
	reg.Register(ast.KindHeading, r.chainRenderers(r.renderBlockSeparator, r.renderHeading))
	reg.Register(ast.KindCodeBlock, r.chainRenderers(r.renderBlockSeparator, r.renderCodeBlock))
	reg.Register(ast.KindParagraph, r.chainRenderers(r.renderBlockSeparator, r.renderNothing))
	// text blocks are essentially paragraphs that aren't wrapped with a <p> tag,
	// a distinction which doesn't apply to this markdown renderer.
	reg.Register(ast.KindTextBlock, r.chainRenderers(r.renderBlockSeparator, r.renderNothing))
	reg.Register(ast.KindThematicBreak, r.chainRenderers(r.renderBlockSeparator, r.renderThematicBreak))
	reg.Register(ast.KindFencedCodeBlock, r.chainRenderers(r.renderBlockSeparator, r.renderFencedCodeBlock))
	reg.Register(ast.KindHTMLBlock, r.chainRenderers(r.renderBlockSeparator, r.renderHTMLBlock))
	reg.Register(ast.KindList, r.chainRenderers(r.renderBlockSeparator, r.renderList))
	reg.Register(ast.KindListItem, r.chainRenderers(r.renderBlockSeparator, r.renderListItem))
	/* TODO
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	*/

	// inlines
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindLink, r.renderLink)
	/* TODO
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	*/
}

func (r *Renderer) renderBlockSeparator(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// Add blank previous line if applicable
		if node.HasBlankPreviousLines() && node.PreviousSibling() != nil {
			r.writer.WriteString(w, "\n")
		}
	} else {
		// Add trailing newline to block if not already present
		if r.writer.lastWrittenByte != byte('\n') {
			r.writer.WriteString(w, "\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) chainRenderers(renderers ...renderer.NodeRendererFunc) renderer.NodeRendererFunc {
	return func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		var walkStatus ast.WalkStatus
		var err error
		for i := range renderers {
			// go through renderers in reverse when exiting
			if !entering {
				i = len(renderers) - 1 - i
			}
			walkStatus, err = renderers[i](w, source, node, entering)
			if err != nil {
				break
			}
		}
		return walkStatus, err
	}
}

func (r *Renderer) renderNothing(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
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
	} else {
		if r.HeadingStyle == HeadingStyleATXSurround {
			atxHeadingChars := strings.Repeat("#", node.Level)
			fmt.Fprintf(w, " %v", atxHeadingChars)
		}
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

func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		breakChar := [...]string{"-", "*", "_"}[r.ThematicBreakStyle]
		var breakLen int
		if r.ThematicBreakLength < ThematicBreakLengthMinimum {
			breakLen = int(ThematicBreakLengthMinimum)
		} else {
			breakLen = int(r.ThematicBreakLength)
		}
		r.writer.WriteString(w, strings.Repeat(breakChar, breakLen))
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.CodeBlock)
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			r.writer.WriteString(w, r.IndentStyle.String())
			r.writer.Write(w, line.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	r.writer.WriteString(w, "```")
	if entering {
		if lang := n.Language(source); lang != nil {
			r.writer.Write(w, lang)
		}
		r.writer.WriteString(w, "\n")
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			r.writer.Write(w, line.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			r.writer.Write(w, line.Value(source))
		}
	} else {
		if n.HasClosure() {
			closure := n.ClosureLine
			r.writer.Write(w, closure.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.List)
		r.rc.listMarker = n.Marker
		r.rc.listIndent += 1
	} else {
		r.rc.listIndent -= 1
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.ListItem)
		r.writer.WriteString(w, r.IndentStyle.Indent(r.rc.listIndent-1))
		r.writer.Write(w, []byte{r.rc.listMarker})
		if n.HasChildren() {
			r.writer.WriteString(w, " ")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Text)
	if entering {
		r.writer.Write(w, n.Text(source))
		if n.SoftLineBreak() {
			r.writer.WriteString(w, "\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		r.writer.Write(w, []byte("["))
	} else {
		link := fmt.Sprintf("](%s", n.Destination)
		r.writer.WriteString(w, link)
		if len(n.Title) > 0 {
			title := fmt.Sprintf(" \"%s\"", n.Title)
			r.writer.WriteString(w, title)
		}
		r.writer.WriteString(w, ")")
	}
	return ast.WalkContinue, nil
}

type renderContext struct {
	// listIndent is the current indentation level for List
	listIndent int
	// listMarker is the marker character used for the current list
	listMarker byte
}

// renderWriter wraps util.BufWriter methods to implement error handling.
type renderWriter struct {
	// err holds the last write error. If non-nil, all write operations become no-ops
	err error
	// lastWrittenByte holds the last byte of the last write operation.
	lastWrittenByte byte
}

// Write writes the given bytes content to the given writer.
func (r *renderWriter) Write(writer util.BufWriter, content []byte) {
	if r.err != nil {
		return
	}
	var writeLen int
	writeLen, r.err = writer.Write(content)
	if writeLen > 0 {
		r.lastWrittenByte = content[writeLen-1]
	}
}

// WriteString writes the given string content to the given writer.
func (r *renderWriter) WriteString(writer util.BufWriter, content string) {
	if r.err != nil {
		return
	}
	var writeLen int
	writeLen, r.err = writer.WriteString(content)
	if writeLen > 0 {
		r.lastWrittenByte = content[writeLen-1]
	}
}

// Reset resets the error and last written byte state of the renderWriter
func (r *renderWriter) Reset() {
	r.err = nil
	r.lastWrittenByte = byte(0)
}
