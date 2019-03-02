package ansi

import (
	"fmt"
	"strconv"
	"strings"
)

// SGR Set Graphics Rendition (affects character attributes)
const (
	SGRCodeClear      = 0 // clear all special attributes
	SGRCodeBold       = 1 // bold or increased intensity
	SGRCodeDim        = 2 // dim or secondary color on gigi
	SGRCodeItalic     = 3 // italic
	SGRCodeUnderscore = 4 // underscore
	SGRCodeSlow       = 5 // slow blink
	SGRCodeFast       = 6 // fast blink
	SGRCodeNegative   = 7 // negative image
	SGRCodeConcealed  = 8 // concealed (do not display character echoed locally)

	// eliding uncommon font codes

	// uncommon vt220 codes
	SGRCodeCancelBold      = 22 // cancel bold or dim attribute only
	SGRCodeCancelUnderline = 24 // cancel underline attribute only
	SGRCodeCancelFast      = 25 // cancel fast or slow blink attribute only
	SGRCodeCancelNegative  = 27 // cancel negative image attribute only

	SGRCodeFGBlack   = 30 // write with black
	SGRCodeFGRed     = 31 // write with red
	SGRCodeFGGreen   = 32 // write with green
	SGRCodeFGYellow  = 33 // write with yellow
	SGRCodeFGBlue    = 34 // write with blue
	SGRCodeFGMagenta = 35 // write with magenta
	SGRCodeFGCyan    = 36 // write with cyan
	SGRCodeFGWhite   = 37 // write with white

	SGRCodeBGBlack   = 40 // set background to black
	SGRCodeBGRed     = 41 // set background to red
	SGRCodeBGGreen   = 42 // set background to green
	SGRCodeBGYellow  = 43 // set background to yellow
	SGRCodeBGBlue    = 44 // set background to blue
	SGRCodeBGMagenta = 45 // set background to magenta
	SGRCodeBGCyan    = 46 // set background to cyan
	SGRCodeBGWhite   = 47 // set background to white
)

// SGRReset resets graphic rendition (foreground, background, text and other
// character attributes) to default.
var SGRReset = SGR.With(SGRCodeClear)

// SGRAttr represents commonly used SGR attributes (ignoring blinks and fonts).
type SGRAttr uint64

// SGRClear is the zero value of SGRAttr, represents no attributes set, and
// will encode to an SGR clear code (CSI 0 m).
var SGRClear SGRAttr

// SGRAttr attribute bitfields.
const (
	// Causes a clear code to be written before the rest of any other attr
	// codes; the attr value isn't additive to whatever current state is.
	SGRAttrClear SGRAttr = 1 << iota

	// Bit fields for the 6 useful classic SGRCode*s
	SGRAttrBold
	SGRAttrDim
	SGRAttrItalic
	SGRAttrUnderscore
	SGRAttrNegative
	SGRAttrConceal

	sgrNumBits = iota
)

const (
	sgrColor24  SGRColor = 1 << 24 // 24-bit color flag
	sgrColorSet SGRColor = 1 << 25 // color set flag (only used when inside SGRAttr)

	sgrColorBitSize = 26
	sgrColorMask    = 0x01ffffff

	sgrFGShift = sgrNumBits
	sgrBGShift = sgrNumBits + sgrColorBitSize

	sgrAttrFGSet = SGRAttr(sgrColorSet) << sgrFGShift
	sgrAttrBGSet = SGRAttr(sgrColorSet) << sgrBGShift

	// SGRAttrMask selects all normal attr bits (excluding FG, BBG, and SGRAttrClear)
	SGRAttrMask = SGRAttrBold | SGRAttrDim | SGRAttrItalic | SGRAttrUnderscore | SGRAttrNegative | SGRAttrConceal

	// SGRAttrFGMask selects any set FG color.
	SGRAttrFGMask = SGRAttr(sgrColorSet|sgrColorMask) << sgrFGShift

	// SGRAttrBGMask selects any set BG color.
	SGRAttrBGMask = SGRAttr(sgrColorSet|sgrColorMask) << sgrBGShift
)

// SGRColor represents an SGR foreground or background color in any generation
// of color space.
type SGRColor uint32

