package anansi

import (
	"fmt"
	"image"
	"io"
	"unicode"

	"github.com/jcorbin/anansi/ansi"
)

var cursorMoves = [4]image.Point{
	image.Pt(0, -1), // CUU ^[[A
	image.Pt(0, 1),  // CUD ^[[B
	image.Pt(1, 0),  // CUF ^[[C
	image.Pt(-1, 0), // CUB ^[[D
}

// Cursor represents terminal cursor state, including position, graphics
// attributes, and visibility.
//
// TODO more things like cursor shape and color
type Cursor struct {
	ansi.Point
	Attr    ansi.SGRAttr
	Visible bool

	attrKnown bool
	visKnown  bool
}

// Screen extends Cursor with a Grid of screen content.  allowing
// consumers to reason about the state of the terminal screen (virtual, real,
// or otherwise).
type Screen struct {
	Cursor Cursor
	Grid
}

// Full returns a shallow copy of the screen with the Grid restored to its full
// area.
func (sc Screen) Full() Screen {
	sc.Grid = sc.Grid.Full()
	return sc
}

// SubAt returns a shallow copy of the screen with a sub-Grid anchored at the
// given point; clamps the cursor to the new bounding rectangle.
func (sc Screen) SubAt(at ansi.Point) Screen {
	return sc.SubRect(ansi.Rectangle{Min: at, Max: sc.Rect.Max})
}

// SubSize returns a shallow copy of the screen with a sub-Grid resized to the
// given size; clamps the cursor to the new bounding rectangle.
func (sc Screen) SubSize(sz image.Point) Screen {
	return sc.SubRect(ansi.Rectangle{Min: sc.Rect.Min, Max: sc.Rect.Min.Add(sz)})
}

// SubRect returns a shallow copy of the screen with a sub-Grid bounded by the
// given rectangle; clamps the cursor to the new bounding rectangle.
func (sc Screen) SubRect(r ansi.Rectangle) Screen {
	sc.Grid = sc.Grid.SubRect(r)
	sc.Cursor.Point = clampPointTo(sc.Cursor.Point, r)
	return sc
}

func (cs Cursor) String() string {
	return fmt.Sprintf("@%v a:%v v:%v", cs.Point, cs.Attr, cs.Visible)
}

func (sc Screen) String() string {
	return fmt.Sprintf("%v gridBounds:%v", sc.Cursor, sc.Grid.Bounds())
}

// Clear the screen grid, and reset cursor state (to invisible nowhere).
func (sc *Screen) Clear() {
	sc.Grid.Clear()
	sc.Cursor = Cursor{}
}

// Resize the underlying Grid, and zero the cursor position if out of bounds.
// Returns true only if the resize was a change, false if it was a no-op.
func (sc *Screen) Resize(size image.Point) bool {
	if sc.Grid.Resize(size) {
		if !sc.Cursor.Point.In(sc.Bounds()) {
			sc.Cursor.Point.Point = image.ZP
		}
		return true
	}
	return false
}

