package anansi

import (
	"fmt"
	"image"
	"unicode"

	"github.com/jcorbin/anansi/ansi"
)

var cursorMoves = [4]image.Point{
	image.Pt(0, -1), // CUU ^[[A
	image.Pt(0, 1),  // CUD ^[[B
	image.Pt(1, 0),  // CUF ^[[C
	image.Pt(-1, 0), // CUB ^[[D
}

// CursorState represents cursor state, allowing consumers to reason virtually
// about it (e.g. to affect relative movement and SGR changes).
type CursorState struct {
	image.Point
	Attr    ansi.SGRAttr
	Visible bool

	attrKnown bool
	visKnown  bool
	// TODO other mode bits? like shape?
}

// ScreenState adds a Grid and UserCusor to CursorState, allowing consumers to
// reason about the virtual state of the terminal screen. It's primary purpose
// is differential draw update, see the Update() method.
type ScreenState struct {
	CursorState
	UserCursor CursorState
	Grid
}

func (cs CursorState) String() string {
	return fmt.Sprintf("@%v a:%v v:%v", cs.Point, cs.Attr, cs.Visible)
}

func (scs ScreenState) String() string {
	return fmt.Sprintf("%v uc:(%v) gsz:%v", scs.CursorState, scs.UserCursor, scs.Grid.Size)
}

// Clear the screen grid, and reset the UserCursor (to invisible nowhere).
func (scs *ScreenState) Clear() {
	scs.Grid.Clear()
	scs.CursorState.Point = image.Pt(0, 0)
	scs.CursorState.Attr = 0
	scs.UserCursor = CursorState{}
}

// Resize the underlying Grid, and zero the cursor position if out of bounds.
// Returns true only if the resize was a change, false if it was a no-op.
func (scs *ScreenState) Resize(size image.Point) bool {
	if scs.Grid.Resize(size) {
		if scs.X > scs.Size.X || scs.Y > scs.Size.Y {
			scs.Point = image.ZP
		}
		return true
	}
	return false
}

// Show returns the control sequence necessary to show the cursor if it is not
// visible, the zero sequence otherwise.
func (cs *CursorState) Show() ansi.Seq {
	// TODO terminfo
	if !cs.visKnown || !cs.Visible {
		cs.visKnown = true
		cs.Visible = true
		return ansi.ShowCursor.Set()
	}
	return ansi.Seq{}
}

// Hide returns the control sequence necessary to hide the cursor if it is
// visible, the zero sequence otherwise.
func (cs *CursorState) Hide() ansi.Seq {
	// TODO terminfo
	if !cs.visKnown || cs.Visible {
		cs.visKnown = true
		cs.Visible = false
		return ansi.ShowCursor.Reset()
	}
	return ansi.Seq{}
}

// MergeSGR merges the given SGR attribute into Attr, returning the difference.
func (cs *CursorState) MergeSGR(attr ansi.SGRAttr) ansi.SGRAttr {
	if !cs.attrKnown {
		cs.Attr = attr
		cs.attrKnown = true
		return attr | ansi.SGRAttrClear
	}
	diff := cs.Attr.Diff(attr)
	cs.Attr = cs.Attr.Merge(diff)
	return diff
}

// To constructs an ansi control sequence that will move the cursor to the
// given point using absolute (ansi.CUP) or relative (ansi.{CUU,CUD,CUF,CUD})
// if possible. Returns a zero sequence if the cursor is already at the given
// point.
func (cs *CursorState) To(pt image.Point) ansi.Seq {
	if cs.Point == image.ZP {
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.CUP.WithInts(pt.Y, pt.X)
	}
	if pt.Y == cs.Y+1 && pt.X == 1 {
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.Escape('\r').With('\n')
	}
	if pt.X != cs.X && pt.Y != cs.Y {
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.CUP.WithInts(pt.Y, pt.X)
	}

	dx := pt.X - cs.X
	cs.X = pt.X
	switch {
	case dx == 0:
	case dx == 1:
		return ansi.CUF.With()
	case dx == -1:
		return ansi.CUB.With()
	case dx > 0:
		return ansi.CUF.WithInts(dx)
	case dx < 0:
		return ansi.CUB.WithInts(-dx)
	}

	dy := pt.X - cs.X
	cs.Y = pt.Y
	switch {
	case dy == 0:
	case dy == 1:
		return ansi.CUD.With()
	case dy == -1:
		return ansi.CUU.With()
	case dy > 0:
		return ansi.CUD.WithInts(dy)
	case dy < 0:
		return ansi.CUU.WithInts(-dy)
	}

	return ansi.Seq{}
}

