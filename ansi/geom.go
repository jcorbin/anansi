package ansi

import "image"

// Point represents an ANSI screen point, relative to a 1,1 column,row origin.
//
// This naturally aligns with the requirements of parsing and building ANSI
// control sequences (e.g. for cursor positioning and mouse events), while
// allowing the 1,1-origin semantic to be type checked.
type Point struct{ image.Point }

// Rectangle represents an ANSI screen rectangle, defined by a pair of ANSI
// screen points.
type Rectangle struct{ Min, Max Point }

// ZR is the zero rectangle value; it is not a valid rectangle, but useful only
// for signalling an undefined rectangle.
var ZR Rectangle

// ZP is the zero point value; it is not a valid point, but useful only for
// signalling an undefined point.
var ZP Point

// Rect constructs an ANSI screen rectangle; panics if either of the pairs of
// x,y components are invalid.
func Rect(x0, y0, x1, y1 int) (r Rectangle) {
	if x0 < 1 || y0 < 1 || x1 < 1 || y1 < 1 {
		panic("invalid ansi.Rectangle value")
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	r.Min.X, r.Min.Y = x0, y0
	r.Max.X, r.Max.Y = x1, y1
	return r
}

// RectFromImage creates an ANSI screen rectangle from an image rectangle,
// converting from 0,0 origin to 1,1 origin.
// Panics if either of the return value's points would not be Point.Valid().
func RectFromImage(ir image.Rectangle) (r Rectangle) {
	if ir.Min.X < 0 || ir.Min.Y < 0 || ir.Max.X < 0 || ir.Max.Y < 0 {
		panic("out of bounds image.Rectangle value")
	}
	return r
}

// Pt constructs an ANSI screen point; panics if either of the x or y
// components is not a counting number (> 0).
func Pt(x, y int) Point {
	if x < 1 || y < 1 {
		panic("invalid ansi.Point value")
	}
	return Point{image.Pt(x, y)}
}

// PtFromImage creates an ANSI screen point from an image point, converting from
// 0,0 origin to 1,1 origin.
// Panics if the return value would have been not Point.Valid().
func PtFromImage(p image.Point) Point {
	if p.X < 0 || p.Y < 0 {
		panic("out of bounds image.Point value")
	}
	p.X++
	p.Y++
	return Point{p}
}

// Valid returns true only if both X and Y components are >= 0.
func (p Point) Valid() bool {
	return p.X >= 1 && p.Y >= 1
}

// ToImage converts to a normal 0,0 origin image point.
// Panics if the point is not Valid().
func (p Point) ToImage() image.Point {
	if p.X < 1 || p.Y < 1 {
		panic("invalid ansi.Point value")
	}
	return image.Pt(p.X-1, p.Y-1)
}

// ToImage converts to a normal 0,0 origin image rectangle.
// Panics if either Min or Max point are not Valid().
func (r Rectangle) ToImage() image.Rectangle {
	if r.Min.X < 1 || r.Min.Y < 1 || r.Max.X < 1 || r.Max.Y < 1 {
		panic("invalid ansi.Rectangle value")
	}
	return image.Rect(r.Min.X-1, r.Min.Y-1, r.Max.X-1, r.Max.Y-1)
}

// Add the given relative image point to a copy of the receiver screen point,
// returning the copy.
func (p Point) Add(q image.Point) Point {
	p.X += q.X
	p.Y += q.Y
	return p
}

// Sub tract the given relative image point to a copy of the receiver screen
// point, returning the copy.
func (p Point) Sub(q image.Point) Point {
	p.X -= q.X
	p.Y -= q.Y
	return p
}

// Diff computes the relative difference (an image point) between two screen points.
func (p Point) Diff(q Point) image.Point {
	return image.Pt(p.X-q.X, p.Y-q.Y)
}

// Mul returns the vector p*k.
func (p Point) Mul(k int) Point {
	p.X *= k
	p.Y *= k
	return p
}

// Div returns the vector p/k.
func (p Point) Div(k int) Point {
	p.X /= k
	p.Y /= k
	return p
}

// In reports whether p is in r.
func (p Point) In(r Rectangle) bool {
	return r.Min.X <= p.X && p.X < r.Max.X &&
		r.Min.Y <= p.Y && p.Y < r.Max.Y
}

// Eq reports whether p and q are equal.
func (p Point) Eq(q Point) bool {
	return p == q
}

// String returns a string representation of r like "(3,4)-(6,5)".
func (r Rectangle) String() string {
	return r.Min.String() + "-" + r.Max.String()
}

// Dx returns r's width.
func (r Rectangle) Dx() int {
	return r.Max.X - r.Min.X
}

// Dy returns r's height.
func (r Rectangle) Dy() int {
	return r.Max.Y - r.Min.Y
}

// Size returns r's width and height.
func (r Rectangle) Size() image.Point {
	return image.Pt(
		r.Max.X-r.Min.X,
		r.Max.Y-r.Min.Y,
	)
}

// Add returns the rectangle r translated by p.
func (r Rectangle) Add(p image.Point) Rectangle {
	r.Min.X += p.X
	r.Min.Y += p.Y
	r.Max.X += p.X
	r.Max.Y += p.Y
	return r
}

// Sub returns the rectangle r translated by -p.
func (r Rectangle) Sub(p image.Point) Rectangle {
	r.Min.X -= p.X
	r.Min.Y -= p.Y
	r.Max.X -= p.X
	r.Max.Y -= p.Y
	return r
}

// Inset returns the rectangle r inset by n, which may be negative. If either
// of r's dimensions is less than 2*n then an empty rectangle near the center
// of r will be returned.
func (r Rectangle) Inset(n int) Rectangle {
	if r.Dx() < 2*n {
		r.Min.X = (r.Min.X + r.Max.X) / 2
		r.Max.X = r.Min.X
	} else {
		r.Min.X += n
		r.Max.X -= n
	}
	if r.Dy() < 2*n {
		r.Min.Y = (r.Min.Y + r.Max.Y) / 2
		r.Max.Y = r.Min.Y
	} else {
		r.Min.Y += n
		r.Max.Y -= n
	}
	return r
}

// Empty reports whether the rectangle contains no points.
func (r Rectangle) Empty() bool {
	return r.Min.X >= r.Max.X || r.Min.Y >= r.Max.Y
}

// Intersect returns the largest rectangle contained by both r and s. If the
// two rectangles do not overlap then an empty rectangle at r.Min.
func (r Rectangle) Intersect(s Rectangle) Rectangle {
	if r.Min.X < s.Min.X {
		r.Min.X = s.Min.X
	}
	if r.Min.Y < s.Min.Y {
		r.Min.Y = s.Min.Y
	}
	if r.Max.X > s.Max.X {
		r.Max.X = s.Max.X
	}
	if r.Max.Y > s.Max.Y {
		r.Max.Y = s.Max.Y
	}
	// Letting r0 and s0 be the values of r and s at the time that the method
	// is called, this next line is equivalent to:
	//
	// if max(r0.Min.X, s0.Min.X) >= min(r0.Max.X, s0.Max.X) || likewiseForY { etc }
	if r.Empty() {
		r.Max = r.Min
	}
	return r
}

// Union returns the smallest rectangle that contains both r and s.
func (r Rectangle) Union(s Rectangle) Rectangle {
	if r.Empty() {
		return s
	}
	if s.Empty() {
		return r
	}
	if r.Min.X > s.Min.X {
		r.Min.X = s.Min.X
	}
	if r.Min.Y > s.Min.Y {
		r.Min.Y = s.Min.Y
	}
	if r.Max.X < s.Max.X {
		r.Max.X = s.Max.X
	}
	if r.Max.Y < s.Max.Y {
		r.Max.Y = s.Max.Y
	}
	return r
}

// Eq reports whether r and s contain the same set of points. All empty
// rectangles are considered equal.
func (r Rectangle) Eq(s Rectangle) bool {
	return r == s || r.Empty() && s.Empty()
}

// Overlaps reports whether r and s have a non-empty intersection.
func (r Rectangle) Overlaps(s Rectangle) bool {
	return !r.Empty() && !s.Empty() &&
		r.Min.X < s.Max.X && s.Min.X < r.Max.X &&
		r.Min.Y < s.Max.Y && s.Min.Y < r.Max.Y
}

// In reports whether every point in r is in s.
func (r Rectangle) In(s Rectangle) bool {
	if r.Empty() {
		return true
	}
	// Note that r.Max is an exclusive bound for r, so that r.In(s)
	// does not require that r.Max.In(s).
	return s.Min.X <= r.Min.X && r.Max.X <= s.Max.X &&
		s.Min.Y <= r.Min.Y && r.Max.Y <= s.Max.Y
}

// Canon returns the canonical version of r. The returned rectangle has minimum
// and maximum coordinates swapped if necessary so that it is well-formed.
func (r Rectangle) Canon() Rectangle {
	if r.Max.X < r.Min.X {
		r.Min.X, r.Max.X = r.Max.X, r.Min.X
	}
	if r.Max.Y < r.Min.Y {
		r.Min.Y, r.Max.Y = r.Max.Y, r.Min.Y
	}
	return r
}