// Show returns the control sequence necessary to show the cursor if it is not
// visible, the zero sequence otherwise.
func (cs *Cursor) Show() ansi.Seq {
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
func (cs *Cursor) Hide() ansi.Seq {
	// TODO terminfo
	if !cs.visKnown || cs.Visible {
		cs.visKnown = true
		cs.Visible = false
		return ansi.ShowCursor.Reset()
	}
	return ansi.Seq{}
}

// MergeSGR merges the given SGR attribute into Attr, returning the difference.
func (cs *Cursor) MergeSGR(attr ansi.SGRAttr) ansi.SGRAttr {
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
func (cs *Cursor) To(pt ansi.Point) ansi.Seq {
	if !cs.Point.Valid() {
		// TODO more nuanced: if X in unknown / Y is unknown?
		cs.X, cs.Y = pt.X, pt.Y
		return ansi.CUP.WithPoint(pt)
	}
	if pt.Y == cs.Y+1 && pt.X == 1 {
		return cs.NewLine()
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

// NewLine moves the cursor to the start of the next line, returning the
// necessary ansi sequence.
func (cs *Cursor) NewLine() ansi.Seq {
	cs.Y++
	cs.X = 1
	return ansi.Escape('\r').With('\n')
}

func (sc *Screen) clamp(pt ansi.Point) ansi.Point {
	r := sc.Bounds()
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
func (sc *Screen) To(pt ansi.Point) {
	sc.Cursor.Point = sc.clamp(pt)
}

// ApplyTo applies the receiver cursor state into the passed state value,
// writing any necessary control sequences into the provided buffer. Returns
// the number of bytes written, and the updated cursor state.
func (cs *Cursor) ApplyTo(w io.Writer, cur Cursor) (int, Cursor, error) {
	return withAnsiCursorWriter(w, cur, func(aw ansiWriter, cur Cursor) (int, Cursor) {
		return cs.applyTo(aw, cur)
	})
}

func (cs *Cursor) applyTo(aw ansiWriter, cur Cursor) (n int, _ Cursor) {
	if cs.Visible && cs.Point.Valid() {
		n += aw.WriteSeq(cur.To(cs.Point))
		n += aw.WriteSGR(cur.MergeSGR(cs.Attr))
		n += aw.WriteSeq(cur.Show())
	} else {
		n += aw.WriteSeq(cur.Hide())
	}
	return n, cur
}

// Update syncs screen state from the receiver against the given prior screen
// state, generating writing any/all necessary ansi control sequences into the
// given writer. Returns the number of bytes written and the final screen state,
// which will now equal the receiver state.
func (sc *Screen) Update(w io.Writer, prior Screen) (int, Screen, error) {
	return withAnsiScreenWriter(w, prior, sc.update)
}

func (sc *Screen) update(aw ansiWriter, prior Screen) (int, Screen) {
	var n, m int
	n += aw.WriteSeq(prior.Cursor.Hide())
	m, prior = writeGrid(aw, sc.Grid, prior, NoopStyle)
	n += m
	m, prior.Cursor = sc.Cursor.applyTo(aw, prior.Cursor)
	n += m
	prior.Resize(sc.Grid.Bounds().Size())
	copy(prior.Rune, sc.Rune)
	copy(prior.Attr, sc.Attr)
	return n, prior
}

// ProcessANSI updates cursor state to reflect having written the given escape
// value or rune to a terminal.
//
// Graphic runes advance the cursor X position.
//
// Supported escape sequences:
//   - CUU, CUD, CUF, and CUB all relatively update Point
//   - CUP sets Point absolutely
//   - SGR merges into Attr (see SGRAttr.Merge)
//   - SM and RM implement modes:
//     - private mode 25 updates Visible
//
// Any errors decoding escape arguments are silenced, and the offending
// escape sequence(s) ignored.
func (cs *Cursor) ProcessANSI(e ansi.Escape, a []byte) {
	switch {
	case e.IsEscape():
		cs.processEscape(e, a, ptID)
	case e == '\x0A': // LF
		cs.Y++
	case e == '\x0D': // CR
		cs.X = 1
	// TODO anything for other control runes?
	case unicode.IsGraphic(rune(e)):
		cs.X++ // TODO support double-width runes
	}
}

func ptID(pt ansi.Point) ansi.Point { return pt }

// processEscape implements cursor escape processing shared with ScreenState,
// which passes a non-identity clamp function.
func (cs *Cursor) processEscape(
	e ansi.Escape, a []byte,
	clamp func(pt ansi.Point) ansi.Point,
) {
	switch e {
	case ansi.CUU, ansi.CUD, ansi.CUF, ansi.CUB: // relative cursor motion
		b, _ := e.CSI()
		d := cursorMoves[b-'A']
		if len(a) > 0 {
			n, _, err := ansi.DecodeNumber(a)
			if err != nil {
				return
			}
			d = d.Mul(n)
		}
		cs.Point = clamp(cs.Point.Add(d))

	case ansi.CUP: // absolute cursor motion
		p := ansi.Pt(1, 1)
		if len(a) > 0 {
			var err error
			p, _, err = ansi.DecodePoint(a)
			if err != nil {
				return
			}
		}
		cs.Point = clamp(p)

	case ansi.SGR:
		if attr, _, err := ansi.DecodeSGR(a); err == nil {
			cs.Attr = cs.Attr.Merge(attr)
		}

	case ansi.SM:
		// TODO better mode decoding (follow SGRAttr's example, and mature the ansi.Mode type)
		// TODO better mode processing: injected handler for ScreenState
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

// ProcessANSI updates screen state to reflect having written the given escape
// value or rune to a terminal; in addition to CursorState.ProcessANSI semantics:
//
// Graphic runes update the virtual cell grid, using the current cursor SGR
// attribute, at the current cursor point.
//
// Supported escape sequences:
//   - ED to erase display
//   - EL to erase line
//   - cursor movement sequences, as per CursorState.ProcessANSI, but clamped
//     to the screen bounds
//
// Any errors decoding escape arguments are silenced, and the offending
// escape sequence(s) ignored.
func (sc *Screen) ProcessANSI(e ansi.Escape, a []byte) {
	if sc.Cursor.Point.Point == image.ZP {
		sc.Cursor.Point = ansi.Pt(1, 1)
	}
	switch {
	case e.IsEscape():
		sc.processEscape(e, a)
	case e == '\x0A': // LF
		sc.linefeed()
	case e == '\x0D': // CR
		sc.Cursor.X = 1
	// TODO anything for other control runes?
	case unicode.IsGraphic(rune(e)):
		br := sc.Bounds()
		if i, ok := sc.Grid.CellOffset(sc.Cursor.Point); ok {
			sc.Grid.Rune[i], sc.Grid.Attr[i] = rune(e), sc.Cursor.Attr
		}
		if sc.Cursor.X++; sc.Cursor.X >= br.Max.X {
			sc.Cursor.X = br.Min.X
			sc.linefeed()
		}
	}
}

func (sc *Screen) processEscape(e ansi.Escape, a []byte) {
	switch e {
	case ansi.ED:
		var val byte
		if len(a) == 1 {
			val = a[0]
		} else {
			return
		}
		switch val {
		case '0': // Erase from current position to bottom of screen inclusive
			if i, ok := sc.CellOffset(sc.Cursor.Point); ok {
				sc.clearRegion(i+1, len(sc.Rune))
			}
		case '1': // Erase from top of screen to current position inclusive
			if i, ok := sc.CellOffset(sc.Cursor.Point); ok {
				sc.clearRegion(0, i+1)
			}
		case '2': // Erase entire screen (without moving the cursor)
			sc.clearRegion(0, len(sc.Rune))
		}

	case ansi.EL:
		var val byte
		if len(a) == 1 {
			val = a[0]
		} else {
			return
		}

		lo := sc.Cursor.Point
		hi := sc.Bounds().Max.Sub(image.Pt(1, 1))
		switch val {
		case '0': // Erase from current position to end of line inclusive
			hi.Y = sc.Cursor.Y
		case '1': // Erase from beginning of line to current position inclusive
			lo.X = 1
			hi = sc.Cursor.Point
		case '2': // Erase entire line (without moving cursor)
			lo.X = 1
			hi.Y = sc.Cursor.Y
		default:
			return
		}

		i, iok := sc.CellOffset(lo)
		j, jok := sc.CellOffset(hi)
		if iok && jok {
			sc.clearRegion(i, j+1)
		}

		// case ansi.DECSTBM: TODO
		// [12;24r Set scrolling region to lines 12 thru 24.  If a linefeed or an
		//         INDex is received while on line 24, the former line 12 is
		//         deleted and rows 13-24 move up.  If a RI (reverse Index) is
		//         received while on line 12, a blank line is inserted there as
		//         rows 12-13 move down.  All VT100 compatible terminals (except
		//         GIGI) have this feature.

	default:
		sc.Cursor.processEscape(e, a, sc.clamp)
	}
}

func (sc *Screen) clearRegion(i, max int) {
	for ; i < max; i++ {
		sc.Grid.Rune[i] = 0
		sc.Grid.Attr[i] = 0
	}
}

func (sc *Screen) linefeed() {
	if sc.Cursor.Y+1 < sc.Bounds().Max.Y {
		sc.Cursor.Y++
	} else {
		sc.scrollBy(1)
	}
}

func (sc *Screen) scrollBy(n int) {
	i, ok := sc.CellOffset(sc.Cursor.Point.Add(image.Pt(1, 2)))
	if !ok {
		return
	}
	for j := copy(sc.Grid.Rune, sc.Grid.Rune[i:]); j < len(sc.Grid.Rune); j++ {
		sc.Grid.Rune[j] = 0
	}
	for j := copy(sc.Grid.Attr, sc.Grid.Attr[i:]); j < len(sc.Grid.Attr); j++ {
		sc.Grid.Attr[j] = 0
	}
}

var (
	_ Processor = &Cursor{}
	_ Processor = &Screen{}
)

// TODO should this by a method or utility in ansi/geom.go?
func clampPointTo(p ansi.Point, r ansi.Rectangle) ansi.Point {
	if p.X < r.Min.X {
		p.X = r.Min.X
	} else if p.X >= r.Max.X {
		p.X = r.Max.X - 1
	}
	if p.Y < r.Min.Y {
		p.Y = r.Min.Y
	} else if p.Y >= r.Max.Y {
		p.Y = r.Max.Y - 1
	}
	return p
}
