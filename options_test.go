package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/renderer"
)

// TestRendererOptions tests the methods for setting configuration options on the renderer
func TestRendererOptions(t *testing.T) {
	cases := []struct {
		name     string
		options  []Option
		expected *Config
	}{
		{
			"Defaults",
			[]Option{},
			NewConfig(),
		},
		{
			"Explicit defaults",
			[]Option{
				WithIndentStyle(IndentStyleSpaces),
				WithHeadingStyle(HeadingStyleATX),
				WithThematicBreakStyle(ThematicBreakStyleDashed),
				WithThematicBreakLength(ThematicBreakLengthMinimum),
				WithNestedListLength(NestedListLengthMinimum),
			},
			NewConfig(),
		},
		{
			"Tab indent",
			[]Option{WithIndentStyle(IndentStyleTabs)},
			NewConfig(WithIndentStyle(IndentStyleTabs)),
		},
		{
			"Underlined thematic breaks",
			[]Option{WithThematicBreakStyle(ThematicBreakStyleUnderlined)},
			NewConfig(WithThematicBreakStyle(ThematicBreakStyleUnderlined)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Set options by passing them directly to NewRenderer
			r := NewRenderer(tc.options...)
			assert.Equal(tc.expected, r.config)

			// Set options by name using AddOptions
			r = NewRenderer()
			// Convert markdown Option interface to renderer.Option interface
			options := make([]renderer.Option, len(tc.options))
			for i, o := range tc.options {
				options[i] = o
			}
			r.AddOptions(options...)
			assert.Equal(tc.expected, r.config)
		})
	}
}
