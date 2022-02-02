package renderer

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Config struct holds configurations for the markdown based renderer.
type Config struct {
	HardWraps    bool
	IndentStyle  IndentStyle
	HeadingStyle HeadingStyle
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		HardWraps:    false,
		IndentStyle:  IndentStyle(IndentStyleSpaces),
		HeadingStyle: HeadingStyle(HeadingStyleATX),
	}
}

// SetOption implements renderer.NodeRenderer.SetOption.
func (c *Config) SetOption(name goldrenderer.OptionName, value interface{}) {
	switch name {
	case optHardWraps:
		c.HardWraps = value.(bool)
	case optIndentStyle:
		c.IndentStyle = value.(IndentStyle)
	case optHeadingStyle:
		c.HeadingStyle = value.(HeadingStyle)
	}
}

// Option is an interface that sets options for Markdown based renderers.
type Option interface {
	SetMarkdownOption(*Config)
}

// ============================================================================
// HardWraps Option
// ============================================================================

// optHardWraps is an option name used in WithHardWraps.
const optHardWraps goldrenderer.OptionName = "HardWraps"

type withHardWraps struct {
}

func (o *withHardWraps) SetConfig(c *goldrenderer.Config) {
	c.Options[optHardWraps] = true
}

func (o *withHardWraps) SetMarkdownOption(c *Config) {
	c.HardWraps = true
}

// WithHardWraps is a functional option that indicates whether soft line breaks
// should be rendered as hard line breaks.
func WithHardWraps() interface {
	goldrenderer.Option
	Option
} {
	return &withHardWraps{}
}

// ============================================================================
// IndentStyle Option
// ============================================================================

// optIndentStyle is an option name used in WithIndentStyle
const optIndentStyle goldrenderer.OptionName = "IndentStyle"

// IndentStyle is an enum expressing how markdown blocks should be indented.
type IndentStyle int

const (
	// IndentStyleSpaces indents with 4 spaces. This is the default as well as the zero-value.
	IndentStyleSpaces = iota
	// IndentStyleTabs indents with tabs.
	IndentStyleTabs
)

// bytes returns the raw bytes representation of the indent style
func (i IndentStyle) bytes() []byte {
	return []byte([...]string{"    ", "\t"}[i])
}

type withIndentStyle struct {
	value IndentStyle
}

func (o *withIndentStyle) SetConfig(c *goldrenderer.Config) {
	c.Options[optIndentStyle] = o.value
}

func (o *withIndentStyle) SetMarkdownOption(c *Config) {
	c.IndentStyle = o.value
}

// WithIndentStyle is a functional option that sets the string used to indent
// markdown blocks.
func WithIndentStyle(style IndentStyle) interface {
	goldrenderer.Option
	Option
} {
	return &withIndentStyle{style}
}

// ============================================================================
// HeadingStyle Option
// ============================================================================

// optHeadingStyle is an option name used in WithHeadingStyle
const optHeadingStyle goldrenderer.OptionName = "HeadingStyle"

// HeadingStyle is an enum expressing how markdown blocks should be indented.
type HeadingStyle int

const (
	// HeadingStyleATX is the #-based style. This is the default heading style.
	HeadingStyleATX = 1 << iota
	// HeadingStyleATXSurround adds closing #s after your header. Ex: '## Foo ##'.
	HeadingStyleATXSurround
	// HeadingStyleSetext uses setext heading underlines ('===' or '---') for heading levels 1 and
	// 2, respectively. Other header levels continue to use ATX headings.
	HeadingStyleSetext
	// HeadingStyleFullWidthSetext extends setext heading underlines to the full width of the
	// header text.
	HeadingStyleFullWidthSetext
)

// Has returns true if the given HeadingStyle is enabled.
func (i HeadingStyle) Has(style HeadingStyle) bool {
	return i&style != 0
}

// HasSetextEnabled returns true if any of the setext heading options are enabled.
func (i HeadingStyle) HasSetextEnabled() bool {
	var setextStyles HeadingStyle = HeadingStyleSetext | HeadingStyleFullWidthSetext
	return i.Has(setextStyles)
}

type withHeadingStyle struct {
	value HeadingStyle
}

func (o *withHeadingStyle) SetConfig(c *goldrenderer.Config) {
	c.Options[optHeadingStyle] = o.value
}

func (o *withHeadingStyle) SetMarkdownOption(c *Config) {
	c.HeadingStyle = o.value
}

// WithHeadingStyle is a functional option that sets the string used to indent
// markdown blocks.
func WithHeadingStyle(style HeadingStyle) interface {
	goldrenderer.Option
	Option
} {
	return &withHeadingStyle{style}
}

// NewRenderer returns a new Renderer that is configured by default values.
func NewRenderer(options ...Option) goldrenderer.Renderer {
	r := &Renderer{
		Config: NewConfig(),
	}
	for _, opt := range options {
		opt.SetMarkdownOption(&r.Config)
	}
	return goldrenderer.NewRenderer(goldrenderer.WithNodeRenderers(util.Prioritized(r, 1000)))
}

// The Renderer struct is an implementation of goldrenderer that renders nodes
// as Markdown
type Renderer struct {
	Config
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *Renderer) RegisterFuncs(reg goldrenderer.NodeRendererFuncRegisterer) {
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
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindText, r.renderText)
	/* TODO
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
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Heading)
	// Empty headers can only be ATX
	// Multiline headers can only be Setext
	// Headers above level 2 can only be ATX
	// Otherwise it's up to the configuration
	if !n.HasChildren() || n.Lines().Len() == 1 && (n.Level > 2 || !r.HeadingStyle.HasSetextEnabled()) {
		return r.renderATXHeading(w, source, node, entering)
	}
	return r.renderSetextHeading(w, source, node, entering)
}

func (r *Renderer) renderATXHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	atxHeadingChars := strings.Repeat("#", n.Level)
	fmt.Fprint(w, atxHeadingChars)
	// Only print space after header if there's more to print
	if n.HasChildren() {
		fmt.Fprintf(w, " %s", n.Text(source))
	}
	if r.HeadingStyle.Has(HeadingStyleATXSurround) {
		fmt.Fprintf(w, " %v", atxHeadingChars)
	}
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderSetextHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	lines := n.Lines()
	underlineChar := [...]string{"", "=", "-"}[n.Level]
	underlineWidth := 3
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		lineValue := line.Value(source)
		lineWidth := len(lineValue)

		if lineWidth > underlineWidth {
			underlineWidth = lineWidth
		}
		fmt.Fprintf(w, "%s", lineValue)
	}
	if !r.HeadingStyle.Has(HeadingStyleFullWidthSetext) {
		underlineWidth = 3
	}
	fmt.Fprintf(w, "\n%v", strings.Repeat(underlineChar, underlineWidth))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.String)
	if entering {
		_, _ = w.Write(n.Value)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Text)
	if entering {
		_, _ = w.Write(n.Segment.Value(source))
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
