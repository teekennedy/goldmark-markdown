// Package markdown is a goldmark renderer that outputs markdown.
package markdown

import (
	"bytes"
	"fmt"
	"io"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
)

// NewRenderer returns a new markdown Renderer that is configured by default values.
func NewRenderer(options ...Option) *Renderer {
	r := &Renderer{
		config: NewConfig(),
		rc:     renderContext{},
	}
	for _, opt := range options {
		opt.SetMarkdownOption(r.config)
	}
	return r
}

// Renderer is an implementation of renderer.Renderer that renders nodes as Markdown
type Renderer struct {
	config *Config
	rc     renderContext
}

// AddOptions implements renderer.Renderer.AddOptions
func (r *Renderer) AddOptions(opts ...renderer.Option) {
	config := renderer.NewConfig()
	for _, opt := range opts {
		opt.SetConfig(config)
	}
	for name, value := range config.Options {
		r.config.SetOption(name, value)
	}
	// TODO handle any config.NodeRenderers set by opts
}

// Render implements renderer.Renderer.Render
func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.rc = newRenderContext(w, source, r.config)
	/* TODO
	reg.Register(ast.KindString, r.renderString)
	*/
	return ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return r.getRenderer(n)(n, entering), r.rc.writer.Err()
	})
}

// nodeRenderer is a markdown node renderer func.
type nodeRenderer func(ast.Node, bool) ast.WalkStatus

func (r *Renderer) getRenderer(node ast.Node) nodeRenderer {
	renderers := []nodeRenderer{}
	switch node.Type() {
	case ast.TypeBlock:
		renderers = append(renderers, r.renderBlockSeparator)
	}
	switch node.Kind() {
	case ast.KindAutoLink:
		renderers = append(renderers, r.renderAutoLink)
	case ast.KindHeading:
		renderers = append(renderers, r.renderHeading)
	case ast.KindBlockquote:
		renderers = append(renderers, r.renderBlockquote)
	case ast.KindCodeBlock:
		renderers = append(renderers, r.renderCodeBlock)
	case ast.KindCodeSpan:
		renderers = append(renderers, r.renderCodeSpan)
	case ast.KindEmphasis:
		renderers = append(renderers, r.renderEmphasis)
	case ast.KindThematicBreak:
		renderers = append(renderers, r.renderThematicBreak)
	case ast.KindFencedCodeBlock:
		renderers = append(renderers, r.renderFencedCodeBlock)
	case ast.KindHTMLBlock:
		renderers = append(renderers, r.renderHTMLBlock)
	case ast.KindImage:
		renderers = append(renderers, r.renderImage)
	case ast.KindList:
		renderers = append(renderers, r.renderList)
	case ast.KindListItem:
		renderers = append(renderers, r.renderListItem)
	case ast.KindRawHTML:
		renderers = append(renderers, r.renderRawHTML)
	case ast.KindText:
		renderers = append(renderers, r.renderText)
	case ast.KindLink:
		renderers = append(renderers, r.renderLink)
	}
	return r.chainRenderers(renderers...)
}

func (r *Renderer) chainRenderers(renderers ...nodeRenderer) nodeRenderer {
	return func(node ast.Node, entering bool) ast.WalkStatus {
		var walkStatus ast.WalkStatus
		for i := range renderers {
			// go through renderers in reverse when exiting
			if !entering {
				i = len(renderers) - 1 - i
			}
			walkStatus = renderers[i](node, entering)
		}
		return walkStatus
	}
}

func (r *Renderer) renderBlockSeparator(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		// Add blank previous line if applicable
		if node.PreviousSibling() != nil && node.HasBlankPreviousLines() {
			r.rc.writer.EndLine()
		}
	} else {
		// Flush line buffer to complete line written by previous block
		r.rc.writer.FlushLine()
	}
	return ast.WalkContinue
}

