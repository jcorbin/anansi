package anansi_test

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	anansitest "github.com/jcorbin/anansi/test"
)

func TestScreenState_processing(t *testing.T) {
	for _, tc := range []struct {
		name  string
		size  image.Point
		input string
		lines []string
	}{
		{
			name: "empty",
			size: image.Pt(10, 5),
			lines: []string{
				"          ",
				"          ",
				"          ",
				"          ",
				"          ",
			},
		},

		{
			name:  "hello-lf-world",
			size:  image.Pt(20, 10),
			input: "\x1b[5;5Hhello\nworld",
			lines: []string{
				"                    ",
				"                    ",
				"                    ",
				"                    ",
				"    hello           ",
				"         world      ",
				"                    ",
				"                    ",
				"                    ",
				"                    ",
			},
		},

		{
			name:  "hello-crlf-world",
			size:  image.Pt(10, 5),
			input: "\x1b[2;1Hhello\r\nworld",
			lines: []string{
				"          ",
				"hello     ",
				"world     ",
				"          ",
				"          ",
			},
		},

		{
			name:  "hello-cud-cub-world",
			size:  image.Pt(20, 10),
			input: "\x1b[5;5Hhello\x1b[B\x1b[5Dworld",
			lines: []string{
				"                    ",
				"                    ",
				"                    ",
				"                    ",
				"    hello           ",
				"    world           ",
				"                    ",
				"                    ",
				"                    ",
				"                    ",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf ansi.Buffer
			var scs ScreenState
			scs.Resize(tc.size)
			buf.WriteString(tc.input)
			buf.Process(&scs)
			assert.Equal(t, tc.lines, anansitest.GridLines(scs.Grid, ' '))
		})
	}
}
