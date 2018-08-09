package terminal

import (
	"bytes"
	"image"
	"os"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

// Output encapsulates terminal output buffering TODO elaborate
type Output struct {
	File *os.File

	buf  bytes.Buffer
	err  error
	tcur Cursor
	bcur Cursor
	tmp  []byte
	writeObserver
}

type writeObserver interface {
	// preWrite gets called before a write to the output buffer giving a
	// chance to flush; n is a best-effort size of the bytes about to be
	// written. NOTE preWrite MUST avoid manipulating cursor state, as it may
	// reflect state about to be implemented by the written bytes.
	preWrite(out *Output, n int) error

	// postWrite gets called after a write to the output buffer giving a chance to flush.
	postWrite(out *Output, n int) error
}

// Curse is a single cursor manipulator; NOTE the type asymmetry is due to
// complying with the shape of Cursor methods like Cursor.Show.
type Curse func(Cursor, []byte) ([]byte, Cursor)

// Init ialize Output state; TODO would be great to eliminate this method.
func (out *Output) Init() {
	out.tmp = make([]byte, 64)
	out.writeObserver = flushWhenFull{}
}

func (out *Output) ioctl(request, arg1, arg2, arg3, arg4 uintptr) error {
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, out.File.Fd(), request, arg1, arg2, arg3, arg4); e != 0 {
		return e
	}
	return nil
}

// GetAttr retrieves terminal attributes.
func (out *Output) GetAttr() (attr syscall.Termios, err error) {
	err = out.ioctl(syscall.TIOCGETA, uintptr(unsafe.Pointer(&attr)), 0, 0, 0)
	return
}

// SetAttr sets terminal attributes.
func (out *Output) SetAttr(attr syscall.Termios) error {
	return out.ioctl(syscall.TIOCSETA, uintptr(unsafe.Pointer(&attr)), 0, 0, 0)
}

// Size reads and returns the current terminal size.
func (out *Output) Size() (size image.Point, err error) {
	// TODO cache last known good? hide error?
	var dim struct {
		rows    uint16
		cols    uint16
		xpixels uint16
		ypixels uint16
	}
	err = out.ioctl(syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&dim)), 0, 0, 0)
	if err == nil {
		size.X = int(dim.cols)
		size.Y = int(dim.rows)
	}
	return size, err
}

// Write into the output buffer, triggering any Flush* options.
func (out *Output) Write(p []byte) (n int, err error) {
	if out.err != nil {
		return 0, out.err
	}
	if out.writeObserver == nil {
		return out.buf.Write(p)
	}

	// TODO would be nice to give writeOption a choice to pass large
	// buffers through rather than append/growing them
	out.err = out.writeObserver.preWrite(out, len(p))
	if out.err != nil {
		return 0, out.err
	}

	n, _ = out.buf.Write(p)
	out.err = out.writeObserver.postWrite(out, n)
	return n, out.err
}

// WriteByte into the output buffer, triggering any Flush* options.
func (out *Output) WriteByte(c byte) error {
	if out.err != nil {
		return out.err
	}
	if out.writeObserver == nil {
		return out.buf.WriteByte(c)
	}

	out.err = out.writeObserver.preWrite(out, 1)
	if out.err != nil {
		return out.err
	}

	_ = out.buf.WriteByte(c)
	out.err = out.writeObserver.postWrite(out, 1)
	return out.err
}

// WriteRune into the output buffer, triggering any Flush* options.
func (out *Output) WriteRune(r rune) (n int, err error) {
	if out.err != nil {
		return 0, out.err
	}
	if out.writeObserver == nil {
		return out.buf.WriteRune(r)
	}

	out.err = out.writeObserver.preWrite(out, utf8.RuneLen(r))
	if out.err != nil {
		return 0, out.err
	}

	n, _ = out.buf.WriteRune(r)
	out.err = out.writeObserver.postWrite(out, n)
	return n, out.err
}

// WriteString into the output buffer, triggering any Flush* options.
func (out *Output) WriteString(s string) (n int, err error) {
	if out.err != nil {
		return 0, out.err
	}
	if out.writeObserver == nil {
		return out.buf.WriteString(s)
	}

	// TODO would be nice to give writeOption a choice to pass large
	// strings through rather than append/growing them
	out.err = out.writeObserver.preWrite(out, len(s))
	if out.err != nil {
		return 0, out.err
	}

	n, _ = out.buf.WriteString(s)
	out.err = out.writeObserver.postWrite(out, n)
	return n, out.err
}

// WriteCursor writes cursor control codes into the output buffer, and updates
// cursor state, triggering any Flush* options.
func (out *Output) WriteCursor(curses ...Curse) (n int, err error) {
	if out.err != nil {
		return 0, out.err
	}
	switch len(curses) {
	case 0:
		return 0, nil
	case 1:
		_, out.tmp, out.bcur = writeCursor(out.bcur, out.tmp[:0], curses[0])
	default:
		out.tmp = out.tmp[:0]
		for i := range curses {
			_, out.tmp, out.bcur = writeCursor(out.bcur, out.tmp, curses[i])
		}
	}
	return out.Write(out.tmp)
}

// Flush any buffered output.
func (out *Output) Flush() error {
	if out.err == nil && out.buf.Len() > 0 {
		_, out.err = out.buf.WriteTo(out.File)
		out.tcur = out.bcur
	}
	return out.err
}

// Discard any un-flushed output.
func (out *Output) Discard() error {
	if out.err == nil {
		out.buf.Reset()
		out.err = out.writeObserver.preWrite(out, 0)
		out.bcur = out.tcur
	}
	return out.err
}

func writeCursor(cur Cursor, buf []byte, curse Curse) (n int, _ []byte, _ Cursor) {
	m := len(buf)
	buf, cur = curse(cur, buf)
	n += len(buf) - m
	return n, buf, cur
}

// TODO maybe pivot writeObserver around a different abstraction like
// Write(p []byte) (n int, err error)
// WriteByte(c byte) error
// WriteRune(r rune) (n int, err error)
// WriteString(s string) (n int, err error)
// WriteCursor(curses ...Curse) (n int, err error)
// Flush() error
// Discard() error
