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
	ansi.Point
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
	return fmt.Sprintf("%v uc:(%v) gridBounds:%v", scs.CursorState, scs.UserCursor, scs.Grid.Bounds())
}

// Clear the screen grid, and reset the UserCursor (to invisible nowhere).
func (scs *ScreenState) Clear() {
	for i := range scs.Grid.Rune {
		scs.Grid.Rune[i] = 0
		scs.Grid.Attr[i] = 0
	}
	scs.Point.Point = image.ZP
	scs.CursorState.Attr = 0
	scs.UserCursor = CursorState{}
}

// Resize the underlying Grid, and zero the cursor position if out of bounds.
// Returns true only if the resize was a change, false if it was a no-op.
func (scs *ScreenState) Resize(size image.Point) bool {
	if scs.Grid.Resize(size) {
		if !scs.Point.In(scs.Bounds()) {
			scs.Point.Point = image.ZP
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
// given screen point using absolute (ansi.CUP) or relative
// (ansi.{CUU,CUD,CUF,CUD}) if possible. Returns a zero sequence if the cursor
// is already at the given point.
func (cs *CursorState) To(pt ansi.Point) ansi.Seq {
	if !cs.Point.Valid() {
		// TODO more nuanced: if X in unknown / Y is unknown?
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.CUP.WithPoint(pt)
	}
	if pt.Y == cs.Y+1 && pt.X == 1 {
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.Escape('\r').With('\n')
	}
	if pt.X != cs.X && pt.Y != cs.Y {
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.CUP.WithPoint(pt)
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

func (scs *ScreenState) clamp(pt ansi.Point) ansi.Point {
	r := scs.Bounds()
	if pt.X < r.Min.X {
		pt.X = r.Min.X
	} else if pt.X >= r.Max.X {
		pt.X = r.Max.X - 1
	}
	if pt.Y < r.Min.Y {
		pt.Y = r.Min.Y
	} else if pt.Y >= r.Max.Y {
		pt.Y = r.Max.Y - 1
	}
	return pt
}

// To sets the virtual cursor point to the supplied one.
func (scs *ScreenState) To(pt ansi.Point) {
	scs.Point = scs.clamp(pt)
}

// ApplyTo applies the receiver cursor state into the passed state value,
// writing any necessary control sequences into the provided buffer. Returns
// the number of bytes written, and the updated cursor state.
func (cs *CursorState) ApplyTo(cur CursorState, buf *ansi.Buffer) (n int, _ CursorState) {
	if cs.Visible && cs.Point.Valid() {
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
func (scs *ScreenState) Update(cur CursorState, buf *ansi.Buffer, prior Grid) (n int, _ CursorState) {
	n += buf.WriteSeq(cur.Hide())
	m, cur := RenderGrid(buf, cur, scs.Grid, prior)
	n += m
	m, cur = scs.UserCursor.ApplyTo(cur, buf)
	n += m
	return n, cur
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
			cs.Point = ansi.Pt(1, 1)
		} else if p, _, err := ansi.DecodePoint(a); err == nil {
			cs.Point = p
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
	br := scs.Bounds()
	switch {
	case unicode.IsGraphic(r):
		if i, ok := scs.Grid.CellOffset(scs.Point); ok {
			scs.Grid.Rune[i], scs.Grid.Attr[i] = r, scs.CursorState.Attr
		}
		if scs.X++; scs.X >= br.Max.X {
			scs.X = br.Min.X
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
			scs.Point = scs.clamp(scs.Point.Add(d))
		} else if n, _, err := ansi.DecodeNumber(a); err == nil {
			d = d.Mul(n)
			scs.Point = scs.clamp(scs.Point.Add(d))
		}

	case ansi.CUP: // absolute cursor motion
		if len(a) == 0 {
			scs.Point = scs.clamp(ansi.Pt(1, 1))
		} else if p, _, err := ansi.DecodePoint(a); err == nil {
			scs.Point = scs.clamp(p)
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
			if i, ok := scs.CellOffset(scs.Point); ok {
				scs.clearRegion(i+1, len(scs.Rune))
			}
		case '1': // Erase from top of screen to current position inclusive
			if i, ok := scs.CellOffset(scs.Point); ok {
				scs.clearRegion(0, i+1)
			}
		case '2': // Erase entire screen (without moving the cursor)
			scs.clearRegion(0, len(scs.Rune))
		}

	case ansi.EL:
		var val byte
		if len(a) == 1 {
			val = a[0]
		} else {
			return
		}

		lo := scs.Point
		hi := scs.Bounds().Max.Sub(image.Pt(1, 1))
		switch val {
		case '0': // Erase from current position to end of line inclusive
			hi.Y = scs.Y
		case '1': // Erase from beginning of line to current position inclusive
			lo.X = 1
			hi = scs.Point
		case '2': // Erase entire line (without moving cursor)
			lo.X = 1
			hi.Y = scs.Y
		default:
			return
		}

		i, iok := scs.CellOffset(lo)
		j, jok := scs.CellOffset(hi)
		if iok && jok {
			scs.clearRegion(i, j+1)
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

func (scs *ScreenState) clearRegion(i, max int) {
	for ; i < max; i++ {
		scs.Grid.Rune[i] = 0
		scs.Grid.Attr[i] = 0
	}
}

func (scs *ScreenState) linefeed() {
	if scs.Y+1 < scs.Bounds().Max.Y {
		scs.Y++
	} else {
		scs.scrollBy(1)
	}
}

func (scs *ScreenState) scrollBy(n int) {
	i, ok := scs.CellOffset(scs.Point.Add(image.Pt(1, 2)))
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
