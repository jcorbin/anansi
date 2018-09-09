package ansi

import (
	"errors"
	"fmt"
)

// MouseState represents buttons presses, button releases, motions, and scrolling.
type MouseState uint8

// MouseState constants, mapped directly to xterm's state bit fields. Users
// should usually be better served by functions like MouseState.ButtonID(),
// MouseState.Modifier(), and all the MouseState.Is* variants.
const (
	MouseButton1    MouseState = 0
	MouseButton2    MouseState = 1
	MouseButton3    MouseState = 2
	MouseNoButton   MouseState = 3
	MouseModShift   MouseState = 1 << 2 // 4
	MouseModMeta    MouseState = 1 << 3 // 8
	MouseModControl MouseState = 1 << 4 // 16
	MouseMotion     MouseState = 1 << 5 // 32
	MouseWheel      MouseState = 1 << 6 // 64
	MouseRelease    MouseState = 1 << 7 // 128
)

// ButtonID returns a normalized mouse button number: 1=left,
// 2=middle, 3=right, 4=wheel-up, 5=wheel-down (the value 6 is not
// technically impossible, but shouldn't be seen in practice).
func (ms MouseState) ButtonID() uint8 {
	i := (uint8(ms&(MouseButton1|MouseButton2|MouseButton3)) + 1) & 0x3 // % 4
	if ms&MouseWheel != 0 {
		i += 3
	}
	return i
}

// Modifier returns just the modifier bits, which can be tested against the
// constants MouseModShift, MouseModControl, and MouseModMeta.
func (ms MouseState) Modifier() MouseState {
	return ms & (MouseModShift | MouseModMeta | MouseModControl)
}

// IsMotion returns true if the mouse state represents motion (with or without a button held).
func (ms MouseState) IsMotion() bool { return ms&MouseMotion != 0 }

// IsRelease returns true if the state represents a button release.
func (ms MouseState) IsRelease() bool { return ms&MouseRelease != 0 }

// IsPress returns true if the state represents a button press.
func (ms MouseState) IsPress() (id uint8, is bool) {
	if ms&(MouseRelease|MouseMotion) != 0 {
		return 0, false
	}
	return ms.ButtonID(), true
}

// IsDrag returns true if the state represents motion with a button held.
func (ms MouseState) IsDrag() bool {
	return ms.ButtonID() != 0 &&
		ms&(MouseRelease|MouseMotion) == MouseMotion
}

func (ms MouseState) String() string {
	s := ms.ButtonName()
	switch ms & (MouseMotion | MouseRelease) {
	case MouseMotion:
		if s == "" {
			s = "motion"
		} else {
			s += "-drag"
		}
	case MouseRelease:
		if s == "" {
			s = "??? release"
		} else {
			s += "-release"
		}
	case MouseMotion | MouseRelease: // XXX shouldn't be seen in practice
		if s == "" {
			s = "??? motion+release"
		} else {
			s = "??? " + s + "-motion+release"
		}
	}
	if mod := ms.ModifierName(); mod != "" {
		return mod + "+" + s
	}
	return s
}

// ButtonName returns a string representing the button, like "left".
func (ms MouseState) ButtonName() string {
	switch ms.ButtonID() {
	case 0:
		return "" //
	case 1:
		return "left" //
	case 2:
		return "middle" //
	case 3:
		return "right" //
	case 4:
		return "wheelUp" //
	case 5:
		return "wheelDown" //
	case 6:
		return "thirdWheel" // XXX shouldn't be seen in practice
	default:
		panic("inconceivable") // 0x3 + 3 max above
	}
}

// ModifierName returns a string representing the Modifier() bits, like "shift+ctrl".
func (ms MouseState) ModifierName() string {
	switch ms.Modifier() {
	case 0:
		return ""
	case MouseModControl:
		return "ctrl"
	case MouseModShift:
		return "shift"
	case MouseModMeta:
		return "meta"
	case MouseModControl | MouseModShift:
		return "ctrl+shift"
	case MouseModControl | MouseModMeta:
		return "ctrl+meta"
	case MouseModShift | MouseModMeta:
		return "shift+meta"
	case MouseModControl | MouseModShift | MouseModMeta:
		return "ctrl+shift+meta"
	default:
		panic("inconceivable") // exhaustive cases
	}
}

// MouseDecodeError represents an error decoding a mouse control
// sequence argument.
type MouseDecodeError struct {
	ID   Escape
	Arg  []byte
	What string
	Err  error
}

func (mde MouseDecodeError) Error() string {
	return fmt.Sprintf(
		"invalid mouse %s in %v %s: %v",
		mde.What, mde.ID, mde.Arg, mde.Err)
}

var errExtraBytes = errors.New("unexpected extra argument bytes")

// DecodeXtermExtendedMouse decodes xterm extended (mode 1006) mouse control
// sequences of the form:
//
// 	CSI < Cb ; Cx ; Cy M
// 	CSI < Cb ; Cx ; Cy m
func DecodeXtermExtendedMouse(id Escape, arg []byte) (b MouseState, x, y int, err error) {
	if (id == CSI('M') || id == CSI('m')) && len(arg) > 0 && arg[0] == '<' {
		arg0 := arg
		arg = arg[1:]

		v, n, err := DecodeNumber(arg)
		if err == nil && v > 0xff || v < 0 {
			err = errRange
		}
		if err != nil {
			return 0, 0, 0, MouseDecodeError{id, arg0, "Cb", err}
		}
		b = MouseState(v)
		if id == CSI('m') {
			b |= MouseRelease
		}
		arg = arg[n:]

		x, n, err = DecodeNumber(arg)
		if err != nil {
			return 0, 0, 0, MouseDecodeError{id, arg0, "Cx", err}
		}
		arg = arg[n:]

		y, n, err = DecodeNumber(arg)
		if err != nil {
			return 0, 0, 0, MouseDecodeError{id, arg0, "Cy", err}
		}
		arg = arg[n:]

		if len(arg) > 0 {
			return 0, 0, 0, MouseDecodeError{id, arg0, "sequence", errExtraBytes}
		}
	}
	return b, x, y, nil
}

// TODO support more than just extended xterm mouse decoding?

// Normal tracking mode sends an escape sequence on both button press and
// release. Modifier key (shift, ctrl, meta) information is also sent. It
// is enabled by specifying parameter 1000 to DECSET. On button press or
// release, xterm sends CSI M CbCxCy.
//
// * The low two bits of Cb encode button information:
//   0 = MB1 pressed
//   1 = MB2 pressed
//   2 = MB3 pressed
//   3 = release
//
// * The next three bits encode the modifiers which were down when the
//   button was pressed and are added together:
// 	  4  = Shift
// 	  8  = Meta
// 	  16 = Control
//   Note however that the shift and control bits are normally unavailable
//   because xterm uses the control modifier with mouse for popup menus,
//   and the shift modifier is used in the default translations for button
//   events. The Meta modifier recognized by xterm is the mod1 mask, and is
//   not necessarily the "Meta" key (see xmodmap(1)).
//
// * Cx and Cy are the x and y coordinates of the mouse event, encoded as
//   in X10 mode.
