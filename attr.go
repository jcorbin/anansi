package anansi

import (
	"errors"
	"image"
	"os"
	"syscall"
)

var errAttrNoFile = errors.New("anansi.Attr.ioctl: no File set")

// Attr implements Context-ual manipulation and interrogation of terminal
// state, using the termios IOCTLs and ANSI control sequences where possible.
type Attr struct {
	file *os.File // XXX re-export

	ownFile bool
	orig    syscall.Termios
	cur     syscall.Termios
	raw     bool
	echo    bool
}

// IsTerminal returns true only if the given file is attached to an interactive
// terminal.
func IsTerminal(f *os.File) bool {
	return Attr{file: f}.IsTerminal()
}

// IsTerminal returns true only if both terminal input and output file handles
// are both connected to a valid terminal.
func (term *Term) IsTerminal() bool {
	return IsTerminal(term.Input.File) &&
		IsTerminal(term.Output.File)
}

// IsTerminal returns true only if the underlying file is attached to an
// interactive terminal.
func (at Attr) IsTerminal() bool {
	_, err := at.getAttr()
	return err == nil
}

// Size reads and returns the current terminal size.
func (at Attr) Size() (size image.Point, err error) {
	return at.getSize()
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
	if at.file == nil {
		return nil
	}
	at.cur = at.modifyTermios(at.orig)
	if err := at.setAttr(at.cur); err != nil {
		return err
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
	if at.file == nil {
		return nil
	}
	if echo {
		at.cur.Lflag |= syscall.ECHO
	} else {
		at.cur.Lflag &^= syscall.ECHO
	}
	if err := at.setAttr(at.cur); err != nil {
		return err
	}
	return nil
}

func (at Attr) modifyTermios(attr syscall.Termios) syscall.Termios {
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

// Enter default the Attr's file to the term's Output File, records its
// original termios attributes, and then applies termios attributes.
func (at *Attr) Enter(term *Term) (err error) {
	if at.file == nil {
		at.file = term.Output.File
		at.ownFile = false
	} else {
		at.ownFile = true
	}
	at.orig, err = at.getAttr()
	if err != nil {
		return err
	}
	at.cur = at.modifyTermios(at.orig)
	if err = at.setAttr(at.cur); err != nil {
		return err
	}
	return nil
}

// Exit restores termios attributes, and clears the File pointer if it was set
// by Enter
func (at *Attr) Exit(term *Term) error {
	if at.file == nil {
		return nil
	}
	if err := at.setAttr(at.orig); err != nil {
		return err
	}
	if !at.ownFile {
		at.file = nil
		at.ownFile = false
	}
	return nil
}

func (at Attr) ioctl(request, arg1, arg2, arg3, arg4 uintptr) error {
	if at.file == nil {
		return errAttrNoFile
	}
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, at.file.Fd(), request, arg1, arg2, arg3, arg4); e != 0 {
		return e
	}
	return nil
}
