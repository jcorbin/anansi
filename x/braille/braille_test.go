package braille_test

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/anansi"
	anansitest "github.com/jcorbin/anansi/test"
	. "github.com/jcorbin/anansi/x/braille"
)

func TestBitmap_CopyInto(t *testing.T) {
	for _, tc := range []struct {
		name     string
		gridSize image.Point
		inLines  []string
		outLines []string
		at       image.Point
		styles   []Style
	}{
		{
			name:     "basic test pattern",
			gridSize: image.Pt(3, 3),
			at:       image.Pt(1, 1), // cell space origin
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
			var g anansi.Grid
			g.Resize(tc.gridSize)
			bi := NewBitmapString('#', tc.inLines...)
			bi.CopyInto(g, tc.at, tc.styles...)
			assert.Equal(t, tc.outLines, anansitest.GridLines(g, '_'))
		})
	}
}