// SGRColor constants.
const (
	// The first 8 colors from the 3-bit space.
	SGRBlack SGRColor = iota
	SGRRed
	SGRGreen
	SGRYellow
	SGRBlue
	SGRMagenta
	SGRCyan
	SGRWhite

	// The 8 high intensity colors from 4-bit space.
	SGRBrightBlack
	SGRBrightRed
	SGRBrightGreen
	SGRBrightYellow
	SGRBrightBlue
	SGRBrightMagenta
	SGRBrightCyan
	SGRBrightWhite

	// 8-bit color space: 216 color cube; see colors.go.
	SGRCube16
	SGRCube17
	SGRCube18
	SGRCube19
	SGRCube20
	SGRCube21
	SGRCube22
	SGRCube23
	SGRCube24
	SGRCube25
	SGRCube26
	SGRCube27
	SGRCube28
	SGRCube29
	SGRCube30
	SGRCube31
	SGRCube32
	SGRCube33
	SGRCube34
	SGRCube35
	SGRCube36
	SGRCube37
	SGRCube38
	SGRCube39
	SGRCube40
	SGRCube41
	SGRCube42
	SGRCube43
	SGRCube44
	SGRCube45
	SGRCube46
	SGRCube47
	SGRCube48
	SGRCube49
	SGRCube50
	SGRCube51
	SGRCube52
	SGRCube53
	SGRCube54
	SGRCube55
	SGRCube56
	SGRCube57
	SGRCube58
	SGRCube59
	SGRCube60
	SGRCube61
	SGRCube62
	SGRCube63
	SGRCube64
	SGRCube65
	SGRCube66
	SGRCube67
	SGRCube68
	SGRCube69
	SGRCube70
	SGRCube71
	SGRCube72
	SGRCube73
	SGRCube74
	SGRCube75
	SGRCube76
	SGRCube77
	SGRCube78
	SGRCube79
	SGRCube80
	SGRCube81
	SGRCube82
	SGRCube83
	SGRCube84
	SGRCube85
	SGRCube86
	SGRCube87
	SGRCube88
	SGRCube89
	SGRCube90
	SGRCube91
	SGRCube92
	SGRCube93
	SGRCube94
	SGRCube95
	SGRCube96
	SGRCube97
	SGRCube98
	SGRCube99
	SGRCube100
	SGRCube101
	SGRCube102
	SGRCube103
	SGRCube104
	SGRCube105
	SGRCube106
	SGRCube107
	SGRCube108
	SGRCube109
	SGRCube110
	SGRCube111
	SGRCube112
	SGRCube113
	SGRCube114
	SGRCube115
	SGRCube116
	SGRCube117
	SGRCube118
	SGRCube119
	SGRCube120
	SGRCube121
	SGRCube122
	SGRCube123
	SGRCube124
	SGRCube125
	SGRCube126
	SGRCube127
	SGRCube128
	SGRCube129
	SGRCube130
	SGRCube131
	SGRCube132
	SGRCube133
	SGRCube134
	SGRCube135
	SGRCube136
	SGRCube137
	SGRCube138
	SGRCube139
	SGRCube140
	SGRCube141
	SGRCube142
	SGRCube143
	SGRCube144
	SGRCube145
	SGRCube146
	SGRCube147
	SGRCube148
	SGRCube149
	SGRCube150
	SGRCube151
	SGRCube152
	SGRCube153
	SGRCube154
	SGRCube155
	SGRCube156
	SGRCube157
	SGRCube158
	SGRCube159
	SGRCube160
	SGRCube161
	SGRCube162
	SGRCube163
	SGRCube164
	SGRCube165
	SGRCube166
	SGRCube167
	SGRCube168
	SGRCube169
	SGRCube170
	SGRCube171
	SGRCube172
	SGRCube173
	SGRCube174
	SGRCube175
	SGRCube176
	SGRCube177
	SGRCube178
	SGRCube179
	SGRCube180
	SGRCube181
	SGRCube182
	SGRCube183
	SGRCube184
	SGRCube185
	SGRCube186
	SGRCube187
	SGRCube188
	SGRCube189
	SGRCube190
	SGRCube191
	SGRCube192
	SGRCube193
	SGRCube194
	SGRCube195
	SGRCube196
	SGRCube197
	SGRCube198
	SGRCube199
	SGRCube200
	SGRCube201
	SGRCube202
	SGRCube203
	SGRCube204
	SGRCube205
	SGRCube206
	SGRCube207
	SGRCube208
	SGRCube209
	SGRCube210
	SGRCube211
	SGRCube212
	SGRCube213
	SGRCube214
	SGRCube215
	SGRCube216
	SGRCube217
	SGRCube218
	SGRCube219
	SGRCube220
	SGRCube221
	SGRCube222
	SGRCube223
	SGRCube224
	SGRCube225
	SGRCube226
	SGRCube227
	SGRCube228
	SGRCube229
	SGRCube230
	SGRCube231

	// 8-bit color space: 24 shades of gray; see colors.go.
	SGRGray1
	SGRGray2
	SGRGray3
	SGRGray4
	SGRGray5
	SGRGray6
	SGRGray7
	SGRGray8
	SGRGray9
	SGRGray10
	SGRGray11
	SGRGray12
	SGRGray13
	SGRGray14
	SGRGray15
	SGRGray16
	SGRGray17
	SGRGray18
	SGRGray19
	SGRGray20
	SGRGray21
	SGRGray22
	SGRGray23
	SGRGray24
)

