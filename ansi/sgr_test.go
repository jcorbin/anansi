package ansi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/anansi/ansi"
)

func TestSGRAttr(t *testing.T) {
	for _, tc := range []struct {
		attr   ansi.SGRAttr
		expect string
	}{
		// some classics
		{ansi.SGRAttrClear, "\x1b[0m"},
		{ansi.SGRBlack.FG(), "\x1b[30m"},
		{ansi.SGRBlack.BG(), "\x1b[40m"},
		{ansi.SGRBlack.FG() | ansi.SGRBlack.BG(), "\x1b[30;40m"},
		{ansi.SGRAttrClear | ansi.SGRBlack.FG() | ansi.SGRBlack.BG(), "\x1b[0;30;40m"},
		{ansi.SGRRed.FG(), "\x1b[31m"},
		{ansi.SGRGreen.FG(), "\x1b[32m"},
		{ansi.SGRYellow.FG(), "\x1b[33m"},
		{ansi.SGRBlue.FG(), "\x1b[34m"},
		{ansi.SGRMagenta.FG(), "\x1b[35m"},
		{ansi.SGRCyan.FG(), "\x1b[36m"},
		{ansi.SGRWhite.FG(), "\x1b[37m"},

		// brights
		{ansi.SGRBrightYellow.FG(), "\x1b[93m"},
		{ansi.SGRBrightBlue.BG(), "\x1b[104m"},
		{ansi.SGRBrightYellow.FG() | ansi.SGRBrightBlue.BG(), "\x1b[93;104m"},

		// some 256 colors
		{ansi.SGRCube42.FG(), "\x1b[38;5;42m"},
		{ansi.SGRCube108.BG(), "\x1b[48;5;108m"},
		{ansi.SGRCube42.BG() | ansi.SGRCube108.FG(), "\x1b[38;5;108;48;5;42m"},

		// some 24-bit colors
		{ansi.SGRRed.To24Bit().FG(), "\x1b[38;2;128;0;0m"},
		{ansi.SGRCyan.To24Bit().BG(), "\x1b[48;2;0;128;128m"},
		{ansi.SGRGreen.To24Bit().FG() | ansi.SGRBlue.To24Bit().BG(),
			"\x1b[38;2;0;128;0;48;2;0;0;128m"},
	} {
		t.Run(tc.expect, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.attr.String())
		})
	}
}
