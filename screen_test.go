package anansi_test

import (
	"bytes"
	"fmt"
	"image"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	anansitest "github.com/jcorbin/anansi/test"
)

func TestScreen_steps(t *testing.T) {
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
				i, _ := sc.CellOffset(ansi.Pt(3, 3))
				sc.Grid.Rune[i], sc.Grid.Attr[i] = '@', ansi.SGRBrightGreen.FG()
			}, "\x1b[3;3H\x1b[0;92m@"},
			{func(sc *Screen) {
				sc.Clear()
				i, _ := sc.CellOffset(ansi.Pt(4, 3))
				sc.Grid.Rune[i], sc.Grid.Attr[i] = '@', ansi.SGRBrightYellow.FG()
			}, "\x1b[D\x1b[0m \x1b[93m@"},
			{func(sc *Screen) {
				sc.Clear()
				i, _ := sc.CellOffset(ansi.Pt(4, 4))
				sc.Grid.Rune[i], sc.Grid.Attr[i] = '@', ansi.SGRGreen.FG()
			}, "\x1b[D\x1b[0m \x1b[4;4H\x1b[32m@"}, // 5,4
			{func(sc *Screen) {
				sc.Clear()
				i, _ := sc.CellOffset(ansi.Pt(3, 4))
				sc.Grid.Rune[i], sc.Grid.Attr[i] = '@', ansi.SGRYellow.FG()
			}, "\x1b[2D\x1b[33m@\x1b[0m "},
		}},

		{"write over", []step{
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("hello world!")
			}, "\x1b[?25l\x1b[2J\x1b[1;1H\x1b[0mhello worl\r\nd!"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("\x1b[34mhello world!")
			}, "\x1b[1;1H\x1b[34mhello worl\r\nd!"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("\x1b[34mhello\x1b[0m")
				sc.WriteString(" ")
				sc.WriteString("\x1b[33mworld!\x1b[0m")
			}, "\x1b[1;6H\x1b[0m \x1b[33mworl\r\nd!"},
		}},

		{"dangling style", []step{
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("0) --")
			}, "\x1b[?25l\x1b[2J\x1b[1;1H\x1b[0m0) --"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("1) ")
				sc.WriteString("\x1b[31mred")
			}, "\x1b[5D1\x1b[2C\x1b[31mred"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("2) ")
				sc.WriteString("\x1b[32mgreen")
			}, "\x1b[6D\x1b[0m2\x1b[2C\x1b[32mgreen"},
		}},

		{"writing", []step{
			{func(sc *Screen) {
				sc.Clear()
			}, "\x1b[?25l\x1b[2J"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("hello world")
			}, "\x1b[1;1H\x1b[0mhello\x1b[Cworl\r\nd"},
			{func(sc *Screen) {
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteSGR(ansi.SGRRed.FG())
				sc.WriteString("world")
			}, "\x1b[1;7H\x1b[31mworl\r\nd"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteString("\x1b[32mworld")
			}, "\x1b[1;7H\x1b[32mworl\r\nd"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Invalidate()
				sc.To(ansi.Pt(1, 1))
				sc.WriteString("hello ")
				sc.WriteString("\x1b[33mworld")
			}, "\x1b[2J\x1b[1;1H\x1b[0mhello \x1b[33mworl\r\nd"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Resize(image.Pt(20, 10))
				sc.To(ansi.Pt(1, 1))
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

func TestScreen_blobs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		size  image.Point
		input string
		lines []string
	}{
		{
			name:  "3-line hello world",
			size:  image.Pt(10, 5),
			input: "hello\r\nworld\r\nagain",
			lines: []string{
				"hello     ",
				"world     ",
				"again     ",
				"          ",
				"          ",
			},
		},

		{
			name:  "3-line hello world, sans CR",
			size:  image.Pt(20, 5),
			input: "hello\nworld\nagain",
			lines: []string{
				"hello               ",
				"     world          ",
				"          again     ",
				"                    ",
				"                    ",
			},
		},

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
		t.Run(tc.name, logBuf.With(func(t *testing.T) {
			var sc Screen
			sc.Resize(tc.size)
			sc.WriteString(tc.input)
			assert.Equal(t, tc.lines, anansitest.GridLines(sc.Grid, ' '))
		}))
	}
}

