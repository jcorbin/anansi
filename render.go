package anansi

import (
	"io"

	"github.com/jcorbin/anansi/ansi"
)

// WriteGridUpdate writes a grid's contents into an io.Writer, relative to
// current cursor state and any prior screen contents. To force an absolute
// (non-differential) update, pass an empty prior grid. Returns the number of
// bytes written, final cursor state, and any write error encountered.
func WriteGridUpdate(w io.Writer, cur CursorState, g, prior Grid, styles ...Style) (int, CursorState, error) {
	if g.Stride != g.Rect.Dx() {
		panic("sub-grid update not implemented")
	}
	if g.Rect.Min != ansi.Pt(1, 1) {
		panic("sub-screen update not implemented")
	}
	return withAnsiWriter(w, cur, func(aw ansiWriter, cur CursorState) (int, CursorState) {
		style := Styles(styles...)
		return writeGridUpdate(aw, cur, g, prior, style)
	})
}

func writeGridUpdate(aw ansiWriter, cur CursorState, g, prior Grid, style Style) (int, CursorState) {
	n := 0
	if len(g.Attr) > 0 && len(g.Rune) > 0 {
		if len(prior.Attr) == 0 || len(prior.Rune) == 0 || prior.Rect.Empty() || !prior.Rect.Eq(g.Rect) {
			n, cur = writeGridFull(aw, cur, g, style)
		} else {
			n, cur = writeGridDiff(aw, cur, g, prior, style)
		}
	}
	return n, cur
}

func writeGridFull(aw ansiWriter, cur CursorState, g Grid, style Style) (int, CursorState) {
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

func writeGridDiff(aw ansiWriter, cur CursorState, g, prior Grid, style Style) (int, CursorState) {
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
			mv := cur.To(pt)
			ad := cur.MergeSGR(ga)
			n += aw.WriteSeq(mv)
			n += aw.WriteSGR(ad)
			m, _ := aw.WriteRune(gr)
			n += m
			cur.ProcessANSI(ansi.Escape(gr), nil)
		}

	next:
		i++
		if pt.X++; pt.X >= g.Rect.Max.X {
			pt.X = g.Rect.Min.X
			pt.Y++
		}
	}
	return n, cur
}

// WriteBitmap writes a bitmap's contents as braille runes into an io.Writer.
// Optional style(s) may be passed to control graphical rendition of the
// braille runes.
func WriteBitmap(w io.Writer, bi Bitmap, styles ...Style) (int, error) {
	// TODO deal with CursorState?
	n, _, err := withAnsiWriter(w, CursorState{}, func(aw ansiWriter, cur CursorState) (int, CursorState) {
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

func writeBitmap(aw ansiWriter, cur CursorState, bi Bitmap, style Style) (n int, _ CursorState) {
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

func withAnsiWriter(
	w io.Writer, cur CursorState,
	f func(aw ansiWriter, cur CursorState) (int, CursorState),
) (n int, _ CursorState, err error) {
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

type ansiWriter interface {
	io.Writer

	WriteESC(seqs ...ansi.Escape) int
	WriteSeq(seqs ...ansi.Seq) int
	WriteSGR(attrs ...ansi.SGRAttr) int

	WriteString(s string) (n int, err error)
	WriteRune(r rune) (n int, err error)
	WriteByte(c byte) error
}
