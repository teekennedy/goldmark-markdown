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
			nodeRenderer := NewNodeRenderer(tc.options...).(*Renderer)
			assert.Equal(tc.expected, nodeRenderer.Config)

			// Set options by name using goldmark's renderer.AddOptions
			nodeRenderer = NewNodeRenderer().(*Renderer)
			r := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(nodeRenderer, 1000)))
			// Convert markdown Option interface to renderer.Option interface
			rendererOptions := make([]renderer.Option, len(tc.options))
			for i, o := range tc.options {
				rendererOptions[i] = o
			}
			r.AddOptions(rendererOptions...)
			// Must call Render() to apply options
			err := r.Render(&bytes.Buffer{}, []byte{}, ast.NewDocument())
			assert.NoError(err)
			assert.Equal(tc.expected, nodeRenderer.Config)
		})
	}
}
