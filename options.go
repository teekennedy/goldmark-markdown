package markdown

import "github.com/yuin/goldmark/renderer"

// Config struct holds configurations for the markdown based renderer.
type Config struct {
	IndentStyle         IndentStyle
	HeadingStyle        HeadingStyle
	ThematicBreakStyle  ThematicBreakStyle
	ThematicBreakLength ThematicBreakLength
}

// NewConfig returns a new Config with defaults and the given options.
func NewConfig(options ...Option) Config {
	c := Config{
		IndentStyle:         IndentStyle(IndentStyleSpaces),
		HeadingStyle:        HeadingStyle(HeadingStyleATX),
		ThematicBreakStyle:  ThematicBreakStyle(ThematicBreakStyleDashed),
		ThematicBreakLength: ThematicBreakLength(ThematicBreakLengthMinimum),
	}
	for _, opt := range options {
		opt.SetMarkdownOption(&c)
	}
	return c
}

// SetOption implements renderer.SetOptioner.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case optIndentStyle:
		c.IndentStyle = value.(IndentStyle)
	case optHeadingStyle:
		c.HeadingStyle = value.(HeadingStyle)
	case optThematicBreakStyle:
		c.ThematicBreakStyle = value.(ThematicBreakStyle)
	case optThematicBreakLength:
		c.ThematicBreakLength = value.(ThematicBreakLength)
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

// WithIndentStyle is a functional option that sets the string used to indent markdown blocks.
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
	// HeadingStyleATX is the #-based style. This is the default heading style and zero value.
	// Ex: ## Foo
	HeadingStyleATX = iota
	// HeadingStyleATXSurround adds closing #s after your header.
	// Ex: ## Foo ##
	HeadingStyleATXSurround
	// HeadingStyleSetext uses setext heading underlines ('===' or '---') for heading levels 1 and
	// 2, respectively. Other header levels continue to use ATX headings.
	// Ex: Foo Bar
	//     ---
	HeadingStyleSetext
	// HeadingStyleFullWidthSetext extends setext heading underlines to the full width of the
	// header text.
	// Ex: Foo Bar
	//     -------
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

// WithHeadingStyle is a functional option that sets the style of markdown headings.
func WithHeadingStyle(style HeadingStyle) interface {
	renderer.Option
	Option
} {
	return &withHeadingStyle{style}
}

// ============================================================================
// ThematicBreakStyle Option
// ============================================================================

// optThematicBreakStyle is an option name used in WithThematicBreakStyle
const optThematicBreakStyle renderer.OptionName = "ThematicBreakStyle"

// ThematicBreakStyle is an enum expressing the character used for thematic breaks.
type ThematicBreakStyle int

const (
	// ThematicBreakStyleDashed uses '-' character for thematic breaks. This is the default and
	// zero value.
	// Ex: ---
	ThematicBreakStyleDashed = iota
	// ThematicBreakStyleStarred uses '*' character for thematic breaks.
	// Ex: ***
	ThematicBreakStyleStarred
	// ThematicBreakStyleUnderlined uses '_' character for thematic breaks.
	// Ex: ___
	ThematicBreakStyleUnderlined
)

type withThematicBreakStyle struct {
	value ThematicBreakStyle
}

func (o *withThematicBreakStyle) SetConfig(c *renderer.Config) {
	c.Options[optThematicBreakStyle] = o.value
}

// SetMarkdownOption implements renderer.Option
func (o *withThematicBreakStyle) SetMarkdownOption(c *Config) {
	c.ThematicBreakStyle = o.value
}

// WithThematicBreakStyle is a functional option that sets the character used for thematic breaks.
func WithThematicBreakStyle(style ThematicBreakStyle) interface {
	renderer.Option
	Option
} {
	return &withThematicBreakStyle{style}
}

// ============================================================================
// ThematicBreakLength Option
// ============================================================================

// optThematicBreakLength is an option name used in WithThematicBreakLength
const optThematicBreakLength renderer.OptionName = "ThematicBreakLength"

// ThematicBreakLength configures the character length of thematic breaks
type ThematicBreakLength int

const (
	// ThematicBreakLengthMinimum is the minimum length of a thematic break. This is the default.
	// Any lengths less than this minimum are converted to the minimum.
	// Ex: ---
	ThematicBreakLengthMinimum = 3
)

type withThematicBreakLength struct {
	value ThematicBreakLength
}

func (o *withThematicBreakLength) SetConfig(c *renderer.Config) {
	c.Options[optThematicBreakLength] = o.value
}

// SetMarkdownOption implements renderer.Option
func (o *withThematicBreakLength) SetMarkdownOption(c *Config) {
	c.ThematicBreakLength = o.value
}

// WithThematicBreakLength is a functional option that sets the length of thematic breaks.
func WithThematicBreakLength(style ThematicBreakLength) interface {
	renderer.Option
	Option
} {
	return &withThematicBreakLength{style}
}
