package anansi

import (
	"image"

	"github.com/jcorbin/anansi/ansi"
)

// TODO grid composition/copy function

// Grid is a grid of screen cells.
type Grid struct {
	Size image.Point
	Attr []ansi.SGRAttr
	Rune []rune
	// TODO []string for multi-rune glyphs
}

// Resize the grid to have room for n cells.
// Returns true only if the resize was a change, false if it was a no-op.
func (g *Grid) Resize(size image.Point) bool {
	if size == g.Size {
		return false
	}
	n := size.X * size.Y
	for n > cap(g.Attr) {
		g.Attr = append(g.Attr, 0)
	}
	for n > cap(g.Rune) {
		g.Rune = append(g.Rune, 0)
	}
	g.Attr = g.Attr[:n]
	g.Rune = g.Rune[:n]
	g.Size = size
	return true
}

// Bounds returns the bounding rectangle of the grid in cell space: 1,1 origin,
// with max of Size+1.
func (g *Grid) Bounds() image.Rectangle {
	return image.Rectangle{image.Pt(1, 1), g.Size.Add(image.Pt(1, 1))}
}

// CellOffset returns the offset of the screen cell and true if it's
// within the Grid's Bounds().
func (g *Grid) CellOffset(pt image.Point) (int, bool) {
	if !pt.In(image.Rect(1, 1, g.Size.X+1, pt.Y+1)) {
		return 0, false
	}
	p := pt.Sub(image.Pt(1, 1)) // convert to normal 0-indexed point
	return p.Y*g.Size.X + p.X, true
}

// Update writes the escape sequences and runes into the given buffer necessary
// to affect the receiver Grid's state, relative to the given cursor state, and
// any prior Grid state. If the prior is empty, then a full display erase and
// redraw is done. Returns the number of bytes written into the buffer, and the
// final cursor state.
func (g *Grid) Update(cur CursorState, buf *ansi.Buffer, prior *Grid) (n int, _ CursorState) {
	if len(g.Attr) == 0 || len(g.Rune) == 0 {
		return n, cur
	}
	diffing := true
	if len(prior.Attr) == 0 || len(prior.Rune) == 0 || prior.Size == image.ZP || prior.Size != g.Size {
		diffing = false
		n += buf.WriteSeq(ansi.ED.With('2'))
	}

	for i, pt := 0, image.Pt(1, 1); i < len(g.Rune); /* next: */ {
		gr, ga := g.Rune[i], g.Attr[i]

		if diffing {
			if j, ok := prior.CellOffset(pt); !ok {
				diffing = false // out-of-bounds disengages diffing
			} else {
				pr, pa := prior.Rune[j], prior.Attr[j] // NOTE range ok since pt <= prior.Size
				if gr == 0 {
					gr, ga = ' ', 0
				}
				if pr == 0 {
					pr, pa = ' ', 0
				}
				if gr == pr && ga == pa {
					goto next // continue
				}
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
		if pt.X++; pt.X > g.Size.X {
			pt.X = 1
			pt.Y++
		}
	}
	return n, cur
}
