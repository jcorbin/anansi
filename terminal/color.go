// TODO naturalize ; imported from cops/{color,model,palettes,ansi}.go

package terminal

import (
	"image/color"
	"strconv"
)

// Colors contains the 256 color terminal palette.
// The first 8 correspond to 30-37 foreground and 40-47 background
// in ANSI escape sequences. The second 8 correspond to 90-97 and 100-107 or
// the high intensity variants of the first 8 in more advanced ANSI terminals.
// The next 6x6x6 colors are an RGB cube and the last 24 are shades of gray.
var Colors = []color.RGBA{
	{0, 0, 0, 255},
	{128, 0, 0, 255},
	{0, 128, 0, 255},
	{128, 128, 0, 255},
	{0, 0, 128, 255},
	{128, 0, 128, 255},
	{0, 128, 128, 255},
	{192, 192, 192, 255},
	{128, 128, 128, 255},
	{255, 0, 0, 255},
	{0, 255, 0, 255},
	{255, 255, 0, 255},
	{0, 0, 255, 255},
	{255, 0, 255, 255},
	{0, 255, 255, 255},
	{255, 255, 255, 255},
	{0, 0, 0, 255},
	{0, 0, 95, 255},
	{0, 0, 135, 255},
	{0, 0, 175, 255},
	{0, 0, 215, 255},
	{0, 0, 255, 255},
	{0, 95, 0, 255},
	{0, 95, 95, 255},
	{0, 95, 135, 255},
	{0, 95, 175, 255},
	{0, 95, 215, 255},
	{0, 95, 255, 255},
	{0, 135, 0, 255},
	{0, 135, 95, 255},
	{0, 135, 135, 255},
	{0, 135, 175, 255},
	{0, 135, 215, 255},
	{0, 135, 255, 255},
	{0, 175, 0, 255},
	{0, 175, 95, 255},
	{0, 175, 135, 255},
	{0, 175, 175, 255},
	{0, 175, 215, 255},
	{0, 175, 255, 255},
	{0, 215, 0, 255},
	{0, 215, 95, 255},
	{0, 215, 135, 255},
	{0, 215, 175, 255},
	{0, 215, 215, 255},
	{0, 215, 255, 255},
	{0, 255, 0, 255},
	{0, 255, 95, 255},
	{0, 255, 135, 255},
	{0, 255, 175, 255},
	{0, 255, 215, 255},
	{0, 255, 255, 255},
	{95, 0, 0, 255},
	{95, 0, 95, 255},
	{95, 0, 135, 255},
	{95, 0, 175, 255},
	{95, 0, 215, 255},
	{95, 0, 255, 255},
	{95, 95, 0, 255},
	{95, 95, 95, 255},
	{95, 95, 135, 255},
	{95, 95, 175, 255},
	{95, 95, 215, 255},
	{95, 95, 255, 255},
	{95, 135, 0, 255},
	{95, 135, 95, 255},
	{95, 135, 135, 255},
	{95, 135, 175, 255},
	{95, 135, 215, 255},
	{95, 135, 255, 255},
	{95, 175, 0, 255},
	{95, 175, 95, 255},
	{95, 175, 135, 255},
	{95, 175, 175, 255},
	{95, 175, 215, 255},
	{95, 175, 255, 255},
	{95, 215, 0, 255},
	{95, 215, 95, 255},
	{95, 215, 135, 255},
	{95, 215, 175, 255},
	{95, 215, 215, 255},
	{95, 215, 255, 255},
	{95, 255, 0, 255},
	{95, 255, 95, 255},
	{95, 255, 135, 255},
	{95, 255, 175, 255},
	{95, 255, 215, 255},
	{95, 255, 255, 255},
	{135, 0, 0, 255},
	{135, 0, 95, 255},
	{135, 0, 135, 255},
	{135, 0, 175, 255},
	{135, 0, 215, 255},
	{135, 0, 255, 255},
	{135, 95, 0, 255},
	{135, 95, 95, 255},
	{135, 95, 135, 255},
	{135, 95, 175, 255},
	{135, 95, 215, 255},
	{135, 95, 255, 255},
	{135, 135, 0, 255},
	{135, 135, 95, 255},
	{135, 135, 135, 255},
	{135, 135, 175, 255},
	{135, 135, 215, 255},
	{135, 135, 255, 255},
	{135, 175, 0, 255},
	{135, 175, 95, 255},
	{135, 175, 135, 255},
	{135, 175, 175, 255},
	{135, 175, 215, 255},
	{135, 175, 255, 255},
	{135, 215, 0, 255},
	{135, 215, 95, 255},
	{135, 215, 135, 255},
	{135, 215, 175, 255},
	{135, 215, 215, 255},
	{135, 215, 255, 255},
	{135, 255, 0, 255},
	{135, 255, 95, 255},
	{135, 255, 135, 255},
	{135, 255, 175, 255},
	{135, 255, 215, 255},
	{135, 255, 255, 255},
	{175, 0, 0, 255},
	{175, 0, 95, 255},
	{175, 0, 135, 255},
	{175, 0, 175, 255},
	{175, 0, 215, 255},
	{175, 0, 255, 255},
	{175, 95, 0, 255},
	{175, 95, 95, 255},
	{175, 95, 135, 255},
	{175, 95, 175, 255},
	{175, 95, 215, 255},
	{175, 95, 255, 255},
	{175, 135, 0, 255},
	{175, 135, 95, 255},
	{175, 135, 135, 255},
	{175, 135, 175, 255},
	{175, 135, 215, 255},
	{175, 135, 255, 255},
	{175, 175, 0, 255},
	{175, 175, 95, 255},
	{175, 175, 135, 255},
	{175, 175, 175, 255},
	{175, 175, 215, 255},
	{175, 175, 255, 255},
	{175, 215, 0, 255},
	{175, 215, 95, 255},
	{175, 215, 135, 255},
	{175, 215, 175, 255},
	{175, 215, 215, 255},
	{175, 215, 255, 255},
	{175, 255, 0, 255},
	{175, 255, 95, 255},
	{175, 255, 135, 255},
	{175, 255, 175, 255},
	{175, 255, 215, 255},
	{175, 255, 255, 255},
	{215, 0, 0, 255},
	{215, 0, 95, 255},
	{215, 0, 135, 255},
	{215, 0, 175, 255},
	{215, 0, 215, 255},
	{215, 0, 255, 255},
	{215, 95, 0, 255},
	{215, 95, 95, 255},
	{215, 95, 135, 255},
	{215, 95, 175, 255},
	{215, 95, 215, 255},
	{215, 95, 255, 255},
	{215, 135, 0, 255},
	{215, 135, 95, 255},
	{215, 135, 135, 255},
	{215, 135, 175, 255},
	{215, 135, 215, 255},
	{215, 135, 255, 255},
	{215, 175, 0, 255},
	{215, 175, 95, 255},
	{215, 175, 135, 255},
	{215, 175, 175, 255},
	{215, 175, 215, 255},
	{215, 175, 255, 255},
	{215, 215, 0, 255},
	{215, 215, 95, 255},
	{215, 215, 135, 255},
	{215, 215, 175, 255},
	{215, 215, 215, 255},
	{215, 215, 255, 255},
	{215, 255, 0, 255},
	{215, 255, 95, 255},
	{215, 255, 135, 255},
	{215, 255, 175, 255},
	{215, 255, 215, 255},
	{215, 255, 255, 255},
	{255, 0, 0, 255},
	{255, 0, 95, 255},
	{255, 0, 135, 255},
	{255, 0, 175, 255},
	{255, 0, 215, 255},
	{255, 0, 255, 255},
	{255, 95, 0, 255},
	{255, 95, 95, 255},
	{255, 95, 135, 255},
	{255, 95, 175, 255},
	{255, 95, 215, 255},
	{255, 95, 255, 255},
	{255, 135, 0, 255},
	{255, 135, 95, 255},
	{255, 135, 135, 255},
	{255, 135, 175, 255},
	{255, 135, 215, 255},
	{255, 135, 255, 255},
	{255, 175, 0, 255},
	{255, 175, 95, 255},
	{255, 175, 135, 255},
	{255, 175, 175, 255},
	{255, 175, 215, 255},
	{255, 175, 255, 255},
	{255, 215, 0, 255},
	{255, 215, 95, 255},
	{255, 215, 135, 255},
	{255, 215, 175, 255},
	{255, 215, 215, 255},
	{255, 215, 255, 255},
	{255, 255, 0, 255},
	{255, 255, 95, 255},
	{255, 255, 135, 255},
	{255, 255, 175, 255},
	{255, 255, 215, 255},
	{255, 255, 255, 255},
	{8, 8, 8, 255},
	{18, 18, 18, 255},
	{28, 28, 28, 255},
	{38, 38, 38, 255},
	{48, 48, 48, 255},
	{58, 58, 58, 255},
	{68, 68, 68, 255},
	{78, 78, 78, 255},
	{88, 88, 88, 255},
	{98, 98, 98, 255},
	{108, 108, 108, 255},
	{118, 118, 118, 255},
	{128, 128, 128, 255},
	{138, 138, 138, 255},
	{148, 148, 148, 255},
	{158, 158, 158, 255},
	{168, 168, 168, 255},
	{178, 178, 178, 255},
	{188, 188, 188, 255},
	{198, 198, 198, 255},
	{208, 208, 208, 255},
	{218, 218, 218, 255},
	{228, 228, 228, 255},
	{238, 238, 238, 255},
}

