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

// Cell is a reference to a positioned cell within a Grid.
type Cell struct {
	image.Point
	Grid *Grid
	I    int
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

// Bounds returns the bounding rectangle of the grid in cell space: 1,1 origin,
// with max of Size+1.
func (g *Grid) Bounds() image.Rectangle {
	return image.Rectangle{image.Pt(1, 1), g.Size.Add(image.Pt(1, 1))}
}

// Cell returns the grid cell for the given point, which will be the Cell zero
// value if outside the grid.
func (g *Grid) Cell(pt image.Point) Cell {
	if i, ok := g.index(pt); ok {
		return Cell{pt, g, i}
	}
	return Cell{}
}

// Get returns the cell's rune value and SGR attributes.
func (c Cell) Get() (rune, ansi.SGRAttr) {
	if c.Grid != nil {
		return c.Grid.Rune[c.I], c.Grid.Attr[c.I]
	}
	return 0, 0
}

// Set sets both the cell's rune value and its SGR attributes.
func (c Cell) Set(r rune, a ansi.SGRAttr) {
	if c.Grid != nil {
		c.Grid.Rune[c.I] = r
		c.Grid.Attr[c.I] = a
	}
}

// Rune returns the cell's rune value.
func (c Cell) Rune() rune {
	if c.Grid != nil {
		return c.Grid.Rune[c.I]
	}
	return 0
}

// Attr returns the cell's SGR attributes.
func (c Cell) Attr() ansi.SGRAttr {
	if c.Grid != nil {
		return c.Grid.Attr[c.I]
	}
	return 0
}

// SetRune sets the cell's rune value.
func (c Cell) SetRune(r rune) {
	if c.Grid != nil {
		c.Grid.Rune[c.I] = r
	}
}

// SetAttr sets the cell's SGR attributes.
func (c Cell) SetAttr(a ansi.SGRAttr) {
	if c.Grid != nil {
		c.Grid.Attr[c.I] = a
	}
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
	if len(prior.Attr) == 0 || len(prior.Rune) == 0 || prior.Size == image.ZP || prior.Size != g.Size {
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
				if gr == prior.Rune[j] && ga == prior.Attr[j] { // NOTE range ok since pt <= prior.Size
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

			mv := cur.To(pt)
			ad := cur.MergeSGR(ga)
			n += buf.WriteSeq(mv)
			n += buf.WriteSGR(ad)
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
