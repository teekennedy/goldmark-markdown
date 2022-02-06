package markdown

import "github.com/yuin/goldmark/renderer"

// Config struct holds configurations for the markdown based renderer.
type Config struct {
	IndentStyle  IndentStyle
	HeadingStyle HeadingStyle
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		IndentStyle:  IndentStyle(IndentStyleSpaces),
		HeadingStyle: HeadingStyle(HeadingStyleATX),
	}
}

// SetOption implements renderer.SetOptioner.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case optIndentStyle:
		c.IndentStyle = value.(IndentStyle)
	case optHeadingStyle:
		c.HeadingStyle = value.(HeadingStyle)
	}
}

// Option is an interface that sets options for Markdown based renderers.
type Option interface {
	renderer.Option
	// SetMarkDownOption sets this option on the markdown renderer config
	SetMarkdownOption(*Config)
}

// ============================================================================
// IndentStyle Option
// ============================================================================

// optIndentStyle is an option name used in WithIndentStyle
const optIndentStyle renderer.OptionName = "IndentStyle"

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

// SetConfig implements renderer.Option.SetConfig
func (o *withIndentStyle) SetConfig(c *renderer.Config) {
	c.Options[optIndentStyle] = o.value
}

// SetMarkdownOption implements Option
func (o *withIndentStyle) SetMarkdownOption(c *Config) {
	c.IndentStyle = o.value
}

// WithIndentStyle is a functional option that sets the string used to indent
// markdown blocks.
func WithIndentStyle(style IndentStyle) interface {
	renderer.Option
	Option
} {
	return &withIndentStyle{style}
}

// ============================================================================
// HeadingStyle Option
// ============================================================================

// optHeadingStyle is an option name used in WithHeadingStyle
const optHeadingStyle renderer.OptionName = "HeadingStyle"

// HeadingStyle is an enum expressing how markdown headings should look.
type HeadingStyle int

const (
	// HeadingStyleATX is the #-based style. This is the default heading style.
	HeadingStyleATX = iota
	// HeadingStyleATXSurround adds closing #s after your header. Ex: '## Foo ##'.
	HeadingStyleATXSurround
	// HeadingStyleSetext uses setext heading underlines ('===' or '---') for heading levels 1 and
	// 2, respectively. Other header levels continue to use ATX headings.
	HeadingStyleSetext
	// HeadingStyleFullWidthSetext extends setext heading underlines to the full width of the
	// header text.
	HeadingStyleFullWidthSetext
)

// IsSetext returns true if heading style is one of the Setext options
func (i HeadingStyle) IsSetext() bool {
	return i == HeadingStyleSetext || i == HeadingStyleFullWidthSetext
}

type withHeadingStyle struct {
	value HeadingStyle
}

func (o *withHeadingStyle) SetConfig(c *renderer.Config) {
	c.Options[optHeadingStyle] = o.value
}

// SetMarkdownOption implements renderer.Option
func (o *withHeadingStyle) SetMarkdownOption(c *Config) {
	c.HeadingStyle = o.value
}

// WithHeadingStyle is a functional option that sets the string used to indent
// markdown blocks.
func WithHeadingStyle(style HeadingStyle) interface {
	renderer.Option
	Option
} {
	return &withHeadingStyle{style}
}
