package anansi

import (
	"io"

	"github.com/jcorbin/anansi/ansi"
)

// WriteGrid writes a grid's contents into an io.Writer, relative to current
// cursor state and any prior screen contents.
//
// To force an absolute (non-differential) update, pass an empty prior grid.
// Returns the number of bytes written, final cursor state, and any write error
// encountered.
func WriteGrid(w io.Writer, g Grid, prior Screen, styles ...Style) (int, Screen, error) {
	if g.Stride != g.Rect.Dx() {
		panic("sub-grid update not implemented")
	}
	if g.Rect.Min != ansi.Pt(1, 1) {
		panic("sub-screen update not implemented")
	}
	return withAnsiScreenWriter(w, prior, func(aw ansiWriter, sc Screen) (int, Screen) {
		return writeGrid(aw, g, sc, Styles(styles...))
	})
}

func writeGrid(aw ansiWriter, g Grid, prior Screen, style Style) (int, Screen) {
	if len(g.Attr) == 0 || len(g.Rune) == 0 {
		return 0, prior
	}
	if len(prior.Attr) == 0 || len(prior.Rune) == 0 || prior.Rect.Empty() || !prior.Rect.Eq(g.Rect) {
		var n int
		n, prior.Cursor = writeGridFull(aw, prior.Cursor, g, style)
		return n, prior
	}
	return writeGridDiff(aw, g, prior, style)
}

func writeGridFull(aw ansiWriter, cur Cursor, g Grid, style Style) (int, Cursor) {
	const empty = ' '
	if fillRune, _ := style.Style(ansi.ZP, 0, 0, 0, 0); fillRune == empty {
		style = Styles(style, ZeroRuneStyle(empty))
	}
	style = Styles(style, DefaultRuneStyle(empty))
	n := aw.WriteSeq(ansi.ED.With('2'))
	for i, pt := 0, ansi.Pt(1, 1); i < len(g.Rune); {
		if gr, ga := style.Style(pt, 0, g.Rune[i], 0, g.Attr[i]); gr != 0 {
			mv := cur.To(pt)
			ad := cur.MergeSGR(ga)
			n += aw.WriteSeq(mv)
			n += aw.WriteSGR(ad)
			m, _ := aw.WriteRune(gr)
			n += m
			cur.ProcessANSI(ansi.Escape(gr), nil)
		}
		i++
		if pt.X++; pt.X >= g.Rect.Max.X {
			pt.X = g.Rect.Min.X
			pt.Y++
		}
	}
	return n, cur
}

func writeGridDiff(aw ansiWriter, g Grid, prior Screen, style Style) (int, Screen) {
	fillRune, fillAttr := style.Style(ansi.ZP, 0, 0, 0, 0)
	const empty = ' '
	if fillRune == 0 {
		fillRune = empty
		style = Styles(style, FillRuneStyle(fillRune))
	}
	style = Styles(style, DefaultRuneStyle(empty))
	n, diffing := 0, true
	for i, pt := 0, ansi.Pt(1, 1); i < len(g.Rune); /* next: */ {
		pr, pa := fillRune, fillAttr
		gr, ga := style.Style(pt, pr, g.Rune[i], pa, g.Attr[i])

		if diffing {
			if j, ok := prior.CellOffset(pt); ok {
				// NOTE range ok since pt <= prior.Size
				if pr = prior.Rune[j]; pr == 0 {
					pr = fillRune
				}
				if pa = prior.Attr[j]; pa == 0 {
					pa = fillAttr
				}
				if gr == pr && ga == pa {
					goto next // continue
				}
			} else {
				diffing = false // out-of-bounds disengages diffing
			}
		}

		if gr != 0 {
			mv := prior.Cursor.To(pt)
			ad := prior.Cursor.MergeSGR(ga)
			n += aw.WriteSeq(mv)
			n += aw.WriteSGR(ad)
			m, _ := aw.WriteRune(gr)
			n += m
			prior.Cursor.ProcessANSI(ansi.Escape(gr), nil)
		}

	next:
		i++
		if pt.X++; pt.X >= g.Rect.Max.X {
			pt.X = g.Rect.Min.X
			pt.Y++
		}
	}
	return n, prior
}

// WriteBitmap writes a bitmap's contents as braille runes into an io.Writer.
// Optional style(s) may be passed to control graphical rendition of the
// braille runes.
func WriteBitmap(w io.Writer, bi Bitmap, styles ...Style) (int, error) {
	// TODO deal with Cursor?
	n, _, err := withAnsiCursorWriter(w, Cursor{}, func(aw ansiWriter, cur Cursor) (int, Cursor) {
		style := Styles(styles...)
		style = Styles(style, StyleFunc(func(p ansi.Point, _ rune, r rune, _ ansi.SGRAttr, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
			if r == 0 {
				return ' ', 0
			}
			return r, a
		}))
		return writeBitmap(aw, cur, bi, style)
	})
	return n, err
}

func writeBitmap(aw ansiWriter, cur Cursor, bi Bitmap, style Style) (n int, _ Cursor) {
	for bp := bi.Rect.Min; bp.Y < bi.Rect.Max.Y; bp.Y += 4 {
		if bp.Y > 0 {
			n += aw.WriteSeq(cur.NewLine())
		}
		for bp.X = bi.Rect.Min.X; bp.X < bi.Rect.Max.X; bp.X += 2 {
			sp := ansi.PtFromImage(bp)
			sr := bi.Rune(bp)
			r, a := style.Style(sp, 0, sr, 0, 0)

			if a != 0 {
				ad := cur.MergeSGR(a)
				n += aw.WriteSGR(ad)
			}

			m, _ := aw.WriteRune(r)
			n += m

			cur.ProcessANSI(ansi.Escape(r), nil)
		}
	}
	return n, cur
}

func withAnsiCursorWriter(
	w io.Writer, cur Cursor,
	f func(aw ansiWriter, cur Cursor) (int, Cursor),
) (n int, _ Cursor, err error) {
	aw, ok := w.(ansiWriter)
	if !ok {
		var buf Buffer
		defer func() {
			var m int64
			m, err = buf.WriteTo(w)
			n = int(m)
		}()
		aw = &buf
	}
	n, cur = f(aw, cur)
	return n, cur, nil
}

func withAnsiScreenWriter(
	w io.Writer, sc Screen,
	f func(aw ansiWriter, sc Screen) (int, Screen),
) (n int, _ Screen, err error) {
	n, sc.Cursor, err = withAnsiCursorWriter(w, sc.Cursor, func(aw ansiWriter, cur Cursor) (int, Cursor) {
		sc.Cursor = cur
		n, sc = f(aw, sc)
		return n, sc.Cursor
	})
	return n, sc, err
}

type ansiWriter interface {
	io.Writer

	WriteESC(seqs ...ansi.Escape) int
	WriteSeq(seqs ...ansi.Seq) int
	WriteSGR(attrs ...ansi.SGRAttr) int

	WriteString(s string) (n int, err error)
	WriteRune(r rune) (n int, err error)
	WriteByte(c byte) error
}
