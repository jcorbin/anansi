package anansi

import "image"

// Bitmap is a 2-color bitmap targeting unicode braille runes.
type Bitmap struct {
	Bit    []bool
	Stride int
	Rect   image.Rectangle
}

// NewBitmap creates a new bitmap with the given bounding rectangle.
func NewBitmap(r image.Rectangle) *Bitmap {
	sz := r.Size()
	return &Bitmap{make([]bool, sz.X*sz.Y), sz.X, r}
}

// NewBitmapSize creates a new bitmap with the given size anchored at 0,0.
func NewBitmapSize(sz image.Point) *Bitmap {
	return &Bitmap{make([]bool, sz.X*sz.Y), sz.X, image.Rectangle{image.ZP, sz}}
}

// NewBitmapData creates a new bitmap with the given bit data and stride.
func NewBitmapData(stride int, data ...bool) *Bitmap {
	h := len(data) / stride
	for i, n := 0, len(data)%h; i < n; i++ {
		data = append(data, false)
	}
	sz := image.Pt(stride, h)
	return &Bitmap{data, stride, image.Rectangle{image.ZP, sz}}
}

// NewBitmapString creates a new bitmap from a set of representative strings.
// Within the strings, the `set` rune indicates a 1/true bit. Each string must
// be the same, stride, length.
func NewBitmapString(set rune, lines ...string) *Bitmap {
	var stride int
	var n int
	for _, line := range lines {
		if stride == 0 {
			stride = len(line)
		} else if len(line) != stride {
			panic("inconsistent line length")
		}
		n += stride
	}
	data := make([]bool, 0, n)
	for _, line := range lines {
		for _, r := range line {
			if r == set {
				data = append(data, true)
			} else {
				data = append(data, false)
			}
		}
	}
	return &Bitmap{data, stride, image.Rectangle{image.ZP, image.Pt(stride, len(lines))}}
}

// RuneSize returns the size of the bitmap in runes.
func (bi *Bitmap) RuneSize() (sz image.Point) {
	sz.X = (bi.Rect.Dx() + 1) / 2
	sz.Y = (bi.Rect.Dy() + 3) / 4
	return sz
}

// Get a single bitmap cell value.
func (bi *Bitmap) Get(p image.Point) bool {
	if !p.In(bi.Rect) {
		return false
	}
	return bi.Bit[p.Y*bi.Stride+p.X]
}

// Set a single bitmap cell value.
func (bi *Bitmap) Set(p image.Point, b bool) {
	if p.In(bi.Rect) {
		bi.Bit[p.Y*bi.Stride+p.X] = b
	}
}

// Rune builds a unicode braille rune representing a single 2x8 rectangle of
// bits, anchored at the give top-left point.
func (bi *Bitmap) Rune(p image.Point) (c rune) {
	// 0x2800
	// 0x0001 0x0008
	// 0x0002 0x0010
	// 0x0004 0x0020
	// 0x0040 0x0080
	if bi.Get(image.Pt(p.X, p.Y)) {
		c |= 1 << 0
	}
	if bi.Get(image.Pt(p.X, p.Y+1)) {
		c |= 1 << 1
	}
	if bi.Get(image.Pt(p.X, p.Y+2)) {
		c |= 1 << 2
	}
	if bi.Get(image.Pt(p.X+1, p.Y)) {
		c |= 1 << 3
	}
	if bi.Get(image.Pt(p.X+1, p.Y+1)) {
		c |= 1 << 4
	}
	if bi.Get(image.Pt(p.X+1, p.Y+2)) {
		c |= 1 << 5
	}
	if bi.Get(image.Pt(p.X, p.Y+3)) {
		c |= 1 << 6
	}
	if bi.Get(image.Pt(p.X+1, p.Y+3)) {
		c |= 1 << 7
	}
	return 0x2800 | c
}
