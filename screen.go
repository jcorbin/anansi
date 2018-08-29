package anansi

import (
	"image"
	"io"

	"github.com/jcorbin/anansi/ansi"
)

// Screen provides differential terminal updating: writes are decoded and
// update the pending ScreenState.
type Screen struct {
	ScreenState
	prior Grid
	proc  ansi.Buffer
	out   Cursor
}

// Reset the internal buffer and restore cursor state to last state affected by
// WriteTo.
func (sc *Screen) Reset() {
	sc.ScreenState.Clear()
	sc.proc.Reset()
	sc.out.Reset()
}

// Resize the current screen state, and invalidate to cause a full redraw.
func (sc *Screen) Resize(size image.Point) bool {
	if sc.ScreenState.Resize(size) {
		sc.Invalidate()
		sc.proc.Reset()
		sc.out.Reset()
		return true
	}
	return false
}

// Invalidate forces the next WriteTo() to perform a full redraw.
func (sc *Screen) Invalidate() {
	sc.prior.Resize(image.ZP)
}

// WriteTo builds and writes output based on the current ScreenState, doing a
// differential update if possible, or a full redraw otherwise. If the internal
// output buffer isn't empty, then the build step is skipped, and another
// attempt is made to flush the output buffer.
func (sc *Screen) WriteTo(w io.Writer) (n int64, err error) {
	if sc.out.buf.Len() == 0 {
		_, sc.out.CursorState = sc.ScreenState.Update(sc.out.CursorState, &sc.out.buf, &sc.prior)
	}
	n, err = sc.out.WriteTo(w)
	if err == nil {
		sc.ScreenState.Grid.CopyTo(&sc.prior)
	} else if !isEWouldBlock(err) {
		sc.Reset()
		sc.Invalidate()
	}
	return n, err
}

// Write to the internal buffer, updating cursor state per any ANSI escape
// sequences, and advancing cursor position by rune count (clamped to screen
// size).
func (sc *Screen) Write(p []byte) (n int, err error) {
	n, _ = sc.proc.Write(p)
	sc.process()
	return n, nil
}

// WriteString to the internal buffer, updating cursor state per any ANSI
// escape sequences, and advancing cursor position by rune count (clamped to
// screen size).
func (sc *Screen) WriteString(s string) (n int, err error) {
	n, _ = sc.proc.WriteString(s)
	sc.process()
	return n, nil
}

// WriteRune to the internal buffer, advancing cursor position (clamped to
// screen size).
func (sc *Screen) WriteRune(r rune) (n int, err error) {
	n, _ = sc.proc.WriteRune(r)
	sc.process()
	return n, nil
}

// WriteByte to the internal buffer, advancing cursor position (clamped to
// screen size).
func (sc *Screen) WriteByte(b byte) error {
	_ = sc.proc.WriteByte(b)
	sc.process()
	return nil
}

// WriteESC writes one or more ANSI escapes to the internal buffer, returning
// the number of bytes written; updates cursor state as appropriate.
func (sc *Screen) WriteESC(seqs ...ansi.Escape) int {
	n := sc.proc.WriteESC(seqs...)
	sc.process()
	return n
}

// WriteSeq writes one or more ANSI escape sequences to the internal buffer,
// returning the number of bytes written; updates cursor state as appropriate.
func (sc *Screen) WriteSeq(seqs ...ansi.Seq) int {
	n := sc.proc.WriteSeq(seqs...)
	sc.process()
	return n
}

// WriteSGR writes one or more ANSI SGR sequences to the internal buffer,
// returning the number of bytes written; updates Attr cursor state.
func (sc *Screen) WriteSGR(attrs ...ansi.SGRAttr) (n int) {
	for i := range attrs {
		if attr := sc.ScreenState.MergeSGR(attrs[i]); attr != 0 {
			n += sc.proc.WriteSGR(attr)
		}
	}
	if n > 0 {
		sc.proc.Skip(n)
	}
	return n

}

func (sc *Screen) process() {
	sc.proc.Process(sc)
	sc.proc.Discard()
}
