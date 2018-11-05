package anansi_test

import (
	"fmt"
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
		bi       Bitmap
		outLines []string
		at       ansi.Point
		styles   []Style
	}{

		{
			name:     "2x2 alternating into 2x2 grid",
			gridSize: image.Pt(2, 2),
			bi:       fillTestBitmap(image.Pt(2*2, 2*4), alternating),
			outLines: []string{
				"⢕⢕",
				"⢕⢕",
			},
		},

		{
			name:     "2x2 alternating into 3x3 grid",
			gridSize: image.Pt(3, 3),
			bi:       fillTestBitmap(image.Pt(2*2, 2*4), alternating),
			outLines: []string{
				"⢕⢕_",
				"⢕⢕_",
				"___",
			},
		},

		{
			name:     "2x2 alternating into 3x3 grid @2,2",
			gridSize: image.Pt(3, 3),
			at:       ansi.Pt(2, 2),
			bi:       fillTestBitmap(image.Pt(2*2, 2*4), alternating),
			outLines: []string{
				"___",
				"_⢕⢕",
				"_⢕⢕",
			},
		},

		{
			name:     "2x2 alternating into 3x3 grid @3,3",
			gridSize: image.Pt(3, 3),
			at:       ansi.Pt(3, 3),
			bi:       fillTestBitmap(image.Pt(2*2, 2*4), alternating),
			outLines: []string{
				"___",
				"___",
				"__⢕",
			},
		},

		{
			name:     "8x4 alternating into 16x8 grid",
			gridSize: image.Pt(16, 8),
			bi:       fillTestBitmap(image.Pt(8*2, 4*4), alternating),
			outLines: []string{
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"________________",
				"________________",
				"________________",
				"________________",
			},
		},

		{
			name:     "8x4 alternating into 16x8 grid @9,1",
			gridSize: image.Pt(16, 8),
			at:       ansi.Pt(9, 1),
			bi:       fillTestBitmap(image.Pt(8*2, 4*4), alternating),
			outLines: []string{
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________________",
				"________________",
				"________________",
				"________________",
			},
		},

		{
			name:     "8x4 alternating into 16x8 grid @9,5",
			gridSize: image.Pt(16, 8),
			at:       ansi.Pt(9, 5),
			bi:       fillTestBitmap(image.Pt(8*2, 4*4), alternating),
			outLines: []string{
				"________________",
				"________________",
				"________________",
				"________________",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
				"________⢕⢕⢕⢕⢕⢕⢕⢕",
			},
		},

		{
			name:     "8x4 alternating into 16x8 grid @1,5",
			gridSize: image.Pt(16, 8),
			at:       ansi.Pt(1, 5),
			bi:       fillTestBitmap(image.Pt(8*2, 4*4), alternating),
			outLines: []string{
				"________________",
				"________________",
				"________________",
				"________________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
				"⢕⢕⢕⢕⢕⢕⢕⢕________",
			},
		},

		{
			name:     "16x8 alternating into 16x8 grid",
			gridSize: image.Pt(16, 8),
			bi:       fillTestBitmap(image.Pt(16*2, 8*4), alternating),
			outLines: []string{
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
			},
		},

		{
			name:     "32x16 alternating into 16x8 grid",
			gridSize: image.Pt(16, 8),
			bi:       fillTestBitmap(image.Pt(32*2, 16*4), alternating),
			outLines: []string{
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
				"⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕",
			},
		},

		// DrawBitmap
		// 	dst.Rect:(1,1)-(119,56) dst.Stride:118
		// 	src.Rect:(0,0)-(236,219) src.Stride:236

		// DrawBitmap
		// 	dst.Rect:(1,1)-(119,56)
		// 	dst.Stride:118
		// 	src.Rect:(0,0)-(236,219)
		// 	src.Stride:236

		// lolwut
		// 	screenSize:(118,55)
		// 	canvasSize:(236,219)
		// 	squareSide:43
		// 	squaresPerCol:5
		// 	padding:2

	} {
		t.Run(tc.name, func(t *testing.T) {
			var g Grid
			g.Resize(tc.gridSize)
			dg := g
			if tc.at != ansi.ZP {
				dg = dg.SubAt(tc.at)
			}
			DrawBitmap(dg, tc.bi, tc.styles...)
			assert.Equal(t, tc.outLines, anansitest.GridLines(g, '_'))
		})
	}
}

func BenchmarkDrawBitmap(b *testing.B) {
	for _, bc := range []struct {
		name     string
		minSize  image.Point
		maxSize  image.Point
		sizeStep image.Point
		pattern  func(image.Point) bool
		styles   []Style
	}{
		{
			name:     "alternating",
			minSize:  image.Pt(2, 2),
			maxSize:  image.Pt(100, 50),
			sizeStep: image.Pt(2, 1),
			pattern:  alternating,
		},
	} {
		b.Run(bc.name, func(b *testing.B) {
			for sz := bc.minSize; sz.X <= bc.maxSize.X && sz.Y <= bc.maxSize.Y; sz = sz.Add(bc.sizeStep) {
				b.Run(fmt.Sprintf("size:%v", sz), func(b *testing.B) {
					var g Grid
					g.Resize(sz)
					bi := fillTestBitmap(image.Pt(2*sz.X, 4*sz.Y), bc.pattern)
					for i := 0; i < b.N; i++ {
						resetTestGrid(g)
						DrawBitmap(g, bi, bc.styles...)
					}
				})
			}
		})
	}
}

func alternating(pt image.Point) bool {
	return pt.X%2 == pt.Y%2
}

func fillTestBitmap(sz image.Point, f func(image.Point) bool) Bitmap {
	var bi Bitmap
	bi.Resize(sz)
	var pt image.Point
	i := 0
	for pt.Y = bi.Rect.Min.Y; pt.Y < bi.Rect.Max.Y; pt.Y++ {
		for pt.X = bi.Rect.Min.X; pt.X < bi.Rect.Max.X; pt.X++ {
			bi.Bit[i] = f(pt)
			i++
		}
	}
	return bi
}

func resetTestGrid(g Grid) {
	for i := range g.Attr {
		g.Attr[i] = 0
	}
	for i := range g.Rune {
		g.Rune[i] = 0
	}
}
