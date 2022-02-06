package markdown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// TestRendererOptions tests the methods for setting configuration options on the renderer
func TestRendererOptions(t *testing.T) {
	var cases = []struct {
		name     string
		options  []Option
		expected Config
	}{
		{
			"Defaults",
			[]Option{},
			NewConfig(),
		},
		{
			"Zero values",
			[]Option{},
			Config{},
		},
		{
			"Explicit defaults",
			[]Option{WithIndentStyle(IndentStyleSpaces), WithHeadingStyle(HeadingStyleATX)},
			NewConfig(),
		},
		{
			"Tab indent",
			[]Option{WithIndentStyle(IndentStyleTabs)},
			Config{IndentStyle: IndentStyle(IndentStyleTabs)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Set options by passing them directly to NewNodeRenderer
			node_renderer := NewNodeRenderer(tc.options...).(*Renderer)
			assert.Equal(tc.expected, node_renderer.Config)

			// Set options by name using goldmark's renderer.AddOptions
			node_renderer = NewNodeRenderer().(*Renderer)
			r := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(node_renderer, 1000)))
			// Convert markdown Option interface to renderer.Option interface
			renderer_options := make([]renderer.Option, len(tc.options), len(tc.options))
			for i, o := range tc.options {
				renderer_options[i] = o
			}
			r.AddOptions(renderer_options...)
			// Must call Render() to apply options
			r.Render(&bytes.Buffer{}, []byte{}, ast.NewDocument())
			assert.Equal(tc.expected, node_renderer.Config)
		})
	}
}
