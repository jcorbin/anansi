package anansi

import (
	"image"
	"unicode"

	"github.com/jcorbin/anansi/ansi"
)

var cursorMoves = [4]image.Point{
	image.Pt(-1, 0), // CUU ^[[A
	image.Pt(1, 0),  // CUD ^[[B
	image.Pt(0, 1),  // CUF ^[[C
	image.Pt(0, -1), // CUB ^[[D
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
		attr |= ansi.SGRAttrClear
		cs.attrKnown = true
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

// ProcessRune updates the cursor position by the graphic width of the rune.
func (cs *CursorState) ProcessRune(r rune) {
	switch {
	case unicode.IsGraphic(r):
		cs.X++ // TODO support double-width runes
	// TODO anything for other control runes?
	case r == '\x0A': // LF
		cs.Y++
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