// RGB constructs a 24-bit SGR color from component values.
func RGB(r, g, b uint8) SGRColor {
	return sgrColor24 | SGRColor(r) | SGRColor(g)<<8 | SGRColor(b)<<16
}

// RGBA creates an SGRColor from color.Color() alpha-premultiplied values,
// ignoring the alpha value. Clips components to 0xffff before converting to
// 24-bit color (8-bit per channel).
func RGBA(r, g, b, _ uint32) SGRColor {
	if r > 0xffff {
		r = 0xffff
	}
	if g > 0xffff {
		g = 0xffff
	}
	if b > 0xffff {
		b = 0xffff
	}
	return RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func (c SGRColor) String() string {
	switch {
	case c&sgrColor24 != 0:
		var tmp [12]byte
		p := c.appendRGB(tmp[:0])[1:]
		for i := range p {
			if p[i] == ';' {
				p[i] = ','
			}
		}
		return fmt.Sprintf("rgb(%s)", p)
	case int(c) < len(colorNames):
		return colorNames[c]
	default:
		return fmt.Sprintf("color%d", c)
	}
}

// FG constructs an SGR attribute value with the color as foreground.
func (c SGRColor) FG() SGRAttr {
	return sgrAttrFGSet | SGRAttr(c&sgrColorMask)<<sgrFGShift
}

// BG constructs an SGR attribute value with the color as background.
func (c SGRColor) BG() SGRAttr {
	return sgrAttrBGSet | SGRAttr((c&sgrColorMask))<<sgrBGShift
}

// RGBA implements the color.Color interface.
func (c SGRColor) RGBA() (r, g, b, a uint32) {
	r8, g8, b8 := c.RGB()
	r = uint32(r8)
	g = uint32(g8)
	b = uint32(b8)
	return r | r<<8, g | g<<8, b | b<<8, 0xffff
}

// RGB returns the equivalent RGB components.
func (c SGRColor) RGB() (r, g, b uint8) {
	if c&sgrColor24 == 0 {
		c = Palette8Colors[c&0xff]
	}
	return uint8(c), uint8(c >> 8), uint8(c >> 16)
}

// To24Bit converts the color to 24-bit mode, so that it won't encode as a
// legacy 3, 4, or 8-bit color.
func (c SGRColor) To24Bit() SGRColor {
	if c&sgrColor24 != 0 {
		return c
	}
	return RGB(c.RGB())
}

func (c SGRColor) appendFGTo(p []byte) []byte {
	switch {
	case c&sgrColor24 != 0:
		return c.appendRGB(append(p, "38;2"...)) // TODO support color space identifier?
	case c <= SGRWhite:
		return append(p, '3', '0'+uint8(c))
	case c <= SGRBrightWhite:
		return append(p, '9', '0'+uint8(c)-8)
	case c <= SGRGray24:
		return append(append(p, "38;5"...), colorStrings[uint8(c)]...)
	}
	return p
}

func (c SGRColor) fgSize() int {
	switch {
	case c&sgrColor24 != 0:
		return 4 + c.rgbSize()
	case c <= SGRWhite:
		return 2
	case c <= SGRBrightWhite:
		return 2
	case c <= SGRGray24:
		return 4 + len(colorStrings[uint8(c)])
	}
	return 0
}

func (c SGRColor) appendBGTo(p []byte) []byte {
	switch {
	case c&sgrColor24 != 0:
		return c.appendRGB(append(p, "48;2"...)) // TODO support color space identifier?
	case c <= SGRWhite:
		return append(p, '4', '0'+uint8(c))
	case c <= SGRBrightWhite:
		return append(p, '1', '0', '0'+uint8(c)-8)
	case c <= SGRGray24:
		return append(append(p, "48;5"...), colorStrings[uint8(c)]...)
	}
	return p
}

func (c SGRColor) bgSize() int {
	switch {
	case c&sgrColor24 != 0:
		return 4 + c.rgbSize()
	case c <= SGRWhite:
		return 2
	case c <= SGRBrightWhite:
		return 3
	case c <= SGRGray24:
		return 4 + len(colorStrings[uint8(c)])
	}
	return 0
}

func (c SGRColor) appendRGB(p []byte) []byte {
	p = append(p, colorStrings[uint8(c)]...)
	p = append(p, colorStrings[uint8(c>>8)]...)
	p = append(p, colorStrings[uint8(c>>16)]...)
	return p
}

func (c SGRColor) rgbSize() int {
	return 0 +
		len(colorStrings[uint8(c)]) +
		len(colorStrings[uint8(c>>8)]) +
		len(colorStrings[uint8(c>>16)])
}

func (c SGRColor) addFGSeq(seq Seq) Seq {
	switch {
	case c&sgrColor24 != 0:
		return seq.WithInts(38, 2, int(c), int(c>>8), int(c>>16)) // TODO support color space identifier?
	case c <= SGRWhite:
		return seq.WithInts(30 + int(c))
	case c <= SGRBrightWhite:
		return seq.WithInts(90 + int(c) - 8)
	case c <= SGRGray24:
		return seq.WithInts(38, 5, int(c))
	}
	return seq
}

func (c SGRColor) addBGSeq(seq Seq) Seq {
	switch {
	case c&sgrColor24 != 0:
		return seq.WithInts(48, 2, int(c), int(c>>8), int(c>>16)) // TODO support color space identifier?
	case c <= SGRWhite:
		return seq.WithInts(40 + int(c))
	case c <= SGRBrightWhite:
		return seq.WithInts(100 + int(c) - 8)
	case c <= SGRGray24:
		return seq.WithInts(48, 5, int(c))
	}
	return seq
}

var colorNames = [16]string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",

	"bright-black",
	"bright-red",
	"bright-green",
	"bright-yellow",
	"bright-blue",
	"bright-magenta",
	"bright-cyan",
	"bright-white",
}

