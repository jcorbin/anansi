package anansitest

import (
	"image"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// GridLines returns a slice of line strings built from the grid's cell data.
func GridLines(g anansi.Grid, fill rune) (lines []string) {
	var ca ansi.SGRAttr
	for i, p := 0, image.ZP; p.Y < g.Size.Y; p.Y++ {
		var b []byte
		for p.X = 0; p.X < g.Size.X; p.X++ {
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
	// NOTE p is in array space, not 1,1-based screen space
	for i, p := 0, image.ZP; p.Y < g.Size.Y; p.Y++ {
		j := p.Y * g.Size.X
		rs = append(rs, g.Rune[i:j])
		as = append(as, g.Attr[i:j])
		i = j
	}
	return rs, as
}
