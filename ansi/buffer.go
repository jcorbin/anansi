package ansi

import (
	"bytes"
	"io"
)

// Buffer implements a deferred buffer of ANSI output, providing
// convenience methods for writing various ansi escape sequences, and keeping
// an observant processor up to date.
type Buffer struct {
	buf bytes.Buffer
}

// Len returns the number of unwritten bytes in the buffer.
func (b *Buffer) Len() int {
	return b.buf.Len()
}

// Grow the internal buffer to have room for at least n bytes.
func (b *Buffer) Grow(n int) {
	b.buf.Grow(n)
}

// Reset the internal buffer.
func (b *Buffer) Reset() {
	b.buf.Reset()
}

// WriteTo writes all bytes from the internal buffer to the given io.Writer.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	n, err = b.buf.WriteTo(w)
	return n, err
}

// WriteESC writes one or more ANSI escapes to the internal buffer, returning
// the number of bytes written.
func (b *Buffer) WriteESC(seqs ...Escape) int {
	need := 0
	for i := range seqs {
		need += seqs[i].Size()
	}
	b.buf.Grow(need)
	p := b.buf.Bytes()
	p = p[len(p):]
	for i := range seqs {
		p = seqs[i].AppendTo(p)
	}
	n, _ := b.buf.Write(p)
	return n
}

// WriteSeq writes one or more ANSI escape sequences to the internal buffer,
// returning the number of bytes written. Skips any zero sequences provided.
func (b *Buffer) WriteSeq(seqs ...Seq) int {
	need := 0
	for i := range seqs {
		if seqs[i].id != 0 {
			need += seqs[i].Size()
		}
	}
	b.buf.Grow(need)
	p := b.buf.Bytes()
	p = p[len(p):]
	for i := range seqs {
		if seqs[i].id != 0 {
			p = seqs[i].AppendTo(p)
		}
	}
	n, _ := b.buf.Write(p)
	return n
}

// WriteSGR writes one or more ANSI SGR sequences to the internal buffer,
// returning the number of bytes written; updates Attr cursor state. Skips any
// zero attr values (NOTE 0 attr value is merely implicit clear, not the
// explicit SGRAttrClear).
func (b *Buffer) WriteSGR(attrs ...SGRAttr) int {
	need := 0
	for i := range attrs {
		if attrs[i] != 0 {
			need += attrs[i].Size()
		}
	}
	b.buf.Grow(need)
	p := b.buf.Bytes()
	p = p[len(p):]
	for i := range attrs {
		if attrs[i] != 0 {
			p = attrs[i].AppendTo(p)
		}
	}
	n, _ := b.buf.Write(p)
	return n
}

// Write to the internal buffer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

// WriteString to the internal buffer.
func (b *Buffer) WriteString(s string) (n int, err error) {
	return b.buf.WriteString(s)
}

// WriteRune to the internal buffer.
func (b *Buffer) WriteRune(r rune) (n int, err error) {
	return b.buf.WriteRune(r)
}

// WriteByte to the internal buffer.
func (b *Buffer) WriteByte(c byte) error {
	return b.buf.WriteByte(c)
}
