package anansi

import (
	"image"
	"io"
	"syscall"

	"github.com/jcorbin/anansi/ansi"
)

// Screen combines a cell grid with cursor and screen state, supporting
// primitive vt100 emulation and differential terminal updating.
type Screen struct {
	ScreenState
	prior Grid
	buf   Buffer
	cur   Cursor
}

// Reset the internal buffer and restore cursor state to last state affected by
// WriteTo.
func (sc *Screen) Reset() {
	sc.ScreenState.Clear()
	sc.buf.Reset()
	sc.cur.Reset()
}

// Resize the current screen state, and invalidate to cause a full redraw.
func (sc *Screen) Resize(size image.Point) bool {
	if sc.ScreenState.Resize(size) {
		sc.Invalidate()
		sc.buf.Reset()
		sc.cur.Reset()
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
	aw, haveAW := w.(ansiWriter)

	// if caller didn't pass a buffered ansi writer, use internal buffer and
	// then flush to the given io.Writer
	if !haveAW {
		defer func() {
			n, err = sc.cur.WriteTo(w)
			if err == nil {
				sc.prior.Resize(sc.ScreenState.Grid.Bounds().Size())
				copy(sc.prior.Rune, sc.ScreenState.Grid.Rune)
				copy(sc.prior.Attr, sc.ScreenState.Grid.Attr)
			} else if unwrapOSError(err) != syscall.EWOULDBLOCK {
				sc.Reset()
				sc.Invalidate()
			}
		}()

		// continue prior write (e.g. after isEWouldBlock(err) above)
		if sc.cur.buf.Len() > 0 {
			return
		}

		aw = &sc.cur.buf
	}

	// perform (full or differential) update
	var m int
	m, sc.cur.CursorState = sc.ScreenState.update(aw, sc.cur.CursorState, sc.prior)
	return int64(m), err
}

// Write to the internal buffer, updating cursor state per any ANSI escape
// sequences, and advancing cursor position by rune count (clamped to screen
// size).
func (sc *Screen) Write(p []byte) (n int, err error) {
	n, _ = sc.buf.Write(p)
	sc.process()
	return n, nil
}

// WriteString to the internal buffer, updating cursor state per any ANSI
// escape sequences, and advancing cursor position by rune count (clamped to
// screen size).
func (sc *Screen) WriteString(s string) (n int, err error) {
	n, _ = sc.buf.WriteString(s)
	sc.process()
	return n, nil
}

// WriteRune to the internal buffer, advancing cursor position (clamped to
// screen size).
func (sc *Screen) WriteRune(r rune) (n int, err error) {
	n, _ = sc.buf.WriteRune(r)
	sc.process()
	return n, nil
}

// WriteByte to the internal buffer, advancing cursor position (clamped to
// screen size).
func (sc *Screen) WriteByte(b byte) error {
	_ = sc.buf.WriteByte(b)
	sc.process()
	return nil
}

// WriteESC writes one or more ANSI escapes to the internal buffer, returning
// the number of bytes written; updates cursor state as appropriate.
func (sc *Screen) WriteESC(seqs ...ansi.Escape) int {
	n := sc.buf.WriteESC(seqs...)
	sc.process()
	return n
}

// WriteSeq writes one or more ANSI escape sequences to the internal buffer,
// returning the number of bytes written; updates cursor state as appropriate.
func (sc *Screen) WriteSeq(seqs ...ansi.Seq) int {
	n := sc.buf.WriteSeq(seqs...)
	sc.process()
	return n
}

// WriteSGR writes one or more ANSI SGR sequences to the internal buffer,
// returning the number of bytes written; updates Attr cursor state.
func (sc *Screen) WriteSGR(attrs ...ansi.SGRAttr) (n int) {
	for i := range attrs {
		if attr := sc.ScreenState.MergeSGR(attrs[i]); attr != 0 {
			n += sc.buf.WriteSGR(attr)
		}
	}
	if n > 0 {
		sc.buf.Skip(n)
	}
	return n

}

func (sc *Screen) process() {
	sc.buf.Process(sc)
	sc.buf.Discard()
}

// Enter calls SizeToTerm.
func (sc *Screen) Enter(term *Term) error { return sc.SizeToTerm(term) }

// Exit Reset()s all virtual state, and restores real terminal graphics and
// cursor state.
func (sc *Screen) Exit(term *Term) error {
	// discard all virtual state...
	sc.Reset()
	// ...and restore real cursor state
	n := sc.buf.WriteSGR(sc.cur.Real.MergeSGR(0))
	n += sc.buf.WriteSeq(sc.cur.Real.Show())
	if n > 0 {
		return term.Flush(&sc.buf)
	}
	return nil
}

// SizeToTerm invalidates and resizes the screen to match the passed terminal's
// current size.
func (sc *Screen) SizeToTerm(term *Term) error {
	sz, err := term.Size()
	if err == nil {
		sc.Invalidate()
		sc.Resize(sz)
	}
	return nil
}

var (
	_ ansiWriter = &Screen{}
	_ Context    = &Screen{}
)