func (r *Renderer) renderAutoLink(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.AutoLink)
	if entering {
		r.rc.writer.Write([]byte("<"))
		r.rc.writer.Write(n.URL(r.rc.source))
	} else {
		r.rc.writer.Write([]byte(">"))
	}
	return ast.WalkContinue
}

func (r *Renderer) renderBlockquote(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.rc.writer.PushPrefix([]byte("> "))
	} else {
		r.rc.writer.PopPrefix()
	}
	return ast.WalkContinue
}

func (r *Renderer) renderHeading(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.Heading)
	// Empty headings or headings above level 2 can only be ATX
	if !n.HasChildren() || n.Level > 2 {
		return r.renderATXHeading(n, entering)
	}
	// Multiline headings can only be Setext
	if n.Lines().Len() > 1 {
		return r.renderSetextHeading(n, entering)
	}
	// Otherwise it's up to the configuration
	if r.config.HeadingStyle.IsSetext() {
		return r.renderSetextHeading(n, entering)
	}
	return r.renderATXHeading(n, entering)
}

func (r *Renderer) renderATXHeading(node *ast.Heading, entering bool) ast.WalkStatus {
	if entering {
		r.rc.writer.Write(bytes.Repeat([]byte("#"), node.Level))
		// Only print space after heading if non-empty
		if node.HasChildren() {
			r.rc.writer.Write([]byte(" "))
		}
	} else {
		if r.config.HeadingStyle == HeadingStyleATXSurround {
			r.rc.writer.Write([]byte(" "))
			r.rc.writer.Write(bytes.Repeat([]byte("#"), node.Level))
		}
	}
	return ast.WalkContinue
}

func (r *Renderer) renderSetextHeading(node *ast.Heading, entering bool) ast.WalkStatus {
	if entering {
		return ast.WalkContinue
	}
	underlineChar := [...][]byte{[]byte(""), []byte("="), []byte("-")}[node.Level]
	underlineWidth := 3
	if r.config.HeadingStyle == HeadingStyleFullWidthSetext {
		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			lineWidth := line.Len()

			if lineWidth > underlineWidth {
				underlineWidth = lineWidth
			}
		}
	}
	r.rc.writer.Write([]byte("\n"))
	r.rc.writer.Write(bytes.Repeat(underlineChar, underlineWidth))
	return ast.WalkContinue
}

func (r *Renderer) renderThematicBreak(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		breakChars := []byte{'-', '*', '_'}
		breakChar := breakChars[r.config.ThematicBreakStyle : r.config.ThematicBreakStyle+1]
		var breakLen int
		if r.config.ThematicBreakLength < ThematicBreakLengthMinimum {
			breakLen = int(ThematicBreakLengthMinimum)
		} else {
			breakLen = int(r.config.ThematicBreakLength)
		}
		r.rc.writer.Write(bytes.Repeat(breakChar, breakLen))
	}
	return ast.WalkContinue
}

func (r *Renderer) renderCodeBlock(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.rc.writer.PushPrefix(r.config.IndentStyle.Bytes())
		r.renderLines(node, entering)
	} else {
		r.rc.writer.PopPrefix()
	}
	return ast.WalkContinue
}

func (r *Renderer) renderFencedCodeBlock(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.FencedCodeBlock)
	r.rc.writer.Write([]byte("```"))
	if entering {
		if info := n.Info; info != nil {
			r.rc.writer.Write(info.Text(r.rc.source))
		}
		r.rc.writer.FlushLine()
		r.renderLines(node, entering)
	}
	return ast.WalkContinue
}

func (r *Renderer) renderHTMLBlock(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.HTMLBlock)
	if entering {
		r.renderLines(node, entering)
	} else {
		if n.HasClosure() {
			r.rc.writer.WriteLine(n.ClosureLine.Value(r.rc.source))
		}
	}
	return ast.WalkContinue
}

