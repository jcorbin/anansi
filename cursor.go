package anansi

import (
	"io"
	"syscall"

	"github.com/jcorbin/anansi/ansi"
)

// Cursor supports writing buffered ansi output while tracking cursor state.
// Buffered output can be flushed with WriteTo(), or discarded with Reset().
// Real cursor state is only affected after a WriteTo(), and is restored after
// a Reset().
type Cursor struct {
	CursorState
	Real CursorState

	buf Buffer
}

// Reset the internal buffer and restore cursor state to last state affected by
// WriteTo.
func (c *Cursor) Reset() {
	c.CursorState = c.Real
	c.buf.Reset()
}

// WriteTo writes all bytes from the internal buffer to the given io.Writer. If
// that succeeds, then the current CursorState is set to the Real cursor state;
// otherwise, CursorState and Real are both zeroed.
func (c *Cursor) WriteTo(w io.Writer) (n int64, err error) {
	n, err = c.buf.WriteTo(w)
	if unwrapOSError(err) == syscall.EWOULDBLOCK {
		c.Real = CursorState{}
	} else if err != nil {
		c.Real = CursorState{}
		c.Reset()
	} else {
		c.Real = c.CursorState
	}
	return n, err
}

// Write to the internal buffer, updating cursor state per any ANSI escape
// sequences, and advancing cursor position by rune count (clamped to screen
// size).
func (c *Cursor) Write(p []byte) (n int, err error) {
	n, _ = c.buf.Write(p)
	c.buf.Process(c)
	return n, nil
}

// WriteString to the internal buffer, updating cursor state per any ANSI
// escape sequences, and advancing cursor position by rune count (clamped to
// screen size).
func (c *Cursor) WriteString(s string) (n int, err error) {
	n, _ = c.buf.WriteString(s)
	c.buf.Process(c)
	return n, nil
}

// WriteRune to the internal buffer, advancing cursor position (clamped to
// screen size).
func (c *Cursor) WriteRune(r rune) (n int, err error) {
	n, _ = c.buf.WriteRune(r)
	c.buf.Process(c)
	return n, nil
}

// WriteByte to the internal buffer, advancing cursor position (clamped to
// screen size).
func (c *Cursor) WriteByte(b byte) error {
	_ = c.buf.WriteByte(b)
	c.buf.Process(c)
	return nil
}

// WriteESC writes one or more ANSI escapes to the internal buffer, returning
// the number of bytes written; updates cursor state as appropriate.
func (c *Cursor) WriteESC(seqs ...ansi.Escape) int {
	n := c.buf.WriteESC(seqs...)
	c.buf.Process(c)
	return n
}

// WriteSeq writes one or more ANSI escape sequences to the internal buffer,
// returning the number of bytes written; updates cursor state as appropriate.
func (c *Cursor) WriteSeq(seqs ...ansi.Seq) int {
	n := c.buf.WriteSeq(seqs...)
	c.buf.Process(c)
	return n
}

// WriteSGR writes one or more ANSI SGR sequences to the internal buffer,
// returning the number of bytes written; updates Attr cursor state.
func (c *Cursor) WriteSGR(attrs ...ansi.SGRAttr) (n int) {
	for i := range attrs {
		n += c.buf.WriteSGR(c.CursorState.MergeSGR(attrs[i]))
	}
	if n > 0 {
		c.buf.Skip(n)
	}
	return n
}

// To moves the cursor to the given point using absolute (ansi.CUP) or relative
// (ansi.{CUU,CUD,CUF,CUD}) if possible.
func (c *Cursor) To(pt ansi.Point) {
	c.buf.Skip(c.buf.WriteSeq(c.CursorState.To(pt)))
}

// Show ensures that the cursor is visible, writing the necessary control
// sequence into the internal buffer if this is a change.
func (c *Cursor) Show() {
	c.buf.Skip(c.buf.WriteSeq(c.CursorState.Show()))
}

// Hide ensures that the cursor is not visible, writing the necessary control
// sequence into the internal buffer if this is a change.
func (c *Cursor) Hide() {
	c.buf.Skip(c.buf.WriteSeq(c.CursorState.Hide()))
}

// Apply the given cursor state, writing any necessary escape sequences into
// the internal buffer.
func (c *Cursor) Apply(cs CursorState) {
	_, c.CursorState = cs.applyTo(&c.buf, c.CursorState)
}

var _ ansiWriter = &Cursor{}
