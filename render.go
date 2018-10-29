package anansi

import (
	"github.com/jcorbin/anansi/ansi"
)

// RenderGrid writes a grid's contents into an ansi buffer, relative to current
// cursor state and any prior screen contents. To force an absolute
// (non-differential) update, pass an empty prior grid.
// Returns the number of bytes written into the buffer and the final cursor
// state.
func RenderGrid(buf *ansi.Buffer, cur CursorState, g, prior Grid, styles ...Style) (int, CursorState) {
	if g.Stride != g.Rect.Dx() {
		panic("sub-grid update not implemented")
	}
	if g.Rect.Min != ansi.Pt(1, 1) {
		panic("sub-screen update not implemented")
	}
	style := Styles(styles...)
	if len(g.Attr) == 0 || len(g.Rune) == 0 {
		return 0, cur
	}
	if len(prior.Attr) == 0 || len(prior.Rune) == 0 || prior.Rect.Empty() || !prior.Rect.Eq(g.Rect) {
		return renderGrid(buf, cur, g, style)
	}
	return renderGridDiff(buf, cur, g, prior, style)
}

func renderGrid(buf *ansi.Buffer, cur CursorState, g Grid, style Style) (int, CursorState) {
	const empty = ' '
	if fillRune, _ := style.Style(ansi.ZP, 0, 0, 0, 0); fillRune == empty {
		style = Styles(style, ZeroRuneStyle(empty))
	}
	style = Styles(style, DefaultRuneStyle(empty))
	n := buf.WriteSeq(ansi.ED.With('2'))
	for i, pt := 0, ansi.Pt(1, 1); i < len(g.Rune); {
		if gr, ga := style.Style(pt, 0, g.Rune[i], 0, g.Attr[i]); gr != 0 {
			mv := cur.To(pt)
			ad := cur.MergeSGR(ga)
			n += buf.WriteSeq(mv)
			n += buf.WriteSGR(ad)
			m, _ := buf.WriteRune(gr)
			n += m
			cur.ProcessRune(gr)
		}
		i++
		if pt.X++; pt.X >= g.Rect.Max.X {
			pt.X = g.Rect.Min.X
			pt.Y++
		}
	}
	return n, cur
}

func renderGridDiff(buf *ansi.Buffer, cur CursorState, g, prior Grid, style Style) (int, CursorState) {
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
			n += buf.WriteSeq(mv)
			n += buf.WriteSGR(ad)
			m, _ := buf.WriteRune(gr)
			n += m
			cur.ProcessRune(gr)
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

// RenderBitmap writes a bitmap's contents as braille runes into an ansi buffer.
// Optional style(s) may be passed to control graphical rendition of the
// braille runes.
func RenderBitmap(buf *ansi.Buffer, bi *Bitmap, styles ...Style) {
	style := Styles(styles...)
	for bp := bi.Rect.Min; bp.Y < bi.Rect.Max.Y; bp.Y += 4 {
		if bp.Y > 0 {
			buf.WriteByte('\n')
		}
		for bp.X = bi.Rect.Min.X; bp.X < bi.Rect.Max.X; bp.X += 2 {
			sp := ansi.PtFromImage(bp)
			sr := bi.Rune(bp)
			if r, a := style.Style(sp, 0, sr, 0, 0); r != 0 {
				if a != 0 {
					buf.WriteSGR(a)
				}
				buf.WriteRune(r)
			} else {
				buf.WriteRune(' ')
			}
		}
	}
}
