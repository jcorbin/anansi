package termkey

import (
	"fmt"
	"strings"

	"github.com/jcorbin/anansi/terminfo"
)

// Modifier is encodes modifier state during an event.
type Modifier uint8

// Alt modifier constant, see Event.Mod field.
const (
	ModAlt Modifier = 1 << iota
	ModMotion
)

func (mod Modifier) stringParts(parts []string) int {
	// TODO maybe codegen this, and as more direct code
	i := 0
	for _, mp := range []struct {
		mask Modifier
		part string
	}{
		{ModAlt, "Alt"},
		{ModMotion, "Motion"},
	} {
		if mod&mp.mask != 0 {
			parts[i] = mp.part
			i++
		}
	}
	return i
}

func (mod Modifier) String() string {
	var parts [8]string
	switch i := mod.stringParts(parts[:]); i {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return strings.Join(parts[:i], "+")
	}
}

// Key encodes control and special key events.
type Key uint8

// Key constants for ASCII control characters.
const (
	KeyCtrl2 Key = 0x00 // ^@
	KeyCtrlA Key = 0x01 // ^A
	KeyCtrlB Key = 0x02 // ^B
	KeyCtrlC Key = 0x03 // ^C
	KeyCtrlD Key = 0x04 // ^D
	KeyCtrlE Key = 0x05 // ^E
	KeyCtrlF Key = 0x06 // ^F
	KeyCtrlG Key = 0x07 // ^G
	KeyCtrlH Key = 0x08 // ^H
	KeyCtrlI Key = 0x09 // ^I
	KeyCtrlJ Key = 0x0A // ^J
	KeyCtrlK Key = 0x0B // ^K
	KeyCtrlL Key = 0x0C // ^L
	KeyCtrlM Key = 0x0D // ^M
	KeyCtrlN Key = 0x0E // ^N
	KeyCtrlO Key = 0x0F // ^O
	KeyCtrlP Key = 0x10 // ^P
	KeyCtrlQ Key = 0x11 // ^Q
	KeyCtrlR Key = 0x12 // ^R
	KeyCtrlS Key = 0x13 // ^S
	KeyCtrlT Key = 0x14 // ^T
	KeyCtrlU Key = 0x15 // ^U
	KeyCtrlV Key = 0x16 // ^V
	KeyCtrlW Key = 0x17 // ^W
	KeyCtrlX Key = 0x18 // ^X
	KeyCtrlY Key = 0x19 // ^Y
	KeyCtrlZ Key = 0x1A // ^Z
	KeyCtrl3 Key = 0x1B // ^[
	KeyCtrl4 Key = 0x1C // ^\
	KeyCtrl5 Key = 0x1D // ^]
	KeyCtrl6 Key = 0x1E // ^^
	KeyCtrl7 Key = 0x1F // ^_
	KeySpace Key = 0x20
	// 0x21 - 0x7E are printable characters
	KeyCtrl8 Key = 0x7F // ^?

	// functional aliases for control characters
	KeyEsc        = KeyCtrl3
	KeyBackspace2 = KeyCtrl8
	KeyBackspace  = KeyCtrlH
	KeyTab        = KeyCtrlI
	KeyEnter      = KeyCtrlM

	// mnemonic aliases for control characters
	KeyCtrlSpace      = KeyCtrl2
	KeyCtrlTilde      = KeyCtrl2
	KeyCtrlLBracket   = KeyCtrl3
	KeyCtrlBackslash  = KeyCtrl4
	KeyCtrlRBracket   = KeyCtrl5
	KeyCtrlSlash      = KeyCtrl7
	KeyCtrlUnderscore = KeyCtrl7
)