// ColorModel renders colors to a particular terminal color rendering protocol.
type ColorModel func(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor)

var (
	// Model0 is the monochrome color model, which does not print escape
	// sequences for any colors.
	Model0 ColorModel = renderNoColor

	// Model3 supports the first 8 color terminal palette.
	Model3 ColorModel = Palette3.Render

	// Model4 supports the first 16 color terminal palette, the same as Model3
	// but doubled for high intensity variants.
	Model4 ColorModel = Palette4.Render

	// Model8 supports a 256 color terminal palette, comprised of the 16
	// previous colors, a 6x6x6 color cube, and a 24 gray scale.
	Model8 ColorModel = Palette8.Render

	// Model24 supports all 24 bit colors, and renders only to 24-bit terminal
	// sequences.
	Model24 ColorModel = renderJustColor24

	// ModelCompat24 supports all 24 bit colors, using palette colors only for
	// exact matches.
	ModelCompat24 ColorModel = renderCompatColor24
)

// Palette is a limited palette of color for legacy terminals.
type Palette color.Palette

// Render the given colors to their closest palette equivalents.
func (tp Palette) Render(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		i := color.Palette.Index(color.Palette(tp), fg)
		buf = append(buf, fgColorStrings[i]...)
		cur.Foreground = fg
	}
	if bg != cur.Background {
		i := color.Palette.Index(color.Palette(tp), bg)
		buf = append(buf, bgColorStrings[i]...)
		cur.Background = bg
	}
	return buf, cur
}

