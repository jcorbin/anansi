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
				sc.Cell(image.Pt(3, 3)).Set('@', ansi.SGRBrightGreen.FG())
			}, "\x1b[3;3H\x1b[0;92m@"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Cell(image.Pt(4, 3)).Set('@', ansi.SGRBrightYellow.FG())
			}, "\x1b[D\x1b[0m \x1b[93m@"},
			{func(sc *Screen) {
				sc.Clear()
				sc.Cell(image.Pt(4, 4)).Set('@', ansi.SGRGreen.FG())
			}, "\x1b[D\x1b[0m \x1b[4;4H\x1b[32m@"}, // 5,4
			{func(sc *Screen) {
				sc.Clear()
				sc.Cell(image.Pt(3, 4)).Set('@', ansi.SGRYellow.FG())
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
			}, "\x1b[5D1\x1b[2C\x1b[31mred"},
			{func(sc *Screen) {
				sc.Clear()
				sc.To(image.Pt(1, 1))
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
			assert.Equal(t, tc.lines, gridLines(g, ' '))
			if t.Failed() {
				rs, as := gridRowData(g)
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

func gridLines(g Grid, fill rune) (lines []string) {
	var ca ansi.SGRAttr
	for i, p := 0, image.ZP; p.Y < g.Size.Y; p.Y++ {
		var b []byte
		for p.X = 0; p.X < g.Size.X; p.X++ {
			r, a := g.Rune[i], g.Attr[i]
			if a != ca {
				a = ca.Diff(a)
				b = a.AppendTo(b)
				ca = ca.Merge(a)
			}
			var tmp [4]byte
			if r == 0 {
				r = fill
			}
			b = append(b, tmp[:utf8.EncodeRune(tmp[:], r)]...)
			i++
		}
		lines = append(lines, string(b))
	}
	return lines
}

func gridRowData(g Grid) (rs [][]rune, as [][]ansi.SGRAttr) {
	// NOTE p is in array space, not 1,1-based screen space
	for i, p := 0, image.ZP; p.Y < g.Size.Y; p.Y++ {
		j := p.Y * g.Size.X
		rs = append(rs, g.Rune[i:j])
		as = append(as, g.Attr[i:j])
		i = j
	}
	return rs, as
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
			name: "empty",
			sz:   image.Pt(10, 10),
			steps: []string{
				"",
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
					aLines := gridLines(aout.Grid, ' ')

					b.Invalidate()
					_, err = b.WriteTo(&bout)
					require.NoError(t, err, "unexpected write error")
					bLines := gridLines(bout.Grid, ' ')

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
