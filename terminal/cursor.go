// TODO naturalize ; imported from cops/cursor.go

package terminal

import (
	"fmt"
	"image/color"
	"strconv"
	"unicode/utf8"
)

/* TODO distill around

Hide hides the cursor      . "\033[?25l"
Show reveals the cursor    . "\033[?25h"
Erase the whole display    . "\033[2J"
Erase the current line     . "\033[2K"
Reset terminal to default  . "\033[m"
Seeks cursor to the origin . "\033[H"
Jump to location           . "\033[yyy;xxxH"
Return to first column     . "\r"
Next line                  . "\n"
Up                         . "\033[nnnA"
Down                       . "\033[nnnB"
Left                       . "\033[nnnD"
Right                      . "\033[nnnC"

*/

// Cursor models the known or unknown states of a cursor.
//
// Zero color values indicate that the respective color is unknown, so the next
// text must be preceded by an SGR (set graphics) ANSI sequence to set it.
type Cursor struct {
	Point
	Foreground color.RGBA
	Background color.RGBA
	Visibility
}

// Visibility represents the visibility of a Cursor.
type Visibility uint8

const (
	// MaybeVisible maybe not; represents the need to ensure state; this is the
	// zero-value of Visibility so that it's easy to default to unknown state.
	MaybeVisible Visibility = iota

	// Hidden represents a hidden cursor.
	Hidden

	// Visible represents a normal cursor.
	Visible
)

func (v Visibility) String() string {
	switch v {
	case 0:
		return "Maybe"
	case Hidden:
		return "Hidden"
	case Visible:
		return "Visible"
	default:
		return fmt.Sprintf("Invalid<%02x>", uint8(v))
	}
}

// Home is the terimnal screen origin.
var Home = Pt(1, 1)

// Default is the default terminal cursor state after a reset.
var Default = Cursor{
	Point:      Home,
	Foreground: Colors[7],
	Background: Colors[0],
}

// Hide hides the cursor.
func (c Cursor) Hide(buf []byte) ([]byte, Cursor) {
	if c.Visibility != Hidden {
		c.Visibility = Hidden
		buf = append(buf, "\033[?25l"...)
	}
	return buf, c
}

// Show reveals the cursor.
func (c Cursor) Show(buf []byte) ([]byte, Cursor) {
	if c.Visibility != Visible {
		c.Visibility = Visible
		buf = append(buf, "\033[?25h"...)
	}
	return buf, c
}

// Clear erases the whole display; implicitly invalidates the cursor position
// since its behavior is inconsistent across terminal implementations.
func (c Cursor) Clear(buf []byte) ([]byte, Cursor) {
	c.Point = Nowhere
	return append(buf, "\033[2J"...), c
}

// ClearLine erases the current line.
func (c Cursor) ClearLine(buf []byte) ([]byte, Cursor) {
	return append(buf, "\033[2K"...), c
}

// Reset returns the terminal to default white on black colors.
func (c Cursor) Reset(buf []byte) ([]byte, Cursor) {
	if c.Foreground != Colors[7] || c.Background != Colors[0] {
		//lint:ignore SA4005 broken check
		c.Foreground, c.Background = Colors[7], Colors[0]
		buf = append(buf, "\033[m"...)
	}
	return buf, c
}

// Home seeks the cursor to the origin, using display absolute coordinates.
func (c Cursor) Home(buf []byte) ([]byte, Cursor) {
	c.Point = Home
	return append(buf, "\033[H"...), c
}

func (c Cursor) recover(buf []byte, to Point) ([]byte, Cursor) {
	if c.X < 0 && c.Y < 0 {
		// If the cursor position is completely unknown, move relative to
		// screen origin. This mode must be avoided to render relative to
		// cursor position inline with a scrolling log, by setting the cursor
		// position relative to an arbitrary origin before rendering.
		return c.jumpTo(buf, to)
	}

	if c.X < 0 {
		// If only horizontal position is unknown, return to first column and
		// march forward. Rendering a non-ASCII cell of unknown or
		// indeterminate width may invalidate the column number. For example, a
		// skin tone emoji may or may not render as a single column glyph.
		buf = append(buf, "\r"...)
		c.X = 0
		// Continue...
	}

	return buf, c
}

func (c Cursor) jumpTo(buf []byte, to Point) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(to.Y), 10)
	buf = append(buf, ";"...)
	buf = strconv.AppendInt(buf, int64(to.X), 10)
	buf = append(buf, "H"...)
	c.Point = to
	return buf, c
}

func (c Cursor) linedown(buf []byte, n int) ([]byte, Cursor) {
	// Use \r\n to advance cursor Y on the chance it will advance the
	// display bounds.
	buf = append(buf, "\r\n"...)
	for m := n - 1; m > 0; m-- {
		buf = append(buf, "\n"...)
	}
	c.X = 0
	c.Y += n
	return buf, c
}

func (c Cursor) up(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "A"...)
	c.Y -= n
	return buf, c
}

func (c Cursor) down(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "B"...)
	c.Y += n
	return buf, c
}

func (c Cursor) left(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "D"...)
	c.X -= n
	return buf, c
}

func (c Cursor) right(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "C"...)
	c.X += n
	return buf, c
}

// Go does stuff ; TODO eliminate it, expose primitives ; leave this up to a
// movement model / concern of the caller.
//
// ...moves the cursor to another position, preferring to use relative motion,
// using line relative if the column is unknown, using display origin relative
// only if the line is also unknown. If the column is unknown, use "\r" to seek
// to column 0 of the same line.
func (c Cursor) Go(buf []byte, to Point) ([]byte, Cursor) {
	buf, c = c.recover(buf, to)

	if to.X == 1 && to.Y == c.Y+1 {
		buf, c = c.Reset(buf)
		buf = append(buf, "\r\n"...)
		c.X = 1
		c.Y++
	} else if to.X == 1 && c.X != 1 {
		buf, c = c.Reset(buf)
		buf = append(buf, "\r"...)
		c.X = 1

		// In addition to scrolling back to the first column generally, this
		// has the effect of resetting the column if writing a multi-byte
		// string invalidates the cursor's horizontal position. For example, a
		// skin tone emoji may or may not render as a single column glyph.
	}

	if n := to.Y - c.Y; n > 0 {
		// buf, c = c.linedown(buf, n)
		buf, c = c.down(buf, n)
	} else if n < 0 {
		buf, c = c.up(buf, -n)
	}

	if n := to.X - c.X; n > 0 {
		buf, c = c.right(buf, n)
	} else if n < 0 {
		buf, c = c.left(buf, -n)
	}

	return buf, c
}

// WriteGlyph appends the given string's UTF8 bytes into the given buffer,
// invalidating the cursor if the string COULD HAVE rendered to more than one
// glyph; otherwise the cursor's X is advanced by 1.
func (c Cursor) WriteGlyph(buf []byte, s string) ([]byte, Cursor) {
	buf = append(buf, s...)
	if n := utf8.RuneCountInString(s); n == 1 {
		c.X++
	} else {
		// Invalidate cursor column to force position reset before next draw,
		// if the string drawn might be longer than one cell wide or simply
		// empty.
		c.X = -1
	}
	return buf, c
}

// TODO: func (c Cursor) Write(buf, p []byte) ([]byte, Cursor)
