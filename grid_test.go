package anansi_test

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	anansitest "github.com/jcorbin/anansi/test"
)

func TestGrid_SubRect(t *testing.T) {
	type subResult struct {
		desc string
		r    ansi.Rectangle
		out  []string
	}
	for _, tc := range []struct {
		name string
		data []string
		subs []subResult
	}{
		{
			name: "basic 5x5",
			data: []string{
				"12345",
				"67890",
				"abcde",
				"fghij",
				"klmno",
			},
			subs: []subResult{
				{"equal rect", ansi.Rect(1, 1, 6, 6), []string{
					"12345",
					"67890",
					"abcde",
					"fghij",
					"klmno",
				}},

				{"rect max clamp", ansi.Rect(1, 1, 16, 16), []string{
					"12345",
					"67890",
					"abcde",
					"fghij",
					"klmno",
				}},
				{"rect min beyond", ansi.Rect(11, 11, 16, 16), nil},

				{"less column-1", ansi.Rect(2, 1, 6, 6), []string{
					"2345",
					"7890",
					"bcde",
					"ghij",
					"lmno",
				}},
				{"less row-1", ansi.Rect(1, 2, 6, 6), []string{
					"67890",
					"abcde",
					"fghij",
					"klmno",
				}},
				{"less column-5", ansi.Rect(1, 1, 5, 6), []string{
					"1234",
					"6789",
					"abcd",
					"fghi",
					"klmn",
				}},
				{"less row-5", ansi.Rect(1, 1, 6, 5), []string{
					"12345",
					"67890",
					"abcde",
					"fghij",
				}},
				{"inset-1", ansi.Rect(2, 2, 5, 5), []string{
					"789",
					"bcd",
					"ghi",
				}},
				{"inset-2", ansi.Rect(3, 3, 4, 4), []string{
					"c",
				}},

				{"top-left corner", ansi.Rect(1, 1, 4, 4), []string{
					"123",
					"678",
					"abc",
				}},
				{"top-right corner", ansi.Rect(3, 1, 6, 4), []string{
					"345",
					"890",
					"cde",
				}},
				{"bottom-right corner", ansi.Rect(3, 3, 6, 6), []string{
					"cde",
					"hij",
					"mno",
				}},
				{"bottom-left corner", ansi.Rect(1, 3, 4, 6), []string{
					"abc",
					"fgh",
					"klm",
				}},
			},
		},
		// TODO cases that use SGR attrs
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := anansitest.ParseGridLines(tc.data)
			for _, sub := range tc.subs {
				t.Run(sub.desc, func(t *testing.T) {
					out := anansitest.GridLines(g.SubRect(sub.r), '.')
					assert.Equal(t, sub.out, out, "sub %v", sub.r)
				})
			}
		})
	}
}

func TestGrid_Clear(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func() Grid
		out   []string
	}{

		{
			name: "basic",
			build: func() Grid {
				var g Grid
				g.Resize(image.Pt(10, 10))
				for i := range g.Rune {
					g.Rune[i] = 'A'
					g.Attr[i] = ansi.SGRAttrBold
				}
				g.Clear()
				return g
			},
			out: []string{
				"..........",
				"..........",
				"..........",
				"..........",
				"..........",
				"..........",
				"..........",
				"..........",
				"..........",
				"..........",
			},
		},

		{
			name: "middle",
			build: func() Grid {
				var g Grid
				g.Resize(image.Pt(10, 10))
				for i := range g.Rune {
					g.Rune[i] = 'A'
					g.Attr[i] = ansi.SGRAttrBold
				}
				g.SubAt(ansi.Pt(3, 3)).SubSize(image.Pt(5, 5)).Clear()
				return g
			},
			out: []string{
				"\x1b[1mAAAAAAAAAA",
				"AAAAAAAAAA",
				"AA\x1b[0m.....\x1b[1mAAA",
				"AA\x1b[0m.....\x1b[1mAAA",
				"AA\x1b[0m.....\x1b[1mAAA",
				"AA\x1b[0m.....\x1b[1mAAA",
				"AA\x1b[0m.....\x1b[1mAAA",
				"AAAAAAAAAA",
				"AAAAAAAAAA",
				"AAAAAAAAAA",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := tc.build()
			out := anansitest.GridLines(g, '.')
			assert.Equal(t, tc.out, out)
		})
	}
}
