package anansi_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
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
			var buf ansi.Buffer
			RenderBitmap(&buf, tc.bi, tc.styles...)
			assert.Equal(t, tc.outLines, strings.Split(string(buf.Bytes()), "\n"))
		})
	}
}
