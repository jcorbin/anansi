package anansi_test

import (
	"bytes"
	"image"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/anansi"
	anansitest "github.com/jcorbin/anansi/test"
)

func makeTestBitmap(set string, lines ...string) Bitmap {
	var bi Bitmap
	bi.Load(MustParseBitmap(set, lines...))
	return bi
}

func TestWriteBitmap(t *testing.T) {
	for _, tc := range []struct {
		name     string
		sz       image.Point
		bi       Bitmap
		outLines []string
		styles   []Style
	}{

		{
			name: "basic test pattern",
			sz:   image.Pt(2, 2),
			bi: makeTestBitmap("# ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
			),
			outLines: []string{
				"⢕⢕",
				"⢕⢕",
			},
		},

		{
			name: "basic test pattern",
			sz:   image.Pt(3, 3),
			bi: makeTestBitmap("# ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
			),
			outLines: []string{
				"⢕⢕ ",
				"⢕⢕ ",
				"   ",
			},
		},
	} {
		t.Run(tc.name, testWithScreenModes(tc.sz, tc.outLines, func(t *testing.T, w io.Writer) error {
			_, err := WriteBitmap(w, tc.bi, tc.styles...)
			return err
		}))
	}
}

// TODO func TestWriteGrid

func testWithScreenModes(
	sz image.Point, outLines []string,
	f func(t *testing.T, w io.Writer) error,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("direct to anansi.Screen", func(t *testing.T) {
			var sc ScreenDiffer
			sc.Resize(sz)
			require.NoError(t, f(t, &sc))
			outLines := anansitest.GridLines(sc.Grid, ' ')
			assert.Equal(t, outLines, outLines)
		})
		t.Run("buffered flush to anansi.Screen", func(t *testing.T) {
			var sc ScreenDiffer
			sc.Resize(sz)
			var buf bytes.Buffer
			err := f(t, &buf)
			if err == nil {
				_, err = buf.WriteTo(&sc)
			}
			require.NoError(t, err)
			outLines := anansitest.GridLines(sc.Grid, ' ')
			assert.Equal(t, outLines, outLines)
		})
	}
}