var colorStrings [256]string

func init() {
	for i := 0; i < 256; i++ {
		colorStrings[i] = ";" + strconv.Itoa(i)
	}
}

// FG returns any set foreground color, and a bool indicating if it was
// actually set (to distinguish from 0=black).
func (attr SGRAttr) FG() (c SGRColor, set bool) {
	if set = attr&sgrAttrFGSet != 0; set {
		c = SGRColor(attr>>sgrFGShift) & sgrColorMask
	}
	return c, set
}

// BG returns any set background color, and a bool indicating if it was
// actually set (to distinguish from 0=black).
func (attr SGRAttr) BG() (c SGRColor, set bool) {
	if set = attr&sgrAttrBGSet != 0; set {
		c = SGRColor(attr>>sgrBGShift) & sgrColorMask
	}
	return c, set
}

// SansFG returns a copy of the attribute with any FG color unset.
func (attr SGRAttr) SansFG() SGRAttr { return attr & ^SGRAttrFGMask }

// SansBG returns a copy of the attribute with any BG color unset.
func (attr SGRAttr) SansBG() SGRAttr { return attr & ^SGRAttrBGMask }

// Merge an other attr value into a copy of the receiver, returning it.
func (attr SGRAttr) Merge(other SGRAttr) SGRAttr {
	if other&SGRAttrClear != 0 {
		attr = SGRClear
	}
	attr |= other & SGRAttrMask
	if c, set := other.FG(); set {
		attr = attr.SansFG() | c.FG()
	}
	if c, set := other.BG(); set {
		attr = attr.SansBG() | c.BG()
	}
	return attr
}

// Diff returns the attr value which must be merged with the receiver to result
// in the given value.
func (attr SGRAttr) Diff(other SGRAttr) SGRAttr {
	if other&SGRAttrClear != 0 {
		return other
	}

	const (
		fgMask = sgrAttrFGSet | (sgrColorMask << sgrFGShift)
		bgMask = sgrAttrBGSet | (sgrColorMask << sgrBGShift)
	)

	var (
		attrFlags    = attr & SGRAttrMask
		otherFlags   = other & SGRAttrMask
		changedFlags = attrFlags ^ otherFlags
		goneFlags    = attrFlags & changedFlags
		attrFG       = attr & fgMask
		attrBG       = attr & bgMask
		otherFG      = other & fgMask
		otherBG      = other & bgMask
	)

	if goneFlags != 0 ||
		(otherFG == 0 && attrFG != 0) ||
		(otherBG == 0 && attrBG != 0) {
		other |= SGRAttrClear
		return other
	}

	diff := otherFlags & changedFlags
	if otherFG != attrFG {
		diff |= otherFG
	}
	if otherBG != attrBG {
		diff |= otherBG
	}
	return diff
}

