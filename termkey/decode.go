package termkey

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"strconv"
	"unicode/utf8"

	"github.com/jcorbin/anansi/terminfo"
)

// Decoder supports decoding Events under some terminfo encoding.
type Decoder struct {
	ea escapeAutomaton
}

// NewDecoder creates a new Decoder for a given set of terminfo definitions.
func NewDecoder(info *terminfo.Terminfo) *Decoder {
	dec := &Decoder{}
	for i, s := range info.Keys {
		if len(s) > 0 {
			dec.ea.addChain([]byte(s), terminfo.KeyCode(i))
		}
	}
	return dec
}

// Decode an event from the given bytes, returning it and the number of bytes
// consumed.
func (dec Decoder) Decode(buf []byte) (ev Event, n int) {
	if len(buf) == 0 {
		return ev, 0
	}
	defer func() {
		if len(buf) > 16 && n == 0 {
			panic(fmt.Sprintf("FIXME broken key parsing; making no progress on %q", buf))
		}
	}()
	switch c := buf[0]; {
	case c == 0x1b: // escape (maybe sequence)
		ev, n = dec.ea.decode(buf)
	case c < 0x20, c == 0x7f: // ASCII control character
		ev.Key, n = Key(c), 1
	case c < utf8.RuneSelf: // printable ASCII rune
		ev.Ch, n = rune(c), 1
	default: // non-trivial rune
		ev.Ch, n = utf8.DecodeRune(buf)
	}
	return ev, n
}

// TODO unify mouse escape sequence parsing and the escapeAutomaton

var errNoSemicolon = errors.New("missing ; in xterm mouse code")

func decodeMouseEvent(buf []byte) (ev Event, n int) {
	if len(buf) < 4 {
		return ev, 0
	}
	switch buf[2] {
	case 'M':
		if len(buf) < 6 {
			return ev, 0
		}
		ev, n = decodeX10MouseEvent(buf[3:6])
		n += 3
	case '<':
		if len(buf) < 8 {
			return ev, 0
		}
		ev, n = decodeXtermMouseEvent(buf[3:])
		n += 3
	default:
		if len(buf) < 7 {
			return ev, 0
		}
		ev, n = decodeUrxvtMouseEvent(buf[2:])
		n += 2
	}
	ev.X-- // the coord is 1,1
	ev.Y-- // for upper left
	return ev, n
}

// X10 mouse encoding, the simplest one: \033 [ M Cb Cx Cy
func decodeX10MouseEvent(buf []byte) (ev Event, n int) {
	ev = decodeX10MouseEventByte(int64(buf[0]) - 32)
	ev.X = int(buf[1]) - 32
	ev.Y = int(buf[2]) - 32
	return ev, 3
}

// xterm 1006 extended mode: \033 [ < Cb ; Cx ; Cy (M or m)
func decodeXtermMouseEvent(buf []byte) (ev Event, n int) {
	mi := bytes.IndexAny(buf, "Mm")
	if mi == -1 {
		return ev, 0
	}

	b, x, y, err := decodeXtermMouseComponents(buf[:mi])
	if err != nil {
		return ev, 0
	}

	// unlike x10 and urxvt, in xterm Cb is already zero-based
	ev = decodeX10MouseEventByte(b)
	if buf[mi] != 'M' {
		// on xterm mouse release is signaled by lowercase m
		ev.Key = MouseRelease
	}
	ev.Point = image.Pt(int(x), int(y))
	return ev, mi + 1
}

// urxvt 1015 extended mode: \033 [ Cb ; Cx ; Cy M
func decodeUrxvtMouseEvent(buf []byte) (ev Event, n int) {
	mi := bytes.IndexByte(buf, 'M')
	if mi == -1 {
		return ev, 0
	}

	b, x, y, err := decodeXtermMouseComponents(buf[:mi])
	if err != nil {
		return ev, 0
	}

	ev = decodeX10MouseEventByte(b - 32)
	ev.X = int(x)
	ev.Y = int(y)

	return ev, mi + 1
}

// the common "Cb" in decode*MouseEvent
func decodeX10MouseEventByte(b int64) (ev Event) {
	switch b & 3 {
	case 0:
		if b&64 != 0 {
			ev.Key = MouseWheelUp
		} else {
			ev.Key = MouseLeft
		}
	case 1:
		if b&64 != 0 {
			ev.Key = MouseWheelDown
		} else {
			ev.Key = MouseMiddle
		}
	case 2:
		ev.Key = MouseRight

	case 3:
		ev.Key = MouseRelease
	}

	if b&32 != 0 {
		ev.Mod |= ModMotion
	}
	return ev
}

// the "; Cx ; Cy" for xterm and urxvt
func decodeXtermMouseComponents(buf []byte) (b, x, y int64, err error) {
	// Cb ;
	i := bytes.IndexByte(buf, ';')
	if i == -1 {
		return 0, 0, 0, errNoSemicolon
	}
	s := string(buf[:i])
	b, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid Cb=%q: %v", s, err)
	}
	buf = buf[i+1:]

	// Cx ;
	i = bytes.IndexByte(buf, ';')
	if i == -1 {
		return 0, 0, 0, errNoSemicolon
	}
	s = string(buf[:i])
	x, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid Cx=%q: %v", s, err)
	}
	buf = buf[i+1:]

	// Cy
	s = string(buf)
	y, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid Cy=%q: %v", s, err)
	}

	return b, x, y, nil
}
