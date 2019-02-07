package anansi

import (
	"image"
	"io"
	"syscall"

	"github.com/jcorbin/anansi/ansi"
)

// TermScreen supports attaching a ScreenDiffer to a Term's Context.
type TermScreen struct {
	ScreenDiffer
}

// ScreenDiffer supports deferred screen updating by tracking desired virtual screen
// state vs last known (Real) screen state. It also supports tracking final
// desired user cursor state, separate from any cursor state used to
// update the virtual screen. Primitive vt100 emulation is provided through
// Write* methods and Buffer processing.
type ScreenDiffer struct {
	UserCursor Cursor

	VirtualScreen
	Real Screen
}

// VirtualScreen implements minimal terminal emulation around ScreenState.
//
// Normal textual output and control sequences may be written to it, using
// either high-level convenience methods, or the standard low-level
// byte/string/rune/byte writing methods. All such output is collected and
// processed through an internal Buffer, and ScreenState updated accordingly.
//
// TODO the vt100 emulation is certainly too minimal at present; it needs to be
// completed, and validated. E.g. cursor position is currently clamped to the
// screen size, rather than wrapped.
type VirtualScreen struct {
	Screen
	buf Buffer
}

// Reset calls Clear() and restores virtual cursor state.
func (sc *ScreenDiffer) Reset() {
	sc.Clear()
	sc.VirtualScreen.Cursor = sc.Real.Cursor
}

// Clear the virtual screen, user cursor state, and internal buffer.
func (sc *ScreenDiffer) Clear() {
	sc.VirtualScreen.Clear()
	sc.UserCursor = Cursor{}
	sc.buf.Reset()
}

// Resize the current screen state, and invalidate to cause a full redraw.
func (sc *ScreenDiffer) Resize(size image.Point) bool {
	if sc.VirtualScreen.Resize(size) {
		sc.Invalidate()
		sc.buf.Reset()
		return true
	}
	return false
}

// Invalidate forces the next WriteTo() to perform a full redraw.
func (sc *ScreenDiffer) Invalidate() {
	sc.Real.Resize(image.ZP)
}

// WriteTo writes the content and ansi control sequences necessary to
// synchronize Real screen state to match VirtualScreen state.  performs a
// differential update if possible, falling back to a full redraw if necessary
// or forced by Invalidate().
//
// If the given io.Writer implements higher level ansi writing methods they
// are used directly; otherwise an internal Buffer is used to first assemble
// the needed output, then delegating to Buffer.WriteTo to flush output.
//
// When using an internal Buffer, resuming a partial write after EWOULDBLOCK is
// supported, skipping the assembly step described above.
func (sc *ScreenDiffer) WriteTo(w io.Writer) (n int64, err error) {
	state := sc.Real
	defer func() {
		if err == nil {
			sc.Real = state
		}
	}()

	aw, haveAW := w.(ansiWriter)

	// if caller didn't pass a buffered ansi writer, use internal buffer and
	// then flush to the given io.Writer
	if !haveAW {
		defer func() {
			n, err = sc.buf.WriteTo(w)
			if err != nil && unwrapOSError(err) != syscall.EWOULDBLOCK {
				sc.Reset()
				sc.Invalidate()
			}
		}()

		// continue prior write (e.g. after isEWouldBlock(err) above)
		if sc.buf.Len() > 0 {
			return
		}

		aw = &sc.buf
	}

	// enforce final user cursor state
	sc.VirtualScreen.Cursor = sc.UserCursor

	// perform (full or differential) update
	var m int
	m, state = sc.VirtualScreen.update(aw, state)
	return int64(m), err
}

// Write writes bytes to the internal buffer, updating screen state as describe
// on VirtualScreen. Always returns nil error.
func (vsc *VirtualScreen) Write(p []byte) (n int, err error) {
	n, _ = vsc.buf.Write(p)
	vsc.process()
	return n, nil
}

// WriteString writes a string to the internal buffer, updating screen state as
// describe on VirtualScreen. Always returns nil error.
func (vsc *VirtualScreen) WriteString(s string) (n int, err error) {
	n, _ = vsc.buf.WriteString(s)
	vsc.process()
	return n, nil
}

// WriteRune writes a single rune to the internal buffer, updating screen state
// as described on VirtualScreen. Always returns nil error.
func (vsc *VirtualScreen) WriteRune(r rune) (n int, err error) {
	n, _ = vsc.buf.WriteRune(r)
	vsc.process()
	return n, nil
}

// WriteByte writes a single byte to the internal buffer, updating screen state
// as described on VirtualScreen. Always returns nil error.
func (vsc *VirtualScreen) WriteByte(b byte) error {
	_ = vsc.buf.WriteByte(b)
	vsc.process()
	return nil
}

// WriteESC writes one or more ANSI escapes to the internal buffer, updating
// screen state as described on VirtualScreen.
func (vsc *VirtualScreen) WriteESC(seqs ...ansi.Escape) int {
	n := vsc.buf.WriteESC(seqs...)
	vsc.process()
	return n
}

// WriteSeq writes one or more ANSI escape sequences to the internal buffer,
// updating screen state as described on VirtualScreen.
func (vsc *VirtualScreen) WriteSeq(seqs ...ansi.Seq) int {
	n := vsc.buf.WriteSeq(seqs...)
	vsc.process()
	return n
}

// WriteSGR writes one or more ANSI SGR sequences to the internal buffer,
// updating screen state as described on VirtualScreen.
func (vsc *VirtualScreen) WriteSGR(attrs ...ansi.SGRAttr) (n int) {
	for i := range attrs {
		if attr := vsc.Cursor.MergeSGR(attrs[i]); attr != 0 {
			n += vsc.buf.WriteSGR(attr)
		}
	}
	if n > 0 {
		vsc.buf.Skip(n)
	}
	return n

}

func (vsc *VirtualScreen) process() {
	vsc.buf.Process(vsc)
	vsc.buf.Discard()
}

// Enter calls SizeToTerm.
func (tsc *TermScreen) Enter(term *Term) error { return tsc.SizeToTerm(term) }

// Exit Reset()s all virtual state, and restores real terminal graphics and
// cursor state.
func (tsc *TermScreen) Exit(term *Term) error {
	// discard all virtual state...
	tsc.Reset()
	// ...and restore real cursor state
	n := tsc.buf.WriteSGR(tsc.Real.Cursor.MergeSGR(0))
	n += tsc.buf.WriteSeq(tsc.Real.Cursor.Show())
	if n > 0 {
		return term.Flush(&tsc.buf)
	}
	return nil
}

// SizeToTerm invalidates and resizes the screen to match the passed terminal's
// current size.
func (tsc *TermScreen) SizeToTerm(term *Term) error {
	sz, err := term.Size()
	if err == nil {
		tsc.Invalidate()
		tsc.Resize(sz)
	}
	return nil
}

var (
	_ ansiWriter = &ScreenDiffer{}
	_ Context    = &TermScreen{}
)
