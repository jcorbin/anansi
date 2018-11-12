package anansi_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi"
)

func TestRenderBitmap(t *testing.T) {
	for _, tc := range []struct {
		name     string
		bi       *Bitmap
		outLines []string
		styles   []Style
	}{
		{
			name: "basic test pattern",
			bi: NewBitmap(MustParseBitmap("# ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
				"# . # . ",
				". # . # ",
			)),
			outLines: []string{
				"⢕⢕",
				"⢕⢕",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf Buffer
			RenderBitmap(&buf, tc.bi, tc.styles...)
			assert.Equal(t, tc.outLines, strings.Split(string(buf.Bytes()), "\n"))
		})
	}
}
