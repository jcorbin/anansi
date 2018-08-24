package anansi

import (
	"image"
	"os"
	"syscall"
	"unsafe"
)

// Attr implements Context-ual manipulation and interrogation of terminal
// state, using the termios IOCTLs and ANSI control sequences where possible.
type Attr struct {
	orig syscall.Termios
	cur  syscall.Termios
	raw  bool
	echo bool

	f *os.File
}

// Size reads and returns the current terminal size.
func (at *Attr) Size() (size image.Point, err error) {
	var dim struct {
		rows    uint16
		cols    uint16
		xpixels uint16
		ypixels uint16
	}
	err = at.ioctl(syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&dim)), 0, 0, 0)
	if err == nil {
		size.X = int(dim.cols)
		size.Y = int(dim.rows)
	}
	return size, err
}

// SetRaw controls whether the terminal should be in raw mode.
//
// Raw mode is suitable for full-screen terminal user interfaces, eliminating
// keyboard shortcuts for job control, echo, line buffering, and escape key
// debouncing.
func (at *Attr) SetRaw(raw bool) error {
	if raw == at.raw {
		return nil
	}
	at.raw = raw
	if at.f != nil {
		at.cur = at.modifyTermios(at.orig)
		return at.setAttr(at.cur)
	}
	return nil
}

// SetEcho toggles input echoing mode, which is off by default in raw mode, and
// on in normal mode.
func (at *Attr) SetEcho(echo bool) error {
	if echo == at.echo {
		return nil
	}
	at.echo = echo
	if at.f != nil {
		if echo {
			at.cur.Lflag |= syscall.ECHO
		} else {
			at.cur.Lflag &^= syscall.ECHO
		}
		return at.setAttr(at.cur)
	}
	return nil
}

func (at *Attr) modifyTermios(attr syscall.Termios) syscall.Termios {
	if at.raw {
		// TODO read things like antirez's kilo notes again

		// TODO naturalize / decompose
		attr.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
		attr.Oflag &^= syscall.OPOST
		attr.Cflag &^= syscall.CSIZE | syscall.PARENB
		attr.Cflag |= syscall.CS8
		attr.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
		attr.Cc[syscall.VMIN] = 1
		attr.Cc[syscall.VTIME] = 0

	}
	if at.echo {
		attr.Lflag |= syscall.ECHO
	} else {
		attr.Lflag &^= syscall.ECHO
	}
	return attr
}

// Enter applies termios attributes, retaining the file handle so that all
// future calls to Set* now immediately.
func (at *Attr) Enter(f *os.File) (err error) {
	at.f = f
	if at.orig, err = at.getAttr(); err == nil {
		at.cur = at.modifyTermios(at.orig)
		err = at.setAttr(at.cur)
	}
	return err
}

// Exit restores termios attributes only if the given file is the retained one,
// clearing the retained file pointer to transition out of immediate
// application mode.
func (at *Attr) Exit(f *os.File) error {
	if at.f == f {
		err := at.setAttr(at.orig)
		at.f = nil
		return err
	}
	return nil
}

func (at *Attr) ioctl(request, arg1, arg2, arg3, arg4 uintptr) error {
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, at.f.Fd(), request, arg1, arg2, arg3, arg4); e != 0 {
		return e
	}
	return nil
}

func (at *Attr) getAttr() (attr syscall.Termios, err error) {
	err = at.ioctl(syscall.TIOCGETA, uintptr(unsafe.Pointer(&attr)), 0, 0, 0)
	return
}

func (at *Attr) setAttr(attr syscall.Termios) error {
	return at.ioctl(syscall.TIOCSETA, uintptr(unsafe.Pointer(&attr)), 0, 0, 0)
}
