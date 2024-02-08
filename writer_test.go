package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := newMarkdownWriter(buf, NewConfig())

	available := writer.Available()
	assert.Equal(t, 0, available)
	assert.Equal(t, 0, writer.Buffered())

	writer.WriteByte(byte('b'))
	writer.WriteRune('r')
	writer.WriteString("s")

	assert.Less(t, available, writer.Available())
	assert.Equal(t, 3, writer.Buffered())

	err := writer.Flush()
	require.NoError(t, err)
	assert.Equal(t, "brs\n", buf.String())
}

// TestFlushLine tests that the writer will flush the current buffered line if non-empty.
func TestFlushLine(t *testing.T) {
	assert := assert.New(t)
	buf := &bytes.Buffer{}
	writer := newMarkdownWriter(buf, NewConfig())

	writer.FlushLine()
	assert.Equal("", buf.String(), "FlushLine() on an empty buffer should not produce output")
	writer.Write([]byte("foobar"))
	writer.FlushLine()
	assert.Equal("foobar\n", buf.String(), "FlushLine() on partial line should produce output.")
}

// TestEndLine tests that the writer will end the current line whether empty or not.
func TestEndLine(t *testing.T) {
	assert := assert.New(t)
	buf := &bytes.Buffer{}
	writer := newMarkdownWriter(buf, NewConfig())

	writer.EndLine()
	assert.Equal("\n", buf.String(), "EndLine() should write newline to output")
	writer.Write([]byte("A line"))
	assert.Equal("\n", buf.String(), "Writing a partial line should not produce output.")
	writer.FlushLine()
	assert.Equal("\nA line\n", buf.String(), "FlushLine() on partial line should produce output.")
}

// TestWriterOutputs tests that the writer produces expected output in various scenarios.
func TestWriterOutputs(t *testing.T) {
	testCases := []struct {
		name      string
		writeFunc func(writer *markdownWriter)
		expected  string
	}{
		{
			"No prefix",
			func(w *markdownWriter) { w.WriteLine([]byte("foo")) },
			"foo\n",
		},
		{
			"Trailing whitespace",
			func(writer *markdownWriter) {
				writer.WriteLine([]byte("Line with trailing whitespace.\t \t "))
			},
			"Line with trailing whitespace.\n",
		},
		{
			"Prefix current line",
			func(writer *markdownWriter) {
				quotedLines := []string{"You will speak", "an infinite deal", "of nothing"}
				normalLine := "\\- William Shakespeare"
				writer.PushPrefix([]byte("> "))
				for _, line := range quotedLines {
					writer.WriteLine([]byte(line))
				}
				writer.PopPrefix()
				writer.WriteLine([]byte(normalLine))
			},
			`
> You will speak
> an infinite deal
> of nothing
\- William Shakespeare
`,
		},
		{
			"Prefix subsequent lines",
			func(writer *markdownWriter) {
				lines := []string{"Consider me", "As one who loved poetry", "And persimmons."}
				writer.PushPrefix([]byte("  "), 1)
				writer.Write([]byte("- "))
				for _, line := range lines {
					writer.WriteLine([]byte(line))
				}
				writer.PopPrefix()
				writer.WriteLine([]byte("\\- Masaoaka Shiki"))
			},
			`
- Consider me
  As one who loved poetry
  And persimmons.
\- Masaoaka Shiki
`,
		},
		{
			"Nested prefixes",
			func(writer *markdownWriter) {
				quotes := [][]string{
					{"You will speak", "an infinite deal", "of nothing"},
					{"Consider me", "As one who loved poetry", "And persimmons."},
				}
				authors := [][]byte{
					[]byte("\\- William Shakespeare"),
					[]byte("\\- Masaoaka Shiki"),
				}
				for i := range quotes {
					writer.PushPrefix([]byte("- "), 0, 0)
					writer.PushPrefix([]byte("  "), 1)
					writer.PushPrefix([]byte("> "))
					for _, line := range quotes[i] {
						writer.WriteLine([]byte(line))
					}
					writer.PopPrefix() // quote
					writer.WriteLine(authors[i])
					writer.PopPrefix() // list item subsequent lines
					writer.PopPrefix() // list item first line
				}
			},
			`
- > You will speak
  > an infinite deal
  > of nothing
  \- William Shakespeare
- > Consider me
  > As one who loved poetry
  > And persimmons.
  \- Masaoaka Shiki
`,
		},
	}
	output := bytes.Buffer{}
	mdWriter := newMarkdownWriter(&output, NewConfig())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.writeFunc(mdWriter)
			assert.NoError(t, mdWriter.err)
			assert.Equal(t, strings.TrimLeft(tc.expected, "\n"), output.String())
			output.Reset()
			mdWriter.Reset(&output)
		})
	}
}

type errorWriter struct {
	// err is the error to return
	err error
}

func (e *errorWriter) Write([]byte) (int, error) {
	return 0, e.err
}

// TestWriteError tests that the writer will turn all write operations into no-ops if the output
// writer returns an error
func TestWriteError(t *testing.T) {
	assert := assert.New(t)
	err := fmt.Errorf("test error")
	ew := errorWriter{}
	writer := newMarkdownWriter(&ew, NewConfig())
	data := []byte("foo\n")

	var n int
	n, _ = writer.Write(data)
	assert.Equal(len(data), n, "Writes should succeed before error")
	assert.Equal(len(data), writer.WriteLine(data), "Writes should succeed before error")
	ew.err = err
	n, _ = writer.Write(data)
	assert.Equal(0, n, "Once error is set, writes become no-op")
	assert.Equal(0, writer.WriteLine(data), "Once error is set, writes become no-op")
	assert.Equal(err, writer.Err(), "Err() should match error returned by errorWriter")

	ew.err = nil
	writer.Reset(&ew)
	n, _ = writer.Write(data)
	assert.Equal(len(data), n, "Writes should succeed after Reset")
	assert.Equal(len(data), writer.WriteLine(data), "Writes should succeed after Reset")
}
