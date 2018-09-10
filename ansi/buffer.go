package ansi

import (
	"bytes"
	"io"
	"unicode/utf8"
)

// Buffer implements a deferred buffer of ANSI output, providing
// convenience methods for writing various ansi escape sequences, and keeping
// an observant processor up to date.
type Buffer struct {
	buf bytes.Buffer
	off int
}

// Len returns the number of unwritten bytes in the buffer.
func (b *Buffer) Len() int {
	return b.buf.Len()
}

// Grow the internal buffer to have room for at least n bytes.
func (b *Buffer) Grow(n int) {
	b.buf.Grow(n)
}

// Bytes returns a byte slice containing all bytes written into the internal
// buffer. Returned slice is only valid until the next call to a buffer method.
func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}

// Reset the internal buffer.
func (b *Buffer) Reset() {
	b.buf.Reset()
	b.off = 0
}

// WriteTo writes all bytes from the internal buffer to the given io.Writer.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	n, err = b.buf.WriteTo(w)
	b.off -= int(n)
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

// Skip Process()ing of n bytes written to the internal buffer. Useful when the
// processor wants to intermediate a buffer write, handling its own semantic
// update and avoiding (re)parsing the written bytes.
func (b *Buffer) Skip(n int) {
	b.off += n
}

// Discard processed bytes, re-using internal buffer space during the next Write*.
func (b *Buffer) Discard() {
	if b.off > 0 {
		b.buf.Next(b.off)
		b.off = 0
	}
}

// Process bytes written to the internal buffer, decoding runes and escape
// sequences, and passing them to the given processor.
func (b *Buffer) Process(proc Processor) {
	for p := b.buf.Bytes(); b.off < len(p); {
		e, a, n := DecodeEscape(p[b.off:])
		b.off += n
		if e == 0 {
			switch r, n := utf8.DecodeRune(p[b.off:]); r {
			case '\x1b':
				return
			default:
				b.off += n
				proc.ProcessRune(r)
			}
		} else {
			proc.ProcessEscape(e, a)
		}
	}
}

// Processor receives decoded escape sequences and runes from Buffer.Process.
type Processor interface {
	ProcessEscape(e Escape, a []byte)
	ProcessRune(r rune)
}
