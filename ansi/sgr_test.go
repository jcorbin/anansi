package ansi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/anansi/ansi"
)

func TestSGRAttr(t *testing.T) {
	for _, tc := range []struct {
		name string
		attr ansi.SGRAttr
		code string
	}{
		// some classics
		{"clear", ansi.SGRAttrClear, "\x1b[0m"},
		{"fg:black", ansi.SGRBlack.FG(), "\x1b[30m"},
		{"bg:black", ansi.SGRBlack.BG(), "\x1b[40m"},
		{"fg:black bg:black", ansi.SGRBlack.FG() | ansi.SGRBlack.BG(), "\x1b[30;40m"},
		{"clear fg:black bg:black", ansi.SGRAttrClear | ansi.SGRBlack.FG() | ansi.SGRBlack.BG(), "\x1b[0;30;40m"},
		{"fg:red", ansi.SGRRed.FG(), "\x1b[31m"},
		{"fg:green", ansi.SGRGreen.FG(), "\x1b[32m"},
		{"fg:yellow", ansi.SGRYellow.FG(), "\x1b[33m"},
		{"fg:blue", ansi.SGRBlue.FG(), "\x1b[34m"},
		{"fg:magenta", ansi.SGRMagenta.FG(), "\x1b[35m"},
		{"fg:cyan", ansi.SGRCyan.FG(), "\x1b[36m"},
		{"fg:white", ansi.SGRWhite.FG(), "\x1b[37m"},

		// brights
		{"fg:bright-yellow", ansi.SGRBrightYellow.FG(), "\x1b[93m"},
		{"bg:bright-blue", ansi.SGRBrightBlue.BG(), "\x1b[104m"},
		{"fg:bright-yellow bg:bright-blue", ansi.SGRBrightYellow.FG() | ansi.SGRBrightBlue.BG(), "\x1b[93;104m"},

		// some 256 colors
		{"fg:color42", ansi.SGRCube42.FG(), "\x1b[38;5;42m"},
		{"bg:color108", ansi.SGRCube108.BG(), "\x1b[48;5;108m"},
		{"fg:color108 bg:color42", ansi.SGRCube42.BG() | ansi.SGRCube108.FG(), "\x1b[38;5;108;48;5;42m"},

		// some 24-bit colors
		{"fg:rgb(128,0,0)", ansi.SGRRed.To24Bit().FG(), "\x1b[38;2;128;0;0m"},
		{"bg:rgb(0,128,128)", ansi.SGRCyan.To24Bit().BG(), "\x1b[48;2;0;128;128m"},
		{"fg:rgb(0,128,0) bg:rgb(0,0,128)", ansi.SGRGreen.To24Bit().FG() | ansi.SGRBlue.To24Bit().BG(),
			"\x1b[38;2;0;128;0;48;2;0;0;128m"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.name, tc.attr.String(), "expected name string")
			assert.Equal(t, tc.code, string(tc.attr.AppendTo(nil)), "expected code string")
		})
	}
}
