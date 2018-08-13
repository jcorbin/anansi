package ansi_test

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/anansi/ansi"
)

func TestDecodeUTF8Escape(t *testing.T) {
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
