// +build darwin freebsd netbsd openbsd dragonfly

package anansi

import (
	"image"
	"syscall"
	"unsafe"
)

func (at Attr) getSize() (size image.Point, err error) {
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

func (at Attr) getAttr() (attr syscall.Termios, err error) {
	err = at.ioctl(syscall.TIOCGETA, uintptr(unsafe.Pointer(&attr)), 0, 0, 0)
	return
}

func (at Attr) setAttr(attr syscall.Termios) error {
	return at.ioctl(syscall.TIOCSETA, uintptr(unsafe.Pointer(&attr)), 0, 0, 0)
}
