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
		inLines  []string
		outLines []string
		styles   []Style
	}{
		{
			name: "basic test pattern",
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
				"⢕⢕",
				"⢕⢕",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf ansi.Buffer
			bi := NewBitmapString('#', tc.inLines...)
			RenderBitmap(&buf, bi, tc.styles...)
			assert.Equal(t, tc.outLines, strings.Split(string(buf.Bytes()), "\n"))
		})
	}
}
