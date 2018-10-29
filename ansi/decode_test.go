package ansi_test

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/anansi/ansi"
)

func TestDecodeEscape(t *testing.T) {
	type anRead struct {
		e ansi.Escape
		a []byte
		n int
	}
	type utRead struct {
		r rune
		m int
	}
	type ev struct {
		anRead
		utRead
	}

	tcs := []struct {
		in  string
		out []ev
	}{
		{"", nil},
		{"hi", []ev{
			{anRead{}, utRead{'h', 1}},
			{anRead{}, utRead{'i', 1}},
		}},
		{"hi\x7fello", []ev{
			{anRead{}, utRead{'h', 1}},
			{anRead{}, utRead{'i', 1}},
			{anRead{}, utRead{'\x7f', 1}},
			{anRead{}, utRead{'e', 1}},
			{anRead{}, utRead{'l', 1}},
			{anRead{}, utRead{'l', 1}},
			{anRead{}, utRead{'o', 1}},
		}},

		{"iab\x1b\x1b\x18jk", []ev{
			{anRead{}, utRead{'i', 1}},
			{anRead{}, utRead{'a', 1}},
			{anRead{}, utRead{'b', 1}},
			{anRead{}, utRead{'\x1b', 1}},
			{anRead{0, nil, 2}, utRead{'j', 1}},
			{anRead{}, utRead{'k', 1}},
		}},

		{"\x1b>num\x1b=app", []ev{
			{anRead{ansi.Escape(0xEF3E), nil, 2}, utRead{}},
			{anRead{}, utRead{'n', 1}},
			{anRead{}, utRead{'u', 1}},
			{anRead{}, utRead{'m', 1}},
			{anRead{ansi.Escape(0xEF3D), nil, 2}, utRead{}},
			{anRead{}, utRead{'a', 1}},
			{anRead{}, utRead{'p', 1}},
			{anRead{}, utRead{'p', 1}},
		}},
		{"\x1baint", []ev{
			{anRead{ansi.Escape(0xEF61), nil, 2}, utRead{}},
			{anRead{}, utRead{'i', 1}},
			{anRead{}, utRead{'n', 1}},
			{anRead{}, utRead{'t', 1}},
		}},

		{"\x1b(B$", []ev{
			{anRead{ansi.Escape(0xEF28), []byte("B"), 3}, utRead{}},
			{anRead{}, utRead{'$', 1}},
		}},
		{"\x1b\x03(B$", []ev{
			{anRead{0, nil, 0}, utRead{'\x03', 1}},
			{anRead{ansi.Escape(0xEF28), []byte("B"), 3}, utRead{}},
			{anRead{}, utRead{'$', 1}},
		}},
		{"\x1b\x03(\x04B$", []ev{
			{anRead{0, nil, 0}, utRead{'\x03', 1}},
			{anRead{0, nil, 0}, utRead{'\x04', 1}},
			{anRead{ansi.Escape(0xEF28), []byte("B"), 3}, utRead{}},
			{anRead{}, utRead{'$', 1}},
		}},

		{"\x1bø", []ev{
			{anRead{ansi.Escape(0xEF78), nil, 3}, utRead{}},
		}},

		{"\x1b“(B$", []ev{
			{anRead{}, utRead{'“', 3}},
			{anRead{ansi.Escape(0xEF28), []byte("B"), 3}, utRead{}},
			{anRead{}, utRead{'$', 1}},
		}},
		{"\x1b“(”B$", []ev{
			{anRead{}, utRead{'“', 3}},
			{anRead{}, utRead{'”', 3}},
			{anRead{ansi.Escape(0xEF28), []byte("B"), 3}, utRead{}},
			{anRead{}, utRead{'$', 1}},
		}},

		{"\x1b[31mred", []ev{
			{anRead{ansi.Escape(0xEFED), []byte("31"), 5}, utRead{}},
			{anRead{}, utRead{'r', 1}},
			{anRead{}, utRead{'e', 1}},
			{anRead{}, utRead{'d', 1}},
		}},

		{"\u009b31mred", []ev{
			{anRead{ansi.Escape(0xEFED), []byte("31"), 5}, utRead{}},
			{anRead{}, utRead{'r', 1}},
			{anRead{}, utRead{'e', 1}},
			{anRead{}, utRead{'d', 1}},
		}},

		{"(\x1bPdemo\x1b\\)", []ev{
			{anRead{}, utRead{'(', 1}},
			{anRead{ansi.Escape(0x90), []byte("demo"), 8}, utRead{}},
			{anRead{}, utRead{')', 1}},
		}},

		{"(\u0090demo\u009C)", []ev{
			{anRead{}, utRead{'(', 1}},
			{anRead{ansi.Escape(0x90), []byte("demo"), 8}, utRead{}},
			{anRead{}, utRead{')', 1}},
		}},
	}

	const sanity = 100
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%q", tc.in), func(t *testing.T) {
			i := 0
			var out []ev

			for p := []byte(tc.in); len(p) > 0; i++ {
				if i > sanity {
					require.Fail(t, "sanity exhausted")
				}
				var ev ev
				ev.e, ev.a, ev.n = ansi.DecodeEscape(p)
				p = p[ev.n:]
				if ev.e == 0 {
					ev.r, ev.m = utf8.DecodeRune(p)
					p = p[ev.m:]
				}
				out = append(out, ev)
			}
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestDecodeNumber(t *testing.T) {
	for _, tc := range []struct {
		in  string
		out int
		rem string
	}{

		{"1", 1, ""},
		{"2", 2, ""},
		{"10", 10, ""},
		{"20", 20, ""},
		{"13", 13, ""},
		{"24", 24, ""},

		{";1", 1, ""},
		{";2", 2, ""},
		{";10", 10, ""},
		{";20", 20, ""},
		{";13", 13, ""},
		{";24", 24, ""},

		{"-1", -1, ""},
		{"-2", -2, ""},
		{"-10", -10, ""},
		{"-20", -20, ""},
		{"-13", -13, ""},
		{"-24", -24, ""},

		{"1;42", 1, ";42"},
		{"2;42", 2, ";42"},
		{"10;42", 10, ";42"},
		{"20;42", 20, ";42"},
		{"13;42", 13, ";42"},
		{"24;42", 24, ";42"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			p := []byte(tc.in)
			v, n, err := ansi.DecodeNumber(p)
			if assert.NoError(t, err, "unexpected decode error") {
				assert.Equal(t, tc.out, v, "expected value")
				assert.Equal(t, tc.rem, string(p[n:]), "expected remain")
			}
		})
	}
}

