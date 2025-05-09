package markdown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.abhg.dev/goldmark/toc"
)

// TestTOCExtension tests compatibility with the goldmark-toc extension
func TestTOCExtension(t *testing.T) {
	assert := assert.New(t)
	md := goldmark.New(
		goldmark.WithExtensions(NewExtension(), &toc.Extender{}),
	)
	buf := bytes.Buffer{}
	source := `# 1
## 2
### 3
## 2
### 3
`

	err := md.Convert([]byte(source), &buf)
	assert.NoError(err)
	actual := buf.String()
	assert.Contains(actual, source, "Original markdown text should be retained")
	assert.Greater(actual, source, "Converted text should be longer")
}

func TestTypographerExtensionDisabled(t *testing.T) {
	assert := assert.New(t)
	md := goldmark.New(
		goldmark.WithExtensions(NewExtension(), extension.NewTypographer()),
	)
	buf := bytes.Buffer{}
	source := `'LeftSingleQuote
RightSingleQuote'
"LeftDoubleQuote
RightDoubleQuote"
EnDash --
EmDash ---
Ellipsis ...
LeftAngleQuote <<
RightAngleQuote >>
Apostrophe 'twas
`

	err := md.Convert([]byte(source), &buf)
	assert.NoError(err)
	assert.Equal(source, buf.String())
}

func TestTypographerExtensionEnabled(t *testing.T) {
	assert := assert.New(t)
	md := goldmark.New(
		goldmark.WithExtensions(NewExtension(WithTypographerSubstitutions(true)), extension.NewTypographer()),
	)
	buf := bytes.Buffer{}
	source := `'LeftSingleQuote
RightSingleQuote'
"LeftDoubleQuote
RightDoubleQuote"
EnDash --
EmDash ---
Ellipsis ...
LeftAngleQuote <<
RightAngleQuote >>
Apostrophe 'twas
`
	expected := `‘LeftSingleQuote
RightSingleQuote’
“LeftDoubleQuote
RightDoubleQuote”
EnDash –
EmDash —
Ellipsis …
LeftAngleQuote «
RightAngleQuote »
Apostrophe 'twas
`

	err := md.Convert([]byte(source), &buf)
	assert.NoError(err)
	assert.Equal(expected, buf.String())
}
