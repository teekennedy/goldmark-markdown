package markdown

import (
	"bytes"
	"io"
	"unicode"
)

// Line delimiter
const lineDelim byte = '\n'

// linePrefix associates a line prefix with its starting line.
type linePrefix struct {
	// startLine and endLine are the line numbers the prefix starts and ends on
	startLine, endLine int
	// bytes is the bytes of the prefix
	bytes []byte
}

// markdownWriter provides an interface similar to io.Writer for writing markdown files. It handles
// errors returned by the underlying writer, and manages output of some rendering contexts like
// container block prefixes.
type markdownWriter struct {
	buf    *bytes.Buffer
	config *Config
	// output holds the underlying output writer
	output io.Writer
	// prefixes holds the current line prefixes
	prefixes []linePrefix
	// line is the current line number
	line int
	// err holds the last write error. If non-nil, all write operations become no-ops
	err error
}

// newMarkdownWriter returns a new markdownWriter
func newMarkdownWriter(w io.Writer, config *Config) *markdownWriter {
	result := &markdownWriter{
		config: config,
		buf:    &bytes.Buffer{},
	}
	// Reset initializes the rest of the struct
	result.Reset(w)
	return result
}

// Reset resets all internal state and switches writes to the given writer.
func (m *markdownWriter) Reset(w io.Writer) {
	m.buf.Reset()
	m.output = w
	m.prefixes = make([]linePrefix, 0)
	m.line = 0
	m.err = nil
}

// WriteLine writes the given bytes as a finished line, regardless of trailing newline.
func (m *markdownWriter) WriteLine(line []byte) (n int) {
	n = m.Write(line)
	m.FlushLine()

	return n
}

// FlushLine ends the current buffered line if non-empty, flushing the contents to the underlying
// writer.
func (m *markdownWriter) FlushLine() {
	if m.buf.Len() > 0 {
		m.EndLine()
	}
}

// EndLine ends the current line, flushing the line buffer regardless of whether it's empty.
func (m *markdownWriter) EndLine() {
	m.Write([]byte{lineDelim})
}

// PushPrefix adds the given bytes as a prefix for lines written to the output. The prefix
// will be added to the current line and all subsequent lines by default, but can optionally be
// given a start line relative to the current line, and an end line relative to the start line.
func (p *markdownWriter) PushPrefix(bytes []byte, lineRanges ...int) {
	prefix := linePrefix{
		endLine: -1,
		bytes:   bytes,
	}
	if len(lineRanges) > 0 {
		prefix.startLine = p.line + lineRanges[0]
		if len(lineRanges) > 1 {
			prefix.endLine = prefix.startLine + lineRanges[1]
		}
	}
	p.prefixes = append(p.prefixes, prefix)
}

// PopPrefix removes the most recently pushed line prefix from future lines.
func (p *markdownWriter) PopPrefix() {
	p.prefixes = p.prefixes[0 : len(p.prefixes)-1]
}

// Write writes the given data to an internal buffer, then writes any complete lines to the
// underlying writer.
func (m *markdownWriter) Write(data []byte) (n int) {
	if m.err != nil {
		return 0
	}
	// Writing to a bytes.Buffer always returns a nil error
	n, _ = m.buf.Write(data)
	prefixedLine := bytes.Buffer{}
	for bytes.Contains(m.buf.Bytes(), []byte{lineDelim}) {
		// err will only be non-nil if lineDelim is not in m.buf, which we already checked for.
		line, _ := m.buf.ReadBytes(lineDelim)
		// build the prefix for the line
		for _, prefix := range m.prefixes {
			if prefix.startLine <= m.line && (prefix.endLine == -1 || m.line <= prefix.endLine) {
				prefixedLine.Write(prefix.bytes)
			}
		}
		prefixedLine.Write(line)
		// trim whitespace off the end of the line
		trimmedSlice := bytes.TrimRightFunc(prefixedLine.Bytes(), unicode.IsSpace)
		prefixedLine.Truncate(len(trimmedSlice))
		prefixedLine.WriteByte(lineDelim)

		_, err := m.output.Write(prefixedLine.Bytes())
		if err != nil {
			m.err = err
			return 0
		}
		m.line += 1
		prefixedLine.Reset()
	}
	return n
}

// Write writes the given data to an internal buffer, then writes any complete lines to the
// underlying writer.
// TODO reduce copying data by making all write operations byte based and deleting this method.
func (m *markdownWriter) WriteString(data string) int {
	n := 0
	if m.err == nil {
		n, _ = io.WriteString(m.buf, data)
	}
	return n
}

// Err returns the last write error, or nil.
func (m *markdownWriter) Err() error {
	return m.err
}
