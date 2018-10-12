package anansitest

import (
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// GridLines returns a slice of line strings built from the grid's cell data.
func GridLines(g anansi.Grid, fill rune) (lines []string) {
	var ca ansi.SGRAttr
	r := g.Bounds()
	p := r.Min
	for ; p.Y < r.Max.Y; p.Y++ {
		var b []byte
		p.X = r.Min.X
		for i, _ := g.CellOffset(p); p.X < r.Max.X; p.X++ {
			r, a := g.Rune[i], g.Attr[i]
			if a != ca {
				a = ca.Diff(a)
				b = a.AppendTo(b)
				ca = ca.Merge(a)
			}
			var tmp [4]byte
			if r == 0 {
				r = fill
			}
			b = append(b, tmp[:utf8.EncodeRune(tmp[:], r)]...)
			i++
		}
		lines = append(lines, string(b))
	}
	return lines
}

// GridRowData the grid''s cell data in two slices-of-slices.
func GridRowData(g anansi.Grid) (rs [][]rune, as [][]ansi.SGRAttr) {
	r := g.Bounds()
	stride := r.Dx()
	p := r.Min
	i, _ := g.CellOffset(p)
	for ; p.Y < r.Max.Y; p.Y++ {
		rs = append(rs, g.Rune[i:i+stride])
		as = append(as, g.Attr[i:i+stride])
		i += stride
	}
	return rs, as
}