func (r *Renderer) renderList(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		n := node.(*ast.List)
		r.rc.lists = append(r.rc.lists, listContext{
			list: n,
			num:  n.Start,
		})
	} else {
		r.rc.lists = r.rc.lists[:len(r.rc.lists)-1]
	}
	return ast.WalkContinue
}

func (r *Renderer) renderListItem(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		var itemPrefix []byte
		l := r.rc.lists[len(r.rc.lists)-1]

		if l.list.IsOrdered() {
			itemPrefix = append(itemPrefix, []byte(fmt.Sprint(l.num))...)
			r.rc.lists[len(r.rc.lists)-1].num += 1
		}
		itemPrefix = append(itemPrefix, l.list.Marker, ' ')
		// Prefix the current line with the item prefix
		r.rc.writer.PushPrefix(itemPrefix, 0, 0)
		// Prefix subsequent lines with padding the same length as the item prefix
		r.rc.writer.PushPrefix(bytes.Repeat([]byte(" "), len(itemPrefix)), 1)
	} else {
		r.rc.writer.PopPrefix()
		r.rc.writer.PopPrefix()
	}
	return ast.WalkContinue
}

func (r *Renderer) renderRawHTML(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.RawHTML)
	if entering {
		r.renderSegments(n.Segments, false)
	}
	return ast.WalkContinue
}

func (r *Renderer) renderText(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.Text)
	if entering {
		text := n.Text(r.rc.source)

		r.rc.writer.Write(text)
		if n.SoftLineBreak() {
			r.rc.writer.EndLine()
		}
	}
	return ast.WalkContinue
}

func (r *Renderer) renderSegments(segments *text.Segments, asLines bool) {
	for i := 0; i < segments.Len(); i++ {
		segment := segments.At(i)
		value := segment.Value(r.rc.source)
		r.rc.writer.Write(value)
		if asLines {
			r.rc.writer.FlushLine()
		}
	}
}

func (r *Renderer) renderLines(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		lines := node.Lines()
		r.renderSegments(lines, true)
	}
	return ast.WalkContinue
}

func (r *Renderer) renderLink(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.Link)
	return r.renderLinkCommon(n.Title, n.Destination, entering)
}

func (r *Renderer) renderImage(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.Image)
	if entering {
		r.rc.writer.Write([]byte("!"))
	}
	return r.renderLinkCommon(n.Title, n.Destination, entering)
}

func (r *Renderer) renderLinkCommon(title, destination []byte, entering bool) ast.WalkStatus {
	if entering {
		r.rc.writer.Write([]byte("["))
	} else {
		r.rc.writer.Write([]byte("]("))
		r.rc.writer.Write(destination)
		if len(title) > 0 {
			r.rc.writer.Write([]byte(" \""))
			r.rc.writer.Write(title)
			r.rc.writer.Write([]byte("\""))
		}
		r.rc.writer.Write([]byte(")"))
	}
	return ast.WalkContinue
}

func (r *Renderer) renderCodeSpan(node ast.Node, entering bool) ast.WalkStatus {
	if bytes.Count(node.Text(r.rc.source), []byte("`"))%2 != 0 {
		r.rc.writer.Write([]byte("``"))
	} else {
		r.rc.writer.Write([]byte("`"))
	}

	return ast.WalkContinue
}

func (r *Renderer) renderEmphasis(node ast.Node, entering bool) ast.WalkStatus {
	n := node.(*ast.Emphasis)
	r.rc.writer.Write(bytes.Repeat([]byte{'*'}, n.Level))
	return ast.WalkContinue
}

type renderContext struct {
	writer *markdownWriter
	// source is the markdown source
	source []byte
	// listMarkers is the marker character used for the current list
	lists []listContext
}

type listContext struct {
	list *ast.List
	num  int
}

// newRenderContext returns a new renderContext object
func newRenderContext(writer io.Writer, source []byte, config *Config) renderContext {
	return renderContext{
		writer: newMarkdownWriter(writer, config),
		source: source,
	}
}
