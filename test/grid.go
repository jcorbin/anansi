package anansitest

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// ParseGridLines parses grid data from a list of line strings.
// Panics if the lines contain any non-SGR ansi escape sequences.
// Panics if every line after the first isn't the same width as the first.
func ParseGridLines(lines []string) (g anansi.Grid) {
	var (
		rs []rune
		as []ansi.SGRAttr
		at ansi.SGRAttr
	)
	g.Stride = -1
	for _, line := range lines {
		rs, as, at = parseGridLine(line, at)
		if g.Stride < 0 {
			g.Stride = len(rs)
		} else if len(rs) != g.Stride {
			panic("invalid grid line length shape")
		}
		g.Rune = append(g.Rune, rs...)
		g.Attr = append(g.Attr, as...)
	}
	g.Rect = ansi.Rect(1, 1, 1+g.Stride, 1+len(g.Rune)/g.Stride)
	return g
}

func parseGridLine(line string, at ansi.SGRAttr) (rs []rune, as []ansi.SGRAttr, _ ansi.SGRAttr) {
	b := []byte(line)
	for len(b) > 0 {
		e, a, n := ansi.DecodeEscape(b)
		b = b[n:]
		switch e {
		case ansi.SGR:
			attr, _, err := ansi.DecodeSGR(a)
			if err != nil {
				panic(fmt.Sprintf("failed to decode SGR: %v", err))
			}
			at = at.Merge(attr)
		case 0:
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			switch {
			case r == 0:
			case unicode.IsControl(r):
				panic(fmt.Sprintf("unexpected control rune %q", r))
			}
			rs = append(rs, r)
			as = append(as, at)
		default:
			panic(fmt.Sprintf("unexpected %v escape", e))
		}
	}
	return rs, as, at
}

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
	p := r.Min
	i, _ := g.CellOffset(p)
	for ; p.Y < r.Max.Y; p.Y++ {
		rs = append(rs, g.Rune[i:i+g.Stride])
		as = append(as, g.Attr[i:i+g.Stride])
		i += g.Stride
	}
	return rs, as
}