var (
	// Palette3 contains the first 8 Colors.
	Palette3 Palette

	// Palette4 contains the first 16 Colors.
	Palette4 Palette

	// Palette8 contains all 256 paletted virtual terminal colors.
	Palette8 Palette

	// colorIndex maps colors back to their palette index, suitable for mapping
	// arbitrary colors back to palette indexes in the 24 bit color model.
	colorIndex map[color.RGBA]int
)

var (
	byteStrings    [256]string
	fgColorStrings [256]string
	bgColorStrings [256]string
)

// TODO codegen this
func init() {
	for i := 0; i < 8; i++ {
		Palette3 = append(Palette3, color.Color(Colors[i]))
	}
	Model3 = Palette3.Render

	for i := 0; i < 16; i++ {
		Palette4 = append(Palette4, color.Color(Colors[i]))
	}
	Model4 = Palette4.Render

	for i := 0; i < 256; i++ {
		Palette8 = append(Palette8, color.Color(Colors[i]))
	}
	Model8 = Palette8.Render

	colorIndex = make(map[color.RGBA]int, 256)
	for i := 0; i < 256; i++ {
		colorIndex[Colors[i]] = i
	}

	for i := 0; i < len(byteStrings); i++ {
		byteStrings[i] = ";" + strconv.Itoa(i)
	}

	i := 0
	for ; i < 8; i++ {
		fgColorStrings[i] = "\033[" + strconv.Itoa(30+i) + "m"
		bgColorStrings[i] = "\033[" + strconv.Itoa(40+i) + "m"
	}
	for ; i < 16; i++ {
		fgColorStrings[i] = "\033[" + strconv.Itoa(90-8+i) + "m"
		bgColorStrings[i] = "\033[" + strconv.Itoa(100-8+i) + "m"
	}
	for ; i < 256; i++ {
		fgColorStrings[i] = "\033[38;5;" + strconv.Itoa(i) + "m"
		bgColorStrings[i] = "\033[48;5;" + strconv.Itoa(i) + "m"
	}
}

func renderNoColor(buf []byte, cur Cursor, _, _ color.RGBA) ([]byte, Cursor) { return buf, cur }

func renderCompatColor24(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		if i, ok := colorIndex[fg]; ok {
			buf = append(buf, fgColorStrings[i]...)
		} else {
			buf = append(buf, "\033[38;2"...)
			buf = renderColor24(buf, fg)
		}
		cur.Foreground = fg
	}
	if bg != cur.Background {
		if i, ok := colorIndex[bg]; ok {
			buf = append(buf, bgColorStrings[i]...)
		} else {
			buf = append(buf, "\033[48;2"...)
			buf = renderColor24(buf, bg)
		}
		cur.Background = bg
	}
	return buf, cur
}

func renderJustColor24(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		buf = append(buf, "\033[38;2"...)
		buf = renderColor24(buf, fg)
		cur.Foreground = fg
	}
	if bg != cur.Background {
		buf = append(buf, "\033[48;2"...)
		buf = renderColor24(buf, bg)
		cur.Background = bg
	}
	return buf, cur
}

func renderColor24(buf []byte, c color.RGBA) []byte {
	buf = append(buf, byteStrings[c.R]...)
	buf = append(buf, byteStrings[c.G]...)
	buf = append(buf, byteStrings[c.B]...)
	buf = append(buf, "m"...)
	return buf
}
