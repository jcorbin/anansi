package anansi_test

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	anansitest "github.com/jcorbin/anansi/test"
)

func TestDrawGrid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		dst    []string
		src    []string
		dstSub ansi.Rectangle
		srcSub ansi.Rectangle
		out    []string
		styles []Style
	}{

		{
			name: "rune transparency",
			dst: []string{
				"AAA",
				"AAA",
				"AAA",
			},
			src: []string{
				"\x00B\x00",
				"BBB",
				"\x00B\x00",
			},
			out: []string{
				"ABA",
				"BBB",
				"ABA",
			},
			styles: []Style{TransparentRunes},
		},

		{
			name: "rune overwrite",
			dst: []string{
				"AAA",
				"AAA",
				"AAA",
			},
			src: []string{
				"\x00B\x00",
				"BBB",
				"\x00B\x00",
			},
			out: []string{
				".B.",
				"BBB",
				".B.",
			},
		},

		{
			name: "attr transparency",
			dst: []string{
				"\x1b[32;43mAAA",
				"AAA",
				"AAA",
			},
			src: []string{
				"B\x1b[31;44mB\x1b[0mB",
				"\x1b[31;44mBBB\x1b[0m",
				"B\x1b[31;44mB\x1b[0mB",
			},
			styles: []Style{TransparentAttrBGFG},
			out: []string{
				"\x1b[32;43mB\x1b[31;44mB\x1b[32;43mB",
				"\x1b[31;44mBBB",
				"\x1b[32;43mB\x1b[31;44mB\x1b[32;43mB",
			},
		},

		{
			name: "attr fg transparency",
			dst: []string{
				"\x1b[32;43mAAA",
				"AAA",
				"AAA",
			},
			src: []string{
				"B\x1b[31;44mB\x1b[0mB",
				"\x1b[31;44mBBB\x1b[0m",
				"B\x1b[31;44mB\x1b[0mB",
			},
			styles: []Style{TransparentAttrFG},
			out: []string{
				"\x1b[32mB\x1b[31;44mB\x1b[0;32mB",
				"\x1b[31;44mBBB",
				"\x1b[0;32mB\x1b[31;44mB\x1b[0;32mB",
			},
		},

		{
			name: "attr bg transparency",
			dst: []string{
				"\x1b[32;43mAAA",
				"AAA",
				"AAA",
			},
			src: []string{
				"B\x1b[31;44mB\x1b[0mB",
				"\x1b[31;44mBBB\x1b[0m",
				"B\x1b[31;44mB\x1b[0mB",
			},
			styles: []Style{TransparentAttrBG},
			out: []string{
				"\x1b[43mB\x1b[31;44mB\x1b[0;43mB",
				"\x1b[31;44mBBB",
				"\x1b[0;43mB\x1b[31;44mB\x1b[0;43mB",
			},
		},

		{
			name: "attr overwrite",
			dst: []string{
				"\x1b[32;43mAAA",
				"AAA",
				"AAA",
			},
			src: []string{
				"B\x1b[31;44mB\x1b[0mB",
				"\x1b[31;44mBBB\x1b[0m",
				"B\x1b[31;44mB\x1b[0mB",
			},
			out: []string{
				"B\x1b[31;44mB\x1b[0mB",
				"\x1b[31;44mBBB",
				"\x1b[0mB\x1b[31;44mB\x1b[0mB",
			},
		},

		// TODO subgrid cases

	} {
		t.Run(tc.name, func(t *testing.T) {
			src := anansitest.ParseGridLines(tc.src)
			dst := anansitest.ParseGridLines(tc.dst)
			if tc.srcSub != ansi.ZR {
				src = src.SubRect(tc.srcSub)
			}
			if tc.dstSub != ansi.ZR {
				dst = dst.SubRect(tc.dstSub)
			}
			DrawGrid(dst, src, tc.styles...)
			out := anansitest.GridLines(dst, '.')
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestDrawBitmap(t *testing.T) {
	for _, tc := range []struct {
		name     string
		gridSize image.Point
		inLines  []string
		outLines []string
		at       ansi.Point
		styles   []Style
	}{
		{
			name:     "basic test pattern",
			gridSize: image.Pt(3, 3),
			at:       ansi.Pt(1, 1), // cell space origin
			inLines: []string{
				"#.#.",
				".#.#",
				"#.#.",
				".#.#",
				"#.#.",
				".#.#",
				"#.#.",
				".#.#",
			},
			outLines: []string{
				"⢕⢕_",
				"⢕⢕_",
				"___",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var g Grid
			g.Resize(tc.gridSize)
			bi := NewBitmapString('#', tc.inLines...)
			DrawBitmap(g.SubAt(tc.at), bi, tc.styles...)
			assert.Equal(t, tc.outLines, anansitest.GridLines(g, '_'))
		})
	}
}