// To sets the virtual cursor point to the supplied one.
func (scs *ScreenState) To(pt image.Point) {
	scs.Point = clampToScreen(scs.Size, pt)
}

// ApplyTo applies the receiver cursor state into the passed state value,
// writing any necessary control sequences into the provided buffer. Returns
// the number of bytes written, and the updated cursor state.
func (cs *CursorState) ApplyTo(cur CursorState, buf *ansi.Buffer) (n int, _ CursorState) {
	if cs.Visible && cs.Point != image.ZP {
		n += buf.WriteSeq(cur.To(cs.Point))
		n += buf.WriteSGR(cur.MergeSGR(cs.Attr))
		n += buf.WriteSeq(cur.Show())
	} else {
		n += buf.WriteSeq(cur.Hide())
	}
	return n, cur
}

// Update performs a Grid differential update with the cursor hidden, and then
// applies any non-zero UserCursor, returning the number of bytes written into
// the given buffer, and the final cursor state.
func (scs *ScreenState) Update(cur CursorState, buf *ansi.Buffer, p *Grid) (n int, _ CursorState) {
	n += buf.WriteSeq(cur.Hide())
	m, cur := scs.Grid.Update(cur, buf, p)
	n += m
	m, cur = scs.UserCursor.ApplyTo(cur, buf)
	n += m
	return n, cur
}

func clampToScreen(pt, size image.Point) image.Point {
	if pt.X < 1 {
		pt.X = 1
	} else if pt.X > size.X {
		pt.X = size.X
	}
	if pt.Y < 1 {
		pt.Y = 1
	} else if pt.Y > size.Y {
		pt.Y = size.Y
	}
	return pt
}

// ProcessRune updates the cursor position by the graphic width of the rune.
func (cs *CursorState) ProcessRune(r rune) {
	switch {
	case unicode.IsGraphic(r):
		cs.X++ // TODO support double-width runes
	// TODO anything for other control runes?
	case r == '\x0A': // LF
		cs.Y++
	case r == '\x0D': // CR
		cs.X = 0
	}
}

// ProcessEscape decodes cursor movement and attribute changes, updating state.
// Any errors decoding escape arguments are silenced, and the offending escape
// sequence(s) ignored.
func (cs *CursorState) ProcessEscape(e ansi.Escape, a []byte) {
	switch e {
	case ansi.CUU, ansi.CUD, ansi.CUF, ansi.CUB: // relative cursor motion
		b, _ := e.CSI()
		if d := cursorMoves[b-'A']; len(a) == 0 {
			cs.Point = cs.Point.Add(d)
		} else if n, _, err := ansi.DecodeNumber(a); err == nil {
			d = d.Mul(n)
			cs.Point = cs.Point.Add(d)
		}

	case ansi.CUP: // absolute cursor motion
		if len(a) == 0 {
			cs.Point = image.Pt(1, 1)
		} else if y, n, err := ansi.DecodeNumber(a); err == nil {
			if x, _, err := ansi.DecodeNumber(a[n:]); err == nil {
				cs.Point = image.Pt(x, y)
			}
		}

	case ansi.SGR:
		if attr, _, err := ansi.DecodeSGR(a); err == nil {
			cs.Attr = cs.Attr.Merge(attr)
		}

	case ansi.SM:
		// TODO better mode decoding (follow SGRAttr's example, and mature the ansi.Mode type)
		if len(a) > 1 && a[0] == '?' {
			n := 1
			for n < len(a) {
				mode, m, err := ansi.DecodeMode(true, a[n:])
				n += m
				if err != nil {
					return
				}
				switch mode {
				case ansi.ShowCursor: // TODO terminfo
					cs.Visible = true
				}
			}
		}

	case ansi.RM:
		if len(a) > 1 && a[0] == '?' {
			n := 1
			for n < len(a) {
				mode, m, err := ansi.DecodeMode(true, a[n:])
				n += m
				if err != nil {
					return
				}
				switch mode {
				case ansi.ShowCursor: // TODO terminfo
					cs.Visible = false
				}
			}
		}

	}
}