func Test_gridLines(t *testing.T) {
	for _, tc := range []struct {
		name  string
		size  image.Point
		in    string
		lines []string
	}{
		{
			name: "5x5 room",
			size: image.Pt(10, 10),
			in: "" +
				"\x1b[3;3H\x1b[32m#####" +
				"\x1b[4;3H#\x1b[4;7H#" +
				"\x1b[5;3H#\x1b[5;7H#" +
				"\x1b[6;3H#\x1b[6;7H#" +
				"\x1b[7;3H#####",
			lines: []string{
				"          ",
				"          ",
				"  \x1b[32m#####\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#####\x1b[0m   ",
				"          ",
				"          ",
				"          ",
			},
		},

		{
			name: "player inline 5x5 room",
			size: image.Pt(10, 10),
			in: "" +
				"\x1b[3;3H\x1b[32m#####" +
				"\x1b[4;3H#\x1b[4;7H#" +
				"\x1b[5;3H#" +
				"\x1b[5;5H\x1b[31m@" +
				"\x1b[5;7H\x1b[32m#" +
				"\x1b[6;3H#\x1b[6;7H#" +
				"\x1b[7;3H#####",
			lines: []string{
				"          ",
				"          ",
				"  \x1b[32m#####\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#\x1b[0m \x1b[31m@\x1b[0m \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#####\x1b[0m   ",
				"          ",
				"          ",
				"          ",
			},
		},

		{
			name: "player after 5x5 room",
			size: image.Pt(10, 10),
			in: "" +
				"\x1b[3;3H\x1b[32m#####" +
				"\x1b[4;3H#\x1b[4;7H#" +
				"\x1b[5;3H#\x1b[5;7H#" +
				"\x1b[6;3H#\x1b[6;7H#" +
				"\x1b[7;3H#####" +
				"\x1b[5;5H\x1b[31m@",
			lines: []string{
				"          ",
				"          ",
				"  \x1b[32m#####\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#\x1b[0m \x1b[31m@\x1b[0m \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#\x1b[0m   \x1b[32m#\x1b[0m   ",
				"  \x1b[32m#####\x1b[0m   ",
				"          ",
				"          ",
				"          ",
			},
		},
	} {
		t.Run(tc.name, logBuf.With(func(t *testing.T) {
			g := parseGrid(tc.in, tc.size)
			assert.Equal(t, tc.lines, anansitest.GridLines(g, ' '))
			if t.Failed() {
				rs, as := anansitest.GridRowData(g)
				for i := range rs {
					t.Logf("rs[%v]: %q", i, rs[i])
				}
				for i := range as {
					t.Logf("as[%v]: %q", i, as[i])
				}
			}
		}))
	}
}

func parseGrid(s string, sz image.Point) Grid {
	var sc Screen
	sc.Resize(sz)
	sc.WriteString(s)
	return sc.Grid
}

// TestScreen_equiv tests screen grid diffing by functional equivalence with a
// full redraw. A test case is a series of grid states on a statically sized
// grid. The test then loads each grid into a pair of independent screens.
// The first screen is told to writes its output into an output screen, ala
// Screen.WriteTo(). Then the second screen is told to write its output into
// another output screen, but with a full redraw forced, ala
// Screen.Invalidate(). The contents of both output screens is then tested
// for equivalence.
func TestScreen_equiv(t *testing.T) {
	for _, tc := range []struct {
		name  string
		sz    image.Point
		steps []string
	}{
		{
			name: "room",
			sz:   image.Pt(10, 10),
			steps: []string{
				"",

				"" +
					"\x1b[3;3H\x1b[32m#####" +
					"\x1b[4;3H#\x1b[4;7H#" +
					"\x1b[5;3H#\x1b[5;7H#" +
					"\x1b[6;3H#\x1b[6;7H#" +
					"\x1b[7;3H#####",

				"" +
					"\x1b[3;3H\x1b[32m#####" +
					"\x1b[4;3H#\x1b[4;7H#" +
					"\x1b[5;3H#\x1b[5;7H#" +
					"\x1b[6;3H#\x1b[6;7H#" +
					"\x1b[7;3H#####" +
					"\x1b[5;5H\x1b[31m@",
			},
		},
	} {
		t.Run(tc.name, logBuf.With(func(t *testing.T) {
			var a, b, aout, bout Screen
			a.Resize(tc.sz)
			b.Resize(tc.sz)
			aout.Resize(tc.sz)
			bout.Resize(tc.sz)

			for i, s := range tc.steps {
				t.Run(fmt.Sprintf("step_%d", i), logBuf.With(func(t *testing.T) {
					a.Grid = parseGrid(s, tc.sz)
					b.Grid = parseGrid(s, tc.sz)

					_, err := a.WriteTo(&aout)
					require.NoError(t, err, "unexpected write error")
					aLines := anansitest.GridLines(aout.Grid, ' ')

					b.Invalidate()
					_, err = b.WriteTo(&bout)
					require.NoError(t, err, "unexpected write error")
					bLines := anansitest.GridLines(bout.Grid, ' ')

					var aw, bw int
					for i := range aLines {
						aLines[i] = strconv.Quote(aLines[i])
						bLines[i] = strconv.Quote(bLines[i])
						if n := utf8.RuneCountInString(aLines[i]); aw < n {
							aw = n
						}
						if n := utf8.RuneCountInString(bLines[i]); bw < n {
							bw = n
						}
					}
					for i := range aLines {
						t.Logf("%*s %*s", aw, aLines[i], bw, bLines[i])
					}

					assert.Equal(t, aLines, bLines, "[%v] expected equivalent output", i)
				}))
			}
		}))
	}
}