// ControlString returns the appropriate ansi SGR control sequence as a string
// value.
func (attr SGRAttr) ControlString() string {
	// TODO cache
	var b [64]byte // over-estimate of max space for SGRAttr.AppendTo
	p := attr.AppendTo(b[:0])
	return string(p)
}

// AppendTo appends the appropriate ansi SGR control sequence to the given byte
// slice to affect any set bits or fg/bg colors in attr. If no bits or colors
// are set, append a clear code.
func (attr SGRAttr) AppendTo(p []byte) []byte {
	if attr == 0 || attr == SGRAttrClear {
		return SGR.AppendWith(p, '0')
	}
	p = SGR.AppendTo(p)
	final := p[len(p)-1]
	p = p[:len(p)-1]

	// attr arguments
	first := true
	for i, b := range []byte{
		'0', // SGRAttrClear
		'1', // SGRAttrBold
		'2', // SGRAttrDim
		'3', // SGRAttrItalic
		'4', // SGRAttrUnderscore
		'7', // SGRAttrNegative
		'8', // SGRAttrConceal
	} {
		if attr&(1<<uint(i)) != 0 {
			if first {
				p = append(p, b)
				first = false
			} else {
				p = append(p, ';', b)
			}
		}
	}

	// any fg color
	if fg, set := attr.FG(); set {
		if first {
			first = false
		} else {
			p = append(p, ';')
		}
		p = fg.appendFGTo(p)
	}

	// any bg color
	if bg, set := attr.BG(); set {
		if first {
			first = false
		} else {
			p = append(p, ';')
		}
		p = bg.appendBGTo(p)
	}

	return append(p, final)
}

// Seq constructs a control sequence value for the SGR attribute, including
// arguments for attributes, foreground and background color as appropriate.
func (attr SGRAttr) Seq() Seq {
	seq := SGR.Seq()

	if attr == 0 || attr == SGRAttrClear {
		return seq.WithInts(0)
	}

	// attr arguments
	for i, b := range []byte{
		'0', // SGRAttrClear
		'1', // SGRAttrBold
		'2', // SGRAttrDim
		'3', // SGRAttrItalic
		'4', // SGRAttrUnderscore
		'7', // SGRAttrNegative
		'8', // SGRAttrConceal
	} {
		if attr&(1<<uint(i)) != 0 {
			seq = seq.WithInts(int(b))
		}
	}

	// any fg color
	if fg, set := attr.FG(); set {
		seq = fg.addFGSeq(seq)
	}

	// any bg color
	if bg, set := attr.BG(); set {
		seq = bg.addBGSeq(seq)
	}

	return seq
}

// Size returns the number of bytes needed to encode the SGR control sequence needed.
func (attr SGRAttr) Size() int {
	n := -1 // discount the first over-counted ';' below
	if attr&SGRAttrClear != 0 {
		n += 2
	}
	if attr&SGRAttrBold != 0 {
		n += 2
	}
	if attr&SGRAttrDim != 0 {
		n += 2
	}
	if attr&SGRAttrItalic != 0 {
		n += 2
	}
	if attr&SGRAttrUnderscore != 0 {
		n += 2
	}
	if attr&SGRAttrNegative != 0 {
		n += 2
	}
	if attr&SGRAttrConceal != 0 {
		n += 2
	}
	if fg, set := attr.FG(); set {
		n += 1 + fg.fgSize()
	}
	if bg, set := attr.BG(); set {
		n += 1 + bg.bgSize()
	}
	if n < 0 {
		n = 1 // no args added, will append a clear code
	}
	n += 3 // CSI ... m
	return n
}

func (attr SGRAttr) String() string {
	parts := make([]string, 0, 8)
	if attr&SGRAttrClear != 0 {
		parts = append(parts, "clear")
	}
	if attr&SGRAttrBold != 0 {
		parts = append(parts, "bold")
	}
	if attr&SGRAttrDim != 0 {
		parts = append(parts, "dim")
	}
	if attr&SGRAttrItalic != 0 {
		parts = append(parts, "italic")
	}
	if attr&SGRAttrUnderscore != 0 {
		parts = append(parts, "underscore")
	}
	if attr&SGRAttrNegative != 0 {
		parts = append(parts, "negative")
	}
	if attr&SGRAttrConceal != 0 {
		parts = append(parts, "conceal")
	}
	if fg, set := attr.FG(); set {
		parts = append(parts, "fg:"+fg.String())
	}
	if bg, set := attr.BG(); set {
		parts = append(parts, "bg:"+bg.String())
	}
	// let implicit clear stand as ""
	return strings.Join(parts, " ")
}