// ProcessRune sets the rune into the virtual screen grid.
func (scs *ScreenState) ProcessRune(r rune) {
	switch {
	case unicode.IsGraphic(r):
		scs.Grid.Cell(scs.Point).Set(r, scs.CursorState.Attr)
		if scs.X++; scs.X > scs.Size.X {
			scs.X = 1
			scs.linefeed()
		}
	case r == '\x0A':
		scs.linefeed()
	}
}

// ProcessEscape decodes cursor movement and attribute changes, updating cursor
// state, and decodes screen manipulation sequences, updating the virtual cell
// grid.  Any errors decoding escape arguments are silenced, and the offending
// escape sequence(s) ignored.
func (scs *ScreenState) ProcessEscape(e ansi.Escape, a []byte) {
	switch e {
	case ansi.CUU, ansi.CUD, ansi.CUF, ansi.CUB: // relative cursor motion
		b, _ := e.CSI()
		if d := cursorMoves[b-'A']; len(a) == 0 {
			scs.Point = clampToScreen(scs.Point.Add(d), scs.Size)
		} else if n, _, err := ansi.DecodeNumber(a); err == nil {
			d = d.Mul(n)
			scs.Point = clampToScreen(scs.Point.Add(d), scs.Size)
		}

	case ansi.CUP: // absolute cursor motion
		if len(a) == 0 {
			scs.Point = clampToScreen(image.Pt(1, 1), scs.Size)
		} else if y, n, err := ansi.DecodeNumber(a); err == nil {
			if x, _, err := ansi.DecodeNumber(a[n:]); err == nil {
				scs.Point = clampToScreen(image.Pt(x, y), scs.Size)
			}
		}

	case ansi.SGR:
		if attr, _, err := ansi.DecodeSGR(a); err == nil {
			scs.CursorState.Attr = scs.CursorState.Attr.Merge(attr)
		}

	case ansi.ED:
		var val byte
		if len(a) == 1 {
			val = a[0]
		} else {
			return
		}
		switch val {
		case '0': // Erase from current position to bottom of screen inclusive
			if i, ok := scs.index(scs.Point); ok {
				scs.ClearRegion(i+1, len(scs.Rune))
			}
		case '1': // Erase from top of screen to current position inclusive
			if i, ok := scs.index(scs.Point); ok {
				scs.ClearRegion(0, i+1)
			}
		case '2': // Erase entire screen (without moving the cursor)
			scs.ClearRegion(0, len(scs.Rune))
		}

	case ansi.EL:
		var val byte
		if len(a) == 1 {
			val = a[0]
		} else {
			return
		}

		var i, j int
		var iok, jok bool
		switch val {
		case '0': // Erase from current position to end of line inclusive
			i, iok = scs.index(scs.Point)
			j, jok = scs.index(image.Pt(scs.Size.X, scs.Y))
		case '1': // Erase from beginning of line to current position inclusive
			i, iok = scs.index(image.Pt(1, scs.Y))
			j, jok = scs.index(image.Pt(scs.X, scs.Y))
		case '2': // Erase entire line (without moving cursor)
			i, iok = scs.index(image.Pt(1, scs.Y))
			j, jok = scs.index(image.Pt(scs.Size.X, scs.Y))
		default:
			return
		}
		if iok && jok {
			scs.ClearRegion(i, j+1)
		}

		// case ansi.DECSTBM: TODO
		// [12;24r Set scrolling region to lines 12 thru 24.  If a linefeed or an
		//         INDex is received while on line 24, the former line 12 is
		//         deleted and rows 13-24 move up.  If a RI (reverse Index) is
		//         received while on line 12, a blank line is inserted there as
		//         rows 12-13 move down.  All VT100 compatible terminals (except
		//         GIGI) have this feature.
	}
}

func (scs *ScreenState) linefeed() {
	if scs.Y < scs.Size.Y {
		scs.Y++
	} else {
		scs.scrollBy(1)
	}
}

func (scs *ScreenState) scrollBy(n int) {
	i, ok := scs.index(scs.Point.Add(image.Pt(1, 2)))
	if !ok {
		return
	}
	for j := copy(scs.Grid.Rune, scs.Grid.Rune[i:]); j < len(scs.Grid.Rune); j++ {
		scs.Grid.Rune[j] = 0
	}
	for j := copy(scs.Grid.Attr, scs.Grid.Attr[i:]); j < len(scs.Grid.Attr); j++ {
		scs.Grid.Attr[j] = 0
	}
}
