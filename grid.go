package anansi

import (
	"image"

	"github.com/jcorbin/anansi/ansi"
)

// TODO grid composition/copy function

// Grid is a grid of screen cells.
type Grid struct {
	Attr []ansi.SGRAttr
	Rune []rune
	// TODO []string for multi-rune glyphs
	Stride int
	Rect   ansi.Rectangle
}

// Resize the grid to have room for n cells.
// Returns true only if the resize was a change, false if it was a no-op.
func (g *Grid) Resize(size image.Point) bool {
	if size == g.Rect.Size() {
		return false
	}
	if isSubgrid := g.Stride != g.Rect.Dx(); isSubgrid {
		if size.X > g.Stride {
			size.X = g.Stride
		}
		if g.Stride*size.Y > len(g.Rune) {
			size.Y = len(g.Rune) / g.Stride
		}
		g.Rect.Max = g.Rect.Min.Add(size)
	} else {
		if g.Rect.Min.Point == image.ZP {
			g.Rect.Min = ansi.Pt(1, 1)
		}
		g.Rect.Max = g.Rect.Min.Add(size)
		g.Stride = size.X
		n := g.Stride * size.Y
		for n > cap(g.Attr) {
			g.Attr = append(g.Attr, 0)
		}
		for n > cap(g.Rune) {
			g.Rune = append(g.Rune, 0)
		}
		g.Attr = g.Attr[:n]
		g.Rune = g.Rune[:n]
	}
	return true
}

// Bounds returns the screen bounding rectangle of the grid.
func (g Grid) Bounds() ansi.Rectangle {
	return g.Rect
}

// CellOffset returns the offset of the screen cell and true if it's
// within the Grid's Bounds().
func (g Grid) CellOffset(pt ansi.Point) (int, bool) {
	if !pt.In(g.Bounds()) {
		return 0, false
	}
	p := pt.ToImage() // convert to normal 0-indexed point
	return p.Y*g.Stride + p.X, true
}

// SubAt is a convenience for calling SubRect with at as the new Min point, and
// the receiver's Rect.Max point.
func (g Grid) SubAt(at ansi.Point) Grid {
	return g.SubRect(ansi.Rectangle{Min: at, Max: g.Rect.Max})
}

// SubSize is a convenience for calling SubRect with a Max point determined by
// adding the given size to the receiver's Rect.Min point.
func (g Grid) SubSize(sz image.Point) Grid {
	return g.SubRect(ansi.Rectangle{Min: g.Rect.Min, Max: g.Rect.Min.Add(sz)})
}

// SubRect returns a subgrid, sharing the receiver's Rune/Attr/Stride data, but
// with a new bounding Rect.
// Clamps r.Max to g.Rect.Max, and returns the zero Grid if r.Min is not in
// g.Rect.
func (g Grid) SubRect(r ansi.Rectangle) Grid {
	if !r.Min.In(g.Rect) {
		return Grid{}
	}
	if r.Max.X > g.Rect.Max.X {
		r.Max.X = g.Rect.Max.X
	}
	if r.Max.Y > g.Rect.Max.Y {
		r.Max.Y = g.Rect.Max.Y
	}
	return Grid{
		Attr:   g.Attr,
		Rune:   g.Rune,
		Stride: g.Stride,
		Rect:   r,
	}
}