// Key constants for special keys.
const (
	KeyF1       = 0x80 | Key(terminfo.KeyF1)
	KeyF2       = 0x80 | Key(terminfo.KeyF2)
	KeyF3       = 0x80 | Key(terminfo.KeyF3)
	KeyF4       = 0x80 | Key(terminfo.KeyF4)
	KeyF5       = 0x80 | Key(terminfo.KeyF5)
	KeyF6       = 0x80 | Key(terminfo.KeyF6)
	KeyF7       = 0x80 | Key(terminfo.KeyF7)
	KeyF8       = 0x80 | Key(terminfo.KeyF8)
	KeyF9       = 0x80 | Key(terminfo.KeyF9)
	KeyF10      = 0x80 | Key(terminfo.KeyF10)
	KeyF11      = 0x80 | Key(terminfo.KeyF11)
	KeyF12      = 0x80 | Key(terminfo.KeyF12)
	KeyInsert   = 0x80 | Key(terminfo.KeyInsert)
	KeyDelete   = 0x80 | Key(terminfo.KeyDelete)
	KeyHome     = 0x80 | Key(terminfo.KeyHome)
	KeyEnd      = 0x80 | Key(terminfo.KeyEnd)
	KeyPageUp   = 0x80 | Key(terminfo.KeyPageUp)
	KeyPageDown = 0x80 | Key(terminfo.KeyPageDown)
	KeyUp       = 0x80 | Key(terminfo.KeyUp)
	KeyDown     = 0x80 | Key(terminfo.KeyDown)
	KeyLeft     = 0x80 | Key(terminfo.KeyLeft)
	KeyRight    = 0x80 | Key(terminfo.KeyRight)

	maxTerminfoKey = KeyRight
)

// Key constants for mouse keys.
const (
	MouseLeft Key = 0x80 | Key(maxTerminfoKey+1+iota)
	MouseMiddle
	MouseRight
	MouseRelease
	MouseWheelUp
	MouseWheelDown

	// TODO if these were better aligned with X10's first byte, the parser
	// could be simpler.

	minMouseKey   = MouseLeft
	maxMouseKey   = MouseWheelDown
	maxSpecialKey = MouseWheelDown
)

//go:generate sh -c "./scripts/gen_special_strings.sh key.go termkey | gofmt >key_special_strings.go"

// IsSpecial returns true if the key is outside of the ASCII plane and is a
// defined extended Key* constant.
func (k Key) IsSpecial() bool {
	return (k&0x80) != 0 && k < maxSpecialKey
}

// IsMouse returns true if the key codes a mouse key event.
func (k Key) IsMouse() bool {
	return minMouseKey <= k && k <= maxMouseKey
}

func (k Key) String() string {
	const (
		asciiTable = "" +
			/* 0x7f	     DELete */ `^?` +
			/* 0x00	    control */ `^@^A^B^C^D^E^F^G^H^I^J^K^L^M^N^O` +
			/* 0x10 ... control */ `^P^Q^R^S^T^U^V^W^X^Y^Z^[^\^]^^^_` +
			/* 0x20	    symbols */ ` !"#$%&'()*+,-./` +
			/* 0x30	    numbers */ `0123456789` +
			/* 0x3a ... symbols */ `:;<=>?@` +
			/* 0x41	     UPPERs */ `ABCDEFGHIJKLMNOPQRSTUVWXYZ` +
			/* 0x5b ... symbols */ "[\\]^_`" +
			/* 0x60	     lowers */ `abcdefghijklmnopqrstuvwxyz` +
			/* 0x7b ... symbols */ `{|}~`
		printOff = 2 /* DEL */ + 2*0x20 /* controls */
	)
	switch {
	case k < 0x20, k == 0x7f:
		i := 2 * ((k ^ 0x40) - 0x3f)
		return asciiTable[i : i+2]
	case k < 0x7f:
		i := printOff + (k - 0x20)
		return asciiTable[i : i+1]
	case k < maxSpecialKey:
		off := specialOffs[k&0x7f]
		return specialTable[off[0]:off[1]]
	default:
		return fmt.Sprintf("Key<%02x>", uint8(k))
	}
}
