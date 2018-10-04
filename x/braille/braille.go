package braille

import (
	"image"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// NewBitmap creates a new braille bitmap with the given bounding
// rectangle.
func NewBitmap(r image.Rectangle) *Bitmap {
	sz := r.Size()
	return &Bitmap{make([]bool, sz.X*sz.Y), sz.X, r}
}

// NewBitmapSize creates a new braille bitmap with the given size
// anchored at 0,0.
func NewBitmapSize(sz image.Point) *Bitmap {
	return &Bitmap{make([]bool, sz.X*sz.Y), sz.X, image.Rectangle{image.ZP, sz}}
}

// NewBitmapData creates a new braille bitmap with the given bit data and stride.
func NewBitmapData(stride int, data ...bool) *Bitmap {
	h := len(data) / stride
	for i, n := 0, len(data)%h; i < n; i++ {
		data = append(data, false)
	}
	sz := image.Pt(stride, h)
	return &Bitmap{data, stride, image.Rectangle{image.ZP, sz}}
}

// NewBitmapString creates a new braille bitmap from a set of representative
// strings. Within the strings, the `set` rune indicates a 1/true bit. Each
// string must be the same, stride, length.
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

// Bitmap is a 2-color bitmap targeting unicode braille runes.
type Bitmap struct {
	Bit    []bool
	Stride int
	Rect   image.Rectangle
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

// GetRune builds a unicode braille rune representing a single 2x8 rectangle of
// bits, anchored at the give top-left point.
func (bi *Bitmap) GetRune(p image.Point) (c rune) {
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

// CopyInto copies the bitmap's rune representation into an anansi cell grid.
// Passes the Style any prior grid attribute.
// Only sets a cell value if the style returns a non-zero rune.
//
// NOTE The at argument is in grid cell space (1,1 origin-relative).
// TODO eliminate the at argument once Grid gets refactor to be image-like.
func (bi *Bitmap) CopyInto(g anansi.Grid, at image.Point, styles ...Style) {
	style := Styles(styles...)
	for gp, p := at, bi.Rect.Min; p.Y < bi.Rect.Max.Y; p.Y += 4 {
		gp.X = at.X
		for p.X = bi.Rect.Min.X; p.X < bi.Rect.Max.X; p.X += 2 {
			cell := g.Cell(gp)
			if r, a := style.Style(p, bi.GetRune(p), cell.Attr()); r != 0 {
				cell.Set(r, a)
			}
			gp.X++
		}
		gp.Y++
	}
}

// RenderInto renders the bitmap into an ansi buffer, optionally using raw
// cursor position codes, rather than newlines.
func (bi *Bitmap) RenderInto(buf *ansi.Buffer, rawMode bool, styles ...Style) {
	style := Styles(styles...)
	for p := bi.Rect.Min; p.Y < bi.Rect.Max.Y; p.Y += 4 {
		if p.Y > 0 {
			if rawMode {
				buf.WriteESC(ansi.CUD)
				buf.WriteSeq(ansi.CUB.WithInts(p.X))
			} else {
				buf.WriteByte('\n')
			}
		}
		for p.X = bi.Rect.Min.X; p.X < bi.Rect.Max.X; p.X += 2 {
			if r, a := style.Style(p, bi.GetRune(p), 0); r != 0 {
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

// Style allows styling of rendered braille runes. It's eponymous method gets
// called with for each x,y point (in Bitmap space), rendered rune, and ansi
// attr to be used; whatever rune and ansi attribute it returns is rendered.
type Style interface {
	Style(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr)
}

// Styles combines zero or more styles into a non-nil Style; if given none, it
// returns a no-op Style; if given many, it returns a Style that calls each in
// turn.
func Styles(ss ...Style) Style {
	var res styles
	for _, s := range ss {
		switch impl := s.(type) {
		case _noopStyle:
			continue
		case styles:
			res = append(res, impl...)
		default:
			res = append(res, s)
		}
	}
	switch len(res) {
	case 0:
		return NoopStyle
	case 1:
		return res[0]
	default:
		return res
	}
}

// StyleFunc is a convenience type alias for implementing Style.
type StyleFunc func(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr)

// Style calls the aliased function pointer
func (f StyleFunc) Style(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	return f(p, r, a)
}

type _noopStyle struct{}

func (ns _noopStyle) Style(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr) { return r, 0 }

// NoopStyle is a no-op style, used as a zero fill by Styles.
var NoopStyle Style = _noopStyle{}

type styles []Style

func (ss styles) Style(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	for _, s := range ss {
		r, a = s.Style(p, r, a)
	}
	return r, a
}

// FillStyle implements a Style that fills empty runes with a fixed rune value.
type FillStyle rune

// Style replaces the passed rune with the receiver if the passed rune is 0 or
// empty braille character.
func (fs FillStyle) Style(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r == 0 || r == 0x2800 {
		r = rune(fs)
	}
	return r, a
}

// AttrStyle implements a Style that returns a fixed ansi attr for any non-zero runes.
type AttrStyle ansi.SGRAttr

// Style replaces the passed attr with the receiver if the passed rune is non-0.
func (as AttrStyle) Style(p image.Point, r rune, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r != 0 {
		a = ansi.SGRAttr(as)
	}
	return r, a
}
