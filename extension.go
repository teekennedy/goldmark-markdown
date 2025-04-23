package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type rendererExtension struct{ opts []Option }

// NewExtension returns a new goldmark.Markdown extension that uses the markdown renderer.
func NewExtension(opts ...Option) *rendererExtension {
	return &rendererExtension{opts: opts}
}

// Extend implements goldmark.Extension.Extend
func (re *rendererExtension) Extend(md goldmark.Markdown) {
	renderer := NewRenderer(re.opts...)
	md.SetRenderer(renderer)
	if renderer.config.TypographerSubstitutions {
		enableTypographicSubstitutions(md)
	} else {
		disableTypographicSubstitutions(md)
	}
}

// enableTypographicSubstitutions configures the typographer extension
// to substitute punctuations with the corresponding unicode character,
// instead of the default HTML escape sequence.
// These substitutions are only applied IF the typographer extension is enabled.
func enableTypographicSubstitutions(md goldmark.Markdown) {
	subs := make(extension.TypographicSubstitutions)
	subs[extension.LeftSingleQuote] = []byte("‘")
	subs[extension.RightSingleQuote] = []byte("’")
	subs[extension.LeftDoubleQuote] = []byte("“")
	subs[extension.RightDoubleQuote] = []byte("”")
	subs[extension.EnDash] = []byte("–")
	subs[extension.EmDash] = []byte("—")
	subs[extension.Ellipsis] = []byte("…")
	subs[extension.LeftAngleQuote] = []byte("«")
	subs[extension.RightAngleQuote] = []byte("»")
	subs[extension.Apostrophe] = []byte("'")

	md.Parser().AddOptions(extension.WithTypographicSubstitutions(subs))
}

// disableTypographicSubstitutions configures the typographer extension
// to substitute punctuations with their original values,
// effectively turning typographer into a no-op.
func disableTypographicSubstitutions(md goldmark.Markdown) {
	subs := make(extension.TypographicSubstitutions)
	subs[extension.LeftSingleQuote] = []byte("'")
	subs[extension.RightSingleQuote] = []byte("'")
	subs[extension.LeftDoubleQuote] = []byte("\"")
	subs[extension.RightDoubleQuote] = []byte("\"")
	subs[extension.EnDash] = []byte("--")
	subs[extension.EmDash] = []byte("---")
	subs[extension.Ellipsis] = []byte("...")
	subs[extension.LeftAngleQuote] = []byte("<<")
	subs[extension.RightAngleQuote] = []byte(">>")
	subs[extension.Apostrophe] = []byte("'")

	md.Parser().AddOptions(extension.WithTypographicSubstitutions(subs))
}
