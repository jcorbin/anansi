package anansi

import (
	"image"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

// Grid is a grid of screen cells.
type Grid struct {
	Size image.Point
	Attr []ansi.SGRAttr
	Rune []rune
	// TODO []string for multi-rune glyphs
}

// Clear the grid, setting all runes and attrs to 0.
func (g *Grid) Clear() {
	for i := range g.Rune {
		g.Rune[i] = 0
		g.Attr[i] = 0
	}
}

// ClearRegion within the grid, setting all runes and attrs within it to 0.
func (g *Grid) ClearRegion(i, max int) {
	for ; i < max; i++ {
		g.Rune[i] = 0
		g.Attr[i] = 0
	}
}

// Resize the grid to have room for n cells.
// Returns true only if the resize was a change, false if it was a no-op.
func (g *Grid) Resize(size image.Point) bool {
	if size != g.Size {
		g.Size = size
		n := size.X * size.Y
		for n > cap(g.Attr) {
			g.Attr = append(g.Attr, 0)
		}
		for n > cap(g.Rune) {
			g.Rune = append(g.Rune, 0)
		}
		g.Attr = g.Attr[:n]
		g.Rune = g.Rune[:n]
		return true
	}
	return false
}

// CopyTo resizes the dest grid to match the receiver, and copies all receiver
// rune and attr data into it.
func (g *Grid) CopyTo(dest *Grid) {
	dest.Resize(g.Size)
	copy(dest.Rune, g.Rune)
	copy(dest.Attr, g.Attr)
}

// Set the rune and attribute for the given x,y cell; silently ignores
// out-of-bounds points.
func (g *Grid) Set(pt image.Point, r rune, attr ansi.SGRAttr) {
	if i, ok := g.index(pt); ok {
		g.Rune[i] = r
		g.Attr[i] = attr
	}
}

// Get the rune and attribute set for the given x,y cell; always returns 0 for
// out-of-bounds points.
func (g *Grid) Get(pt image.Point) (rune, ansi.SGRAttr) {
	if i, ok := g.index(pt); ok {
		return g.Rune[i], g.Attr[i]
	}
	return 0, 0
}

func (g *Grid) index(pt image.Point) (int, bool) {
	if pt.X < 1 || pt.X > g.Size.X ||
		pt.Y < 1 || pt.Y > g.Size.Y {
		return 0, false
	}
	return (pt.Y-1)*g.Size.X + pt.X - 1, true
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
	if len(prior.Attr) == 0 || len(prior.Rune) == 0 || prior.Size != g.Size {
		diffing = false
		n += buf.WriteSeq(ansi.ED.With('2'))
	}

	var lastUpdate image.Point

	for i, pt := 0, image.Pt(1, 1); i < len(g.Rune); /* next: */ {
		gr, ga := g.Rune[i], g.Attr[i]

		if diffing {
			if pt.Y > prior.Size.Y {
				diffing = false // nothing left to diff with
			} else if pt.X <= prior.Size.X {
				j := prior.Size.X*(pt.Y-1) + pt.X - 1
				if gr == prior.Rune[j] && ga == prior.Attr[j] { // NOTE range ok since pt < prior.Size
					goto next // continue
				}
				if gr == 0 {
					gr, ga = ' ', 0
				}
				if gr == prior.Rune[j] && ga == prior.Attr[j] {
					goto next // continue
				}
			}
		}

		if gr != 0 {

			// check to see if we're indifferent to just writing the runes,
			// rather than a CUF sequence
			if travel := cur.Point.Sub(lastUpdate); lastUpdate != image.ZP && travel.Y == 0 && travel.X <= 4 {
				rn := 0
				var tmp [4]byte
				j, _ := g.index(lastUpdate)
				for _, r := range g.Rune[j:i] {
					if r == 0 {
						r = ' '
					}
					rn += utf8.EncodeRune(tmp[:], r)
				}
				if rn > 0 && rn <= 4 {
					for _, r := range g.Rune[j:i] {
						if r == 0 {
							r = ' '
						}
						m, _ := buf.WriteRune(r)
						n += m
						cur.ProcessRune(r)
					}
				}
			}

			n += buf.WriteSeq(cur.To(pt))
			n += buf.WriteSGR(cur.MergeSGR(ga))
			m, _ := buf.WriteRune(gr)
			n += m
			cur.ProcessRune(gr)
			lastUpdate = cur.Point
		}

	next:
		i++
		pt.X++
		for pt.X > g.Size.X {
			pt.X -= g.Size.X
			pt.Y++
		}
	}
	return n, cur
}
