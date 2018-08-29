package anansi_test

import (
	"bytes"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

func TestScreen(t *testing.T) {
	type step struct {
		run    func(*Screen)
		expect string
	}
	for _, tc := range []struct {
		name  string
		steps []step
	}{
		{"roving set", []step{
			{func(sc *Screen) {
				sc.Clear()
			}, "\x1b[?25l\x1b[2J"},
			{func(sc *Screen) {
				sc.Clear()
			}, ""},
			{func(sc *Screen) {
				sc.Clear()
				sc.Set(image.Pt(3, 3), '@', ansi.SGRBrightGreen.FG())
			}, "\x1b[3;3H\x1b[0;92m@"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Set(image.Pt(4, 3), '@', ansi.SGRBrightYellow.FG())
			}, "\x1b[D\x1b[0m \x1b[93m@"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Set(image.Pt(4, 4), '@', ansi.SGRGreen.FG())
			}, "\x1b[D\x1b[0m \x1b[4;4H\x1b[32m@"}, // 5,4
			{func(sc *Screen) {
				sc.Clear()
				sc.Set(image.Pt(3, 4), '@', ansi.SGRYellow.FG())
			}, "\x1b[2D\x1b[33m@\x1b[0m "},
		}},

		{"write over", []step{
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("hello world!")
			}, "\x1b[?25l\x1b[2J\x1b[1;1H\x1b[0mhello worl\r\nd!"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("\x1b[34mhello world!")
			}, "\x1b[1;1H\x1b[34mhello worl\r\nd!"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("\x1b[34mhello\x1b[0m")
				sc.WriteString(" ")
				sc.WriteString("\x1b[33mworld!\x1b[0m")
			}, "\x1b[1;6H\x1b[0m \x1b[33mworl\r\nd!"},
		}},

		{"dangling style", []step{
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("0) --")
			}, "\x1b[?25l\x1b[2J\x1b[1;1H\x1b[0m0) --"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("1) ")
				sc.WriteString("\x1b[31mred")
			}, "\x1b[5D1) \x1b[31mred"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("2) ")
				sc.WriteString("\x1b[32mgreen")
			}, "\x1b[6D\x1b[0m2) \x1b[32mgreen"},
		}},

		{"writing", []step{
			{func(sc *Screen) {
				sc.Clear()
			}, "\x1b[?25l\x1b[2J"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("hello world")
			}, "\x1b[1;1H\x1b[0mhello worl\r\nd"},
			{func(sc *Screen) {
				sc.To(image.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteSGR(ansi.SGRRed.FG())
				sc.WriteString("world")
			}, "\x1b[1;7H\x1b[31mworl\r\nd"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteString("\x1b[32mworld")
			}, "\x1b[1;7H\x1b[32mworl\r\nd"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Invalidate()
				sc.To(image.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteString("\x1b[33mworld")
			}, "\x1b[2J\x1b[1;1H\x1b[0mhello \x1b[33mworl\r\nd"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Resize(image.Pt(20, 10))
				sc.To(image.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteString("\x1b[34mworld")
			}, "\x1b[2J\x1b[1;1H\x1b[0mhello \x1b[34mworld"},
		}},

		// TODO UserCursor
	} {
		t.Run(tc.name, logBuf.With(func(t *testing.T) {
			var out bytes.Buffer
			var sc Screen
			sc.Resize(image.Pt(10, 10))
			for i, step := range tc.steps {
				out.Reset()
				step.run(&sc)
				_, err := sc.WriteTo(&out)
				require.NoError(t, err)
				assert.Equal(t, step.expect, out.String(), "[%d] expected output", i)
				t.Logf("[%d] %q", i, out.Bytes())
			}
		}))
	}
}