func TestDecodeSGR_roundtrips(t *testing.T) {
	for _, tc := range []struct {
		attr ansi.SGRAttr
		str  string
	}{
		{ansi.SGRAttrClear, "\x1b[0m"},
		{ansi.SGRAttrNegative, "\x1b[7m"},

		{ansi.SGRRed.FG(), "\x1b[31m"},
		{ansi.SGRBrightGreen.FG(), "\x1b[92m"},
		{ansi.SGRCube20.FG(), "\x1b[38;5;20m"},
		{ansi.SGRGray10.FG(), "\x1b[38;5;241m"},
		{ansi.RGB(10, 20, 30).FG(), "\x1b[38;2;10;20;30m"},

		{ansi.SGRRed.BG(), "\x1b[41m"},
		{ansi.SGRBrightGreen.BG(), "\x1b[102m"},
		{ansi.SGRCube20.BG(), "\x1b[48;5;20m"},
		{ansi.SGRGray10.BG(), "\x1b[48;5;241m"},
		{ansi.RGB(10, 20, 30).BG(), "\x1b[48;2;10;20;30m"},

		{ansi.SGRAttrBold | ansi.SGRAttrNegative, "\x1b[1;7m"},
		{ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRAttrNegative, "\x1b[0;1;7m"},

		{ansi.SGRRed.FG() | ansi.SGRRed.BG(), "\x1b[31;41m"},
		{ansi.SGRAttrClear | ansi.SGRRed.FG() | ansi.SGRGreen.BG(), "\x1b[0;31;42m"},
		{ansi.SGRAttrBold | ansi.SGRRed.FG() | ansi.SGRGreen.BG(), "\x1b[1;31;42m"},
		{ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRRed.FG() | ansi.SGRGreen.BG(), "\x1b[0;1;31;42m"},
		{ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRRed.To24Bit().FG() | ansi.SGRGreen.To24Bit().BG(), "\x1b[0;1;38;2;128;0;0;48;2;0;128;0m"},
	} {
		t.Run(tc.str, func(t *testing.T) {
			p := tc.attr.AppendTo(nil)
			e, a, n := ansi.DecodeEscape(p)
			assert.Equal(t, len(p), n)
			require.Equal(t, ansi.SGR, e, "expected full escape decode")
			require.Equal(t, tc.str, string(p), "expected control sequence")

			attr, n, err := ansi.DecodeSGR(a)
			if err != nil {
				err = fmt.Errorf("%v @%v in %q", err, n, a)
			}
			require.NoError(t, err)
			assert.Equal(t, len(a), n, "expected full arg decode")
			if !assert.Equal(t, tc.attr, attr) {
				t.Logf("Encode %016x", uint64(tc.attr))
				t.Logf(
					"clear:%t bold:%t dim:%t italic:%t underscore:%t negative:%t conceal:%t",
					(tc.attr&ansi.SGRAttrClear) != 0,
					(tc.attr&ansi.SGRAttrBold) != 0,
					(tc.attr&ansi.SGRAttrDim) != 0,
					(tc.attr&ansi.SGRAttrItalic) != 0,
					(tc.attr&ansi.SGRAttrUnderscore) != 0,
					(tc.attr&ansi.SGRAttrNegative) != 0,
					(tc.attr&ansi.SGRAttrConceal) != 0,
				)

				t.Logf("Decode %q => %016x", p, uint64(attr))
				t.Logf(
					"clear:%t bold:%t dim:%t italic:%t underscore:%t negative:%t conceal:%t",
					(attr&ansi.SGRAttrClear) != 0,
					(attr&ansi.SGRAttrBold) != 0,
					(attr&ansi.SGRAttrDim) != 0,
					(attr&ansi.SGRAttrItalic) != 0,
					(attr&ansi.SGRAttrUnderscore) != 0,
					(attr&ansi.SGRAttrNegative) != 0,
					(attr&ansi.SGRAttrConceal) != 0,
				)

			}
		})
	}
}

func TestPoint_roundtrip(t *testing.T) {
	for _, tc := range []struct {
		p ansi.Point
	}{
		{
			p: ansi.Pt(3, 5),
		},
	} {
		t.Run(tc.p.String(), func(t *testing.T) {
			seq := ansi.CUP.WithPoint(tc.p)
			b := seq.AppendTo(nil)
			e, a, n := ansi.DecodeEscape(b)
			b = b[n:]
			require.Equal(t, ansi.CUP, e)
			require.Equal(t, 0, len(b))
			p, n, err := ansi.DecodePoint(a)
			a = a[n:]
			require.NoError(t, err)
			require.Equal(t, 0, len(a))
			assert.Equal(t, tc.p, p)
		})
	}
}
