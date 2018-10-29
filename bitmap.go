package anansi

import (
	"errors"
	"image"
	"unicode/utf8"
)

// Bitmap is a 2-color bitmap targeting unicode braille runes.
type Bitmap struct {
	Bit    []bool
	Stride int
	Rect   image.Rectangle
}

// NewBitmap creates a new bitmap with the given bit data and stride.
func NewBitmap(stride int, data []bool) *Bitmap {
	h := len(data) / stride
	for i, n := 0, len(data)%h; i < n; i++ {
		data = append(data, false)
	}
	sz := image.Pt(stride, h)
	return &Bitmap{data, stride, image.Rectangle{image.ZP, sz}}
}

// NewBitmapSize creates a new bitmap with the given size anchored at 0,0.
func NewBitmapSize(sz image.Point) *Bitmap {
	return &Bitmap{make([]bool, sz.X*sz.Y), sz.X, image.Rectangle{image.ZP, sz}}
}

// ParseBitmap parses a convenience representation for creating bitmaps.
// The set string argument indicates how a 1 (or true) bit will be recognized;
// it may be any 1 or 2 rune string. Any other single or double runes in the
// strings will be mapped to zero (allowing the caller to put anything there
// for other, self-documenting, purposes).
func ParseBitmap(set string, lines ...string) (stride int, data []bool, err error) {
	var setRunes [2]rune
	var pat []rune
	switch utf8.RuneCountInString(set) {
	case 1:
		setRunes[0], _ = utf8.DecodeRuneInString(set)
		pat = setRunes[:1]
	case 2:
		var n int
		setRunes[0], n = utf8.DecodeRuneInString(set)
		setRunes[1], _ = utf8.DecodeRuneInString(set[n:])
		pat = setRunes[:2]
	default:
		return 0, nil, errors.New("must use a 1 or 2-rune string")
	}

	var n int
	for _, line := range lines {
		m := utf8.RuneCountInString(line)
		if len(pat) == 2 {
			if m%2 == 1 {
				return 0, nil, errors.New("odd-length line in double rune bitmap parse")
			}
			m /= 2
		}
		if stride == 0 {
			stride = m
		} else if m != stride {
			return 0, nil, errors.New("inconsistent line length")
		}
		n += stride
	}
	data = make([]bool, 0, n)

	for _, line := range lines {
		for len(line) > 0 {
			all := true
			for _, patRune := range pat {
				r, n := utf8.DecodeRuneInString(line)
				line = line[n:]
				all = all && r == patRune
			}
			if all {
				data = append(data, true)
			} else {
				data = append(data, false)
			}
		}
	}
	return stride, data, nil
}

// MustParseBitmap is an infaliable version of ParseBitmapString: it
// panics if any non-nil error is returned by it.
func MustParseBitmap(set string, lines ...string) (stride int, data []bool) {
	stride, data, err := ParseBitmap(set, lines...)
	if err != nil {
		panic(err)
	}
	return stride, data
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

// Rune builds a unicode braille rune representing a single 2x4 rectangle of
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
