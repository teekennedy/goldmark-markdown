// Package markdown is a goldmark renderer that outputs markdown.
package markdown

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
)

// NewRenderer returns a new markdown Renderer that is configured by default values.
func NewRenderer(options ...Option) *Renderer {
	r := &Renderer{
		config:               NewConfig(),
		rc:                   renderContext{},
		maxKind:              20, // a random number slightly larger than the number of default ast kinds
		nodeRendererFuncsTmp: map[ast.NodeKind]renderer.NodeRendererFunc{},
	}
	for _, opt := range options {
		opt.SetMarkdownOption(r.config)
	}
	return r
}

// Renderer is an implementation of renderer.Renderer that renders nodes as Markdown
type Renderer struct {
	config               *Config
	rc                   renderContext
	nodeRendererFuncsTmp map[ast.NodeKind]renderer.NodeRendererFunc
	maxKind              int
	nodeRendererFuncs    []nodeRenderer
	initSync             sync.Once
}

var _ renderer.Renderer = &Renderer{}

// AddOptions implements renderer.Renderer.AddOptions
func (r *Renderer) AddOptions(opts ...renderer.Option) {
	config := renderer.NewConfig()
	for _, opt := range opts {
		opt.SetConfig(config)
	}
	for name, value := range config.Options {
		r.config.SetOption(name, value)
	}

	// handle any config.NodeRenderers set by opts
	config.NodeRenderers.Sort()
	l := len(config.NodeRenderers)
	for i := l - 1; i >= 0; i-- {
		v := config.NodeRenderers[i]
		nr, _ := v.Value.(renderer.NodeRenderer)
		nr.RegisterFuncs(r)
	}
}

func (r *Renderer) Register(kind ast.NodeKind, fun renderer.NodeRendererFunc) {
	r.nodeRendererFuncsTmp[kind] = fun
	if int(kind) > r.maxKind {
		r.maxKind = int(kind)
	}
}

// Render implements renderer.Renderer.Render
func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.rc = newRenderContext(w, source, r.config)
	r.initSync.Do(func() {
		r.nodeRendererFuncs = make([]nodeRenderer, r.maxKind+1)
		// add default functions
		// blocks
		r.nodeRendererFuncs[ast.KindDocument] = r.renderBlockSeparator
		r.nodeRendererFuncs[ast.KindHeading] = r.chainRenderers(r.renderBlockSeparator, r.renderHeading)
		r.nodeRendererFuncs[ast.KindBlockquote] = r.chainRenderers(r.renderBlockSeparator, r.renderBlockquote)
		r.nodeRendererFuncs[ast.KindCodeBlock] = r.chainRenderers(r.renderBlockSeparator, r.renderCodeBlock)
		r.nodeRendererFuncs[ast.KindFencedCodeBlock] = r.chainRenderers(r.renderBlockSeparator, r.renderFencedCodeBlock)
		r.nodeRendererFuncs[ast.KindHTMLBlock] = r.chainRenderers(r.renderBlockSeparator, r.renderHTMLBlock)
		r.nodeRendererFuncs[ast.KindList] = r.chainRenderers(r.renderBlockSeparator, r.renderList)
		r.nodeRendererFuncs[ast.KindListItem] = r.chainRenderers(r.renderBlockSeparator, r.renderListItem)
		r.nodeRendererFuncs[ast.KindParagraph] = r.renderBlockSeparator
		r.nodeRendererFuncs[ast.KindTextBlock] = r.renderBlockSeparator
		r.nodeRendererFuncs[ast.KindThematicBreak] = r.chainRenderers(r.renderBlockSeparator, r.renderThematicBreak)

		// inlines
		r.nodeRendererFuncs[ast.KindAutoLink] = r.renderAutoLink
		r.nodeRendererFuncs[ast.KindCodeSpan] = r.renderCodeSpan
		r.nodeRendererFuncs[ast.KindEmphasis] = r.renderEmphasis
		r.nodeRendererFuncs[ast.KindImage] = r.renderImage
		r.nodeRendererFuncs[ast.KindLink] = r.renderLink
		r.nodeRendererFuncs[ast.KindRawHTML] = r.renderRawHTML
		r.nodeRendererFuncs[ast.KindText] = r.renderText
		// TODO: add KindString
		// r.nodeRendererFuncs[ast.KindString] = r.renderString

		for kind, fun := range r.nodeRendererFuncsTmp {
			r.nodeRendererFuncs[kind] = r.transform(fun)
		}
		r.nodeRendererFuncsTmp = nil
	})
	return ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return r.nodeRendererFuncs[n.Kind()](n, entering), r.rc.writer.Err()
	})
}

func (r *Renderer) transform(fn renderer.NodeRendererFunc) nodeRenderer {
	return func(n ast.Node, entering bool) ast.WalkStatus {
		status, _ := fn(r.rc.writer, r.rc.source, n, entering)
		return status
	}
}

// nodeRenderer is a markdown node renderer func.
type nodeRenderer func(ast.Node, bool) ast.WalkStatus

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
