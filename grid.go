package anansi

import (
	"image"

	"github.com/jcorbin/anansi/ansi"
)

// Grid is a grid of screen cells.
type Grid struct {
	Rect   ansi.Rectangle
	Stride int
	Attr   []ansi.SGRAttr
	Rune   []rune
	// TODO []string for multi-rune glyphs
}

// Resize the grid to have room for n cells.
// Returns true only if the resize was a change, false if it was a no-op.
func (g *Grid) Resize(size image.Point) bool {
	if size == g.Rect.Size() {
		return false
	}
	if g.IsSub() {
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
		if n > cap(g.Rune) {
			as := make([]ansi.SGRAttr, n)
			rs := make([]rune, n)
			copy(as, g.Attr)
			copy(rs, g.Rune)
			g.Attr, g.Rune = as, rs
		} else {
			g.Attr = g.Attr[:n]
			g.Rune = g.Rune[:n]
		}
		// TODO re-stride data
	}
	return true
}

// Clear the (maybe sub) grid; zeros all runes an attributes.
func (g Grid) Clear() {
	if !g.IsSub() {
		for i := range g.Rune {
			g.Rune[i] = 0
			g.Attr[i] = 0
		}
		return
	}

	pt := g.Rect.Min
	i, _ := g.CellOffset(pt)
	dx := g.Rect.Dx()
	for ; pt.Y < g.Rect.Max.Y; pt.Y++ {
		for pt.X = g.Rect.Min.X; pt.X < g.Rect.Max.X; pt.X++ {
			g.Rune[i] = 0
			g.Attr[i] = 0
			i++
		}
		i -= dx       // CR
		i += g.Stride // LF
	}
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

// IsSub returns true if the grid's bounding rectangle only covers a
// sub-section of its underlying data.
func (g *Grid) IsSub() bool {
	return g.Rect.Size() != g.fullSize()
}

func (g *Grid) fullSize() image.Point {
	if g.Stride == 0 {
		return image.ZP
	}
	return image.Pt(g.Stride, len(g.Rune)/g.Stride)
}

// Full returns the full grid containing the receiver grid, reversing any
// sub-grid targeting done by SubRect().
func (g Grid) Full() Grid {
	g.Rect.Min = ansi.Pt(1, 1)
	g.Rect.Max = g.Rect.Min.Add(g.fullSize())
	return g
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

// SubRect returns a sub-grid, sharing the receiver's Rune/Attr/Stride data,
// but with a new bounding Rect. Clamps r.Max to g.Rect.Max, and returns the
// zero Grid if r.Min is not in g.Rect.
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

// Eq returns true only if the other grid has the same size and contents as the
// receiver.
func (g Grid) Eq(other Grid, zero rune) bool {
	n := len(g.Rune)
	if n != len(other.Rune) {
		return false
	}
	i := 0
	for ; i < n; i++ {
		if g.Attr[i] != other.Attr[i] {
			return false
		}
		gr, or := g.Rune[i], other.Rune[i]
		if gr == 0 {
			gr = zero
		}
		if or == 0 {
			or = zero
		}
		if gr != or {
			return false
		}
	}
	return true
}

// SetCell stores the given rune and attribute data into the grid at the given
// point and returns true, only if in range; returns false otherwise.
func (g Grid) SetCell(p ansi.Point, r rune, attr ansi.SGRAttr) (ok bool) {
	var i int
	i, ok = g.CellOffset(p)
	if ok {
		g.Rune[i] = r
		g.Attr[i] = attr
	}
	return ok
}

// GetCell loads rune and attributee data from the given point within the
// grid, returning them with ok=true if in range; returns 0 values and
// ok=false otherwise.
func (g Grid) GetCell(p ansi.Point) (r rune, attr ansi.SGRAttr, ok bool) {
	var i int
	i, ok = g.CellOffset(p)
	if ok {
		r = g.Rune[i]
		attr = g.Attr[i]
	}
	return r, attr, ok
}
