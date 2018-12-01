package platform_test

import (
	"bufio"
	"bytes"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/anansi/ansi"
	. "github.com/jcorbin/anansi/x/platform"
)

type testStep struct {
	in     string
	out    string
	expect func(t *testing.T)
}

type testSteps []testStep

func (steps testSteps) run(t *testing.T, sz image.Point, client Client) {
	var out bytes.Buffer
	p := NewTest(sz, client)
	for i, step := range steps {
		t.Logf("[%d] input %q", i, step.in)
		ctx := p.Context()
		ctx.Input.Clear()
		ctx.Input.DecodeBytes([]byte(step.in))
		ctx.Update()
		require.NoError(t, ctx.Err, "[%d] unexpected update error", i)
		for sc := bufio.NewScanner(&Logs); sc.Scan(); {
			t.Logf("[%d] log: %s", i, sc.Bytes())
		}
		out.Reset()
		_, _ = ctx.Output.WriteTo(&out)
		assert.Equal(t, step.out, out.String(), "[%d] expected output", i)
		if step.expect != nil {
			step.expect(t)
		}
	}
}

func TestEditLine(t *testing.T) {
	var result string
	expectResult := func(expected string) func(*testing.T) {
		return func(t *testing.T) {
			assert.Equal(t, expected, result, "expected result")
		}
	}

	var edl EditLine
	client := ClientFunc(func(ctx *Context) error {
		sz := ctx.Output.Bounds().Size()
		edl.Box = ansi.Rect(sz.X/4, sz.Y/2, 3*sz.X/4, sz.Y/2+1)
		edl.Update(ctx)
		if edl.Done() {
			result = string(edl.Buf)
		} else if edl.Canceled() {
			result = "<CANCELED>"
		}
		return nil
	})

	for _, tc := range []struct {
		name  string
		steps testSteps
	}{

		{"one and gtfo", testSteps{
			{
				in: "",
				out: "\x1b[?25l" + // hide cursor
					"\x1b[2J" + // erase display
					"\x1b[5;5H" + // CUP 5,5 -- show cursor at start of edit line
					"\x1b[0m" + // SGR clear
					"\x1b[?25h", // show cursor
				expect: expectResult(""),
			},
			{
				in:     "\x1b",
				out:    "\x1b[?25l", // hide cursor
				expect: expectResult("<CANCELED>"),
			},
		}},

		{"hello", testSteps{
			{
				in: "",
				out: "\x1b[?25l\x1b[2J" +
					"\x1b[5;5H\x1b[0m\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "h",
				out:    "\x1b[?25lh\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "e",
				out:    "\x1b[?25le\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "l",
				out:    "\x1b[?25ll\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "l",
				out:    "\x1b[?25ll\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "o",
				out:    "\x1b[?25lo\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x0d",
				out: "\x1b[?25l" +
					"\x1B[5D     ",
				expect: expectResult("hello"),
			},
		}},

		{"many backspace", testSteps{
			{
				in: "",
				out: "\x1b[?25l\x1b[2J" +
					"\x1b[5;5H\x1b[0m\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "hello",
				out:    "\x1b[?25lhello\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f",
				out: "\x1b[?25l" +
					"\x1b[5D     " +
					"\x1b[5D\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "\x0d",
				out:    "\x1b[?25l",
				expect: expectResult(""),
			},
		}},

		{"hello alice<BS>bob", testSteps{
			{
				in: "",
				out: "\x1b[?25l\x1b[2J" +
					"\x1b[5;5H\x1b[0m\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "hello alice",
				out:    "\x1b[?25lllo\x1b[Calice\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x7f\x7f\x7f\x7f\x7f",
				out: "\x1b[?25l" +
					"\x1b[9Dhello    " +
					"\x1b[3D\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "bob",
				out:    "\x1b[?25lbob\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x0d",
				out: "\x1b[?25l" +
					"\x1b[9D     \x1b[C   ",
				expect: expectResult("hello bob"),
			},
		}},

		{"hello arrows", testSteps{
			{
				in: "",
				out: "\x1b[?25l\x1b[2J" +
					"\x1b[5;5H\x1b[0m\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "hello alice",
				out:    "\x1b[?25lllo\x1b[Calice\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x1b[5D",
				out: "\x1b[?25l" +
					"\x1b[9De\x1b[Clo alice" +
					"\x1b[5D\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in:     "\x1b[D",
				out:    "\x1b[?25l\x1b[D\x1b[?25h", // TODO ideally optimize this to one CUB
				expect: expectResult(""),
			},
			{
				in: ",",
				out: "\x1b[?25l" +
					", alic" +
					"\x1b[5D\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x1b[80C!",
				out: "\x1b[?25l" +
					"\x1b[5Do, alice! " +
					"\x1b[D\x1b[?25h",
				expect: expectResult(""),
			},
			{
				in: "\x0d",
				out: "\x1b[?25l" +
					"\x1b[9D  \x1b[C      ",
				expect: expectResult("hello, alice!"),
			},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edl.Reset()
			result = ""
			tc.steps.run(t, image.Pt(20, 10), client)
		})
	}
}
