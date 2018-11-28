package ansi

// Color definitions taken from https://en.wikipedia.org/wiki/ANSI_escape_code#Colors.

// Palette is a limited palette of color for legacy terminals.
type Palette []SGRColor

// Palette3 is the classic 3-bit palette of 8 colors.
var Palette3 = Palette{
	RGB(0x00, 0x00, 0x00), // SGRBlack
	RGB(0x80, 0x00, 0x00), // SGRRed
	RGB(0x00, 0x80, 0x00), // SGRGreen
	RGB(0x80, 0x80, 0x00), // SGRYellow
	RGB(0x00, 0x00, 0x80), // SGRBlue
	RGB(0x80, 0x00, 0x80), // SGRMagenta
	RGB(0x00, 0x80, 0x80), // SGRCyan
	RGB(0xC0, 0xC0, 0xC0), // SGRWhite
}

// Palette4 is the extended 4-bit palette of the 8 classic colors and their
// bright counterparts.
var Palette4 = Palette3.concat(
	RGB(0x80, 0x80, 0x80), // SGRBrightBlack
	RGB(0xFF, 0x00, 0x00), // SGRBrightRed
	RGB(0x00, 0xFF, 0x00), // SGRBrightGreen
	RGB(0xFF, 0xFF, 0x00), // SGRBrightYellow
	RGB(0x00, 0x00, 0xFF), // SGRBrightBlue
	RGB(0xFF, 0x00, 0xFF), // SGRBrightMagenta
	RGB(0x00, 0xFF, 0xFF), // SGRBrightCyan
	RGB(0xFF, 0xFF, 0xFF), // SGRBrightWhite
)

// Palette8 is the extended 8-bit palette of the first 16 extended colors, a
// 6x6x6=216 color cube, and 24 shades of gray.
var Palette8 = Palette4.concat(
	// plane 1, row 1
	RGB(0x00, 0x00, 0x00), // SGRCube16
	RGB(0x00, 0x00, 0x5F), // SGRCube17
	RGB(0x00, 0x00, 0x87), // SGRCube18
	RGB(0x00, 0x00, 0xAF), // SGRCube19
	RGB(0x00, 0x00, 0xD7), // SGRCube20
	RGB(0x00, 0x00, 0xFF), // SGRCube21

	// plane 1, row 2
	RGB(0x00, 0x5F, 0x00), // SGRCube22
	RGB(0x00, 0x5F, 0x5F), // SGRCube23
	RGB(0x00, 0x5F, 0x87), // SGRCube24
	RGB(0x00, 0x5F, 0xAF), // SGRCube25
	RGB(0x00, 0x5F, 0xD7), // SGRCube26
	RGB(0x00, 0x5F, 0xFF), // SGRCube27

	// plane 1, row 3
	RGB(0x00, 0x87, 0x00), // SGRCube28
	RGB(0x00, 0x87, 0x5F), // SGRCube29
	RGB(0x00, 0x87, 0x87), // SGRCube30
	RGB(0x00, 0x87, 0xAF), // SGRCube31
	RGB(0x00, 0x87, 0xD7), // SGRCube32
	RGB(0x00, 0x87, 0xFF), // SGRCube33

	// plane 1, row 4
	RGB(0x00, 0xAF, 0x00), // SGRCube34
	RGB(0x00, 0xAF, 0x5F), // SGRCube35
	RGB(0x00, 0xAF, 0x87), // SGRCube36
	RGB(0x00, 0xAF, 0xAF), // SGRCube37
	RGB(0x00, 0xAF, 0xD7), // SGRCube38
	RGB(0x00, 0xAF, 0xFF), // SGRCube39

	// plane 1, row 5
	RGB(0x00, 0xD7, 0x00), // SGRCube40
	RGB(0x00, 0xD7, 0x5F), // SGRCube41
	RGB(0x00, 0xD7, 0x87), // SGRCube42
	RGB(0x00, 0xD7, 0xAF), // SGRCube43
	RGB(0x00, 0xD7, 0xD7), // SGRCube44
	RGB(0x00, 0xD7, 0xFF), // SGRCube45

	// plane 1, row 6
	RGB(0x00, 0xFF, 0x00), // SGRCube46
	RGB(0x00, 0xFF, 0x5F), // SGRCube47
	RGB(0x00, 0xFF, 0x87), // SGRCube48
	RGB(0x00, 0xFF, 0xAF), // SGRCube49
	RGB(0x00, 0xFF, 0xD7), // SGRCube50
	RGB(0x00, 0xFF, 0xFF), // SGRCube51

	// plane 2, row 1
	RGB(0x5F, 0x00, 0x00), // SGRCube52
	RGB(0x5F, 0x00, 0x5F), // SGRCube53
	RGB(0x5F, 0x00, 0x87), // SGRCube54
	RGB(0x5F, 0x00, 0xAF), // SGRCube55
	RGB(0x5F, 0x00, 0xD7), // SGRCube56
	RGB(0x5F, 0x00, 0xFF), // SGRCube57

	// plane 2, row 2
	RGB(0x5F, 0x5F, 0x00), // SGRCube58
	RGB(0x5F, 0x5F, 0x5F), // SGRCube59
	RGB(0x5F, 0x5F, 0x87), // SGRCube60
	RGB(0x5F, 0x5F, 0xAF), // SGRCube61
	RGB(0x5F, 0x5F, 0xD7), // SGRCube62
	RGB(0x5F, 0x5F, 0xFF), // SGRCube63

	// plane 2, row 3
	RGB(0x5F, 0x87, 0x00), // SGRCube64
	RGB(0x5F, 0x87, 0x5F), // SGRCube65
	RGB(0x5F, 0x87, 0x87), // SGRCube66
	RGB(0x5F, 0x87, 0xAF), // SGRCube67
	RGB(0x5F, 0x87, 0xD7), // SGRCube68
	RGB(0x5F, 0x87, 0xFF), // SGRCube69

	// plane 2, row 4
	RGB(0x5F, 0xAF, 0x00), // SGRCube70
	RGB(0x5F, 0xAF, 0x5F), // SGRCube71
	RGB(0x5F, 0xAF, 0x87), // SGRCube72
	RGB(0x5F, 0xAF, 0xAF), // SGRCube73
	RGB(0x5F, 0xAF, 0xD7), // SGRCube74
	RGB(0x5F, 0xAF, 0xFF), // SGRCube75

	// plane 2, row 5
	RGB(0x5F, 0xD7, 0x00), // SGRCube76
	RGB(0x5F, 0xD7, 0x5F), // SGRCube77
	RGB(0x5F, 0xD7, 0x87), // SGRCube78
	RGB(0x5F, 0xD7, 0xAF), // SGRCube79
	RGB(0x5F, 0xD7, 0xD7), // SGRCube80
	RGB(0x5F, 0xD7, 0xFF), // SGRCube81

	// plane 2, row 6
	RGB(0x5F, 0xFF, 0x00), // SGRCube82
	RGB(0x5F, 0xFF, 0x5F), // SGRCube83
	RGB(0x5F, 0xFF, 0x87), // SGRCube84
	RGB(0x5F, 0xFF, 0xAF), // SGRCube85
	RGB(0x5F, 0xFF, 0xD7), // SGRCube86
	RGB(0x5F, 0xFF, 0xFF), // SGRCube87

	// plane 3, row 1
	RGB(0x87, 0x00, 0x00), // SGRCube88
	RGB(0x87, 0x00, 0x5F), // SGRCube89
	RGB(0x87, 0x00, 0x87), // SGRCube90
	RGB(0x87, 0x00, 0xAF), // SGRCube91
	RGB(0x87, 0x00, 0xD7), // SGRCube92
	RGB(0x87, 0x00, 0xFF), // SGRCube93

	// plane 3, row 2
	RGB(0x87, 0x5F, 0x00), // SGRCube94
	RGB(0x87, 0x5F, 0x5F), // SGRCube95
	RGB(0x87, 0x5F, 0x87), // SGRCube96
	RGB(0x87, 0x5F, 0xAF), // SGRCube97
	RGB(0x87, 0x5F, 0xD7), // SGRCube98
	RGB(0x87, 0x5F, 0xFF), // SGRCube99

	// plane 3, row 3
	RGB(0x87, 0x87, 0x00), // SGRCube100
	RGB(0x87, 0x87, 0x5F), // SGRCube101
	RGB(0x87, 0x87, 0x87), // SGRCube102
	RGB(0x87, 0x87, 0xAF), // SGRCube103
	RGB(0x87, 0x87, 0xD7), // SGRCube104
	RGB(0x87, 0x87, 0xFF), // SGRCube105

	// plane 3, row 4
	RGB(0x87, 0xAF, 0x00), // SGRCube106
	RGB(0x87, 0xAF, 0x5F), // SGRCube107
	RGB(0x87, 0xAF, 0x87), // SGRCube108
	RGB(0x87, 0xAF, 0xAF), // SGRCube109
	RGB(0x87, 0xAF, 0xD7), // SGRCube110
	RGB(0x87, 0xAF, 0xFF), // SGRCube111

	// plane 3, row 5
	RGB(0x87, 0xD7, 0x00), // SGRCube112
	RGB(0x87, 0xD7, 0x5F), // SGRCube113
	RGB(0x87, 0xD7, 0x87), // SGRCube114
	RGB(0x87, 0xD7, 0xAF), // SGRCube115
	RGB(0x87, 0xD7, 0xD7), // SGRCube116
	RGB(0x87, 0xD7, 0xFF), // SGRCube117

	// plane 3, row 6
	RGB(0x87, 0xFF, 0x00), // SGRCube118
	RGB(0x87, 0xFF, 0x5F), // SGRCube119
	RGB(0x87, 0xFF, 0x87), // SGRCube120
	RGB(0x87, 0xFF, 0xAF), // SGRCube121
	RGB(0x87, 0xFF, 0xD7), // SGRCube122
	RGB(0x87, 0xFF, 0xFF), // SGRCube123

	// plane 4, row 1
	RGB(0xAF, 0x00, 0x00), // SGRCube124
	RGB(0xAF, 0x00, 0x5F), // SGRCube125
	RGB(0xAF, 0x00, 0x87), // SGRCube126
	RGB(0xAF, 0x00, 0xAF), // SGRCube127
	RGB(0xAF, 0x00, 0xD7), // SGRCube128
	RGB(0xAF, 0x00, 0xFF), // SGRCube129

	// plane 4, row 2
	RGB(0xAF, 0x5F, 0x00), // SGRCube130
	RGB(0xAF, 0x5F, 0x5F), // SGRCube131
	RGB(0xAF, 0x5F, 0x87), // SGRCube132
	RGB(0xAF, 0x5F, 0xAF), // SGRCube133
	RGB(0xAF, 0x5F, 0xD7), // SGRCube134
	RGB(0xAF, 0x5F, 0xFF), // SGRCube135

	// plane 4, row 3
	RGB(0xAF, 0x87, 0x00), // SGRCube136
	RGB(0xAF, 0x87, 0x5F), // SGRCube137
	RGB(0xAF, 0x87, 0x87), // SGRCube138
	RGB(0xAF, 0x87, 0xAF), // SGRCube139
	RGB(0xAF, 0x87, 0xD7), // SGRCube140
	RGB(0xAF, 0x87, 0xFF), // SGRCube141

	// plane 4, row 4
	RGB(0xAF, 0xAF, 0x00), // SGRCube142
	RGB(0xAF, 0xAF, 0x5F), // SGRCube143
	RGB(0xAF, 0xAF, 0x87), // SGRCube144
	RGB(0xAF, 0xAF, 0xAF), // SGRCube145
	RGB(0xAF, 0xAF, 0xD7), // SGRCube146
	RGB(0xAF, 0xAF, 0xFF), // SGRCube147

	// plane 4, row 5
	RGB(0xAF, 0xD7, 0x00), // SGRCube148
	RGB(0xAF, 0xD7, 0x5F), // SGRCube149
	RGB(0xAF, 0xD7, 0x87), // SGRCube150
	RGB(0xAF, 0xD7, 0xAF), // SGRCube151
	RGB(0xAF, 0xD7, 0xD7), // SGRCube152
	RGB(0xAF, 0xD7, 0xFF), // SGRCube153

	// plane 4, row 6
	RGB(0xAF, 0xFF, 0x00), // SGRCube154
	RGB(0xAF, 0xFF, 0x5F), // SGRCube155
	RGB(0xAF, 0xFF, 0x87), // SGRCube156
	RGB(0xAF, 0xFF, 0xAF), // SGRCube157
	RGB(0xAF, 0xFF, 0xD7), // SGRCube158
	RGB(0xAF, 0xFF, 0xFF), // SGRCube159

	// plane 5, row 1
	RGB(0xD7, 0x00, 0x00), // SGRCube160
	RGB(0xD7, 0x00, 0x5F), // SGRCube161
	RGB(0xD7, 0x00, 0x87), // SGRCube162
	RGB(0xD7, 0x00, 0xAF), // SGRCube163
	RGB(0xD7, 0x00, 0xD7), // SGRCube164
	RGB(0xD7, 0x00, 0xFF), // SGRCube165

	// plane 5, row 2
	RGB(0xD7, 0x5F, 0x00), // SGRCube166
	RGB(0xD7, 0x5F, 0x5F), // SGRCube167
	RGB(0xD7, 0x5F, 0x87), // SGRCube168
	RGB(0xD7, 0x5F, 0xAF), // SGRCube169
	RGB(0xD7, 0x5F, 0xD7), // SGRCube170
	RGB(0xD7, 0x5F, 0xFF), // SGRCube171

	// plane 5, row 3
	RGB(0xD7, 0x87, 0x00), // SGRCube172
	RGB(0xD7, 0x87, 0x5F), // SGRCube173
	RGB(0xD7, 0x87, 0x87), // SGRCube174
	RGB(0xD7, 0x87, 0xAF), // SGRCube175
	RGB(0xD7, 0x87, 0xD7), // SGRCube176
	RGB(0xD7, 0x87, 0xFF), // SGRCube177

	// plane 5, row 4
	RGB(0xD7, 0xAF, 0x00), // SGRCube178
	RGB(0xD7, 0xAF, 0x5F), // SGRCube179
	RGB(0xD7, 0xAF, 0x87), // SGRCube180
	RGB(0xD7, 0xAF, 0xAF), // SGRCube181
	RGB(0xD7, 0xAF, 0xD7), // SGRCube182
	RGB(0xD7, 0xAF, 0xFF), // SGRCube183

	// plane 5, row 5
	RGB(0xD7, 0xD7, 0x00), // SGRCube184
	RGB(0xD7, 0xD7, 0x5F), // SGRCube185
	RGB(0xD7, 0xD7, 0x87), // SGRCube186
	RGB(0xD7, 0xD7, 0xAF), // SGRCube187
	RGB(0xD7, 0xD7, 0xD7), // SGRCube188
	RGB(0xD7, 0xD7, 0xFF), // SGRCube189

	// plane 5, row 6
	RGB(0xD7, 0xFF, 0x00), // SGRCube190
	RGB(0xD7, 0xFF, 0x5F), // SGRCube191
	RGB(0xD7, 0xFF, 0x87), // SGRCube192
	RGB(0xD7, 0xFF, 0xAF), // SGRCube193
	RGB(0xD7, 0xFF, 0xD7), // SGRCube194
	RGB(0xD7, 0xFF, 0xFF), // SGRCube195

	// plane 6, row 1
	RGB(0xFF, 0x00, 0x00), // SGRCube196
	RGB(0xFF, 0x00, 0x5F), // SGRCube197
	RGB(0xFF, 0x00, 0x87), // SGRCube198
	RGB(0xFF, 0x00, 0xAF), // SGRCube199
	RGB(0xFF, 0x00, 0xD7), // SGRCube200
	RGB(0xFF, 0x00, 0xFF), // SGRCube201

	// plane 6, row 2
	RGB(0xFF, 0x5F, 0x00), // SGRCube202
	RGB(0xFF, 0x5F, 0x5F), // SGRCube203
	RGB(0xFF, 0x5F, 0x87), // SGRCube204
	RGB(0xFF, 0x5F, 0xAF), // SGRCube205
	RGB(0xFF, 0x5F, 0xD7), // SGRCube206
	RGB(0xFF, 0x5F, 0xFF), // SGRCube207

	// plane 6, row 3
	RGB(0xFF, 0x87, 0x00), // SGRCube208
	RGB(0xFF, 0x87, 0x5F), // SGRCube209
	RGB(0xFF, 0x87, 0x87), // SGRCube210
	RGB(0xFF, 0x87, 0xAF), // SGRCube211
	RGB(0xFF, 0x87, 0xD7), // SGRCube212
	RGB(0xFF, 0x87, 0xFF), // SGRCube213

	// plane 6, row 4
	RGB(0xFF, 0xAF, 0x00), // SGRCube214
	RGB(0xFF, 0xAF, 0x5F), // SGRCube215
	RGB(0xFF, 0xAF, 0x87), // SGRCube216
	RGB(0xFF, 0xAF, 0xAF), // SGRCube217
	RGB(0xFF, 0xAF, 0xD7), // SGRCube218
	RGB(0xFF, 0xAF, 0xFF), // SGRCube219

	// plane 6, row 5
	RGB(0xFF, 0xD7, 0x00), // SGRCube220
	RGB(0xFF, 0xD7, 0x5F), // SGRCube221
	RGB(0xFF, 0xD7, 0x87), // SGRCube222
	RGB(0xFF, 0xD7, 0xAF), // SGRCube223
	RGB(0xFF, 0xD7, 0xD7), // SGRCube224
	RGB(0xFF, 0xD7, 0xFF), // SGRCube225

	// plane 6, row 6
	RGB(0xFF, 0xFF, 0x00), // SGRCube226
	RGB(0xFF, 0xFF, 0x5F), // SGRCube227
	RGB(0xFF, 0xFF, 0x87), // SGRCube228
	RGB(0xFF, 0xFF, 0xAF), // SGRCube229
	RGB(0xFF, 0xFF, 0xD7), // SGRCube230
	RGB(0xFF, 0xFF, 0xFF), // SGRCube231

	//// Grayscale colors
	RGB(0x08, 0x08, 0x08), // SGRGray1
	RGB(0x12, 0x12, 0x12), // SGRGray2
	RGB(0x1C, 0x1C, 0x1C), // SGRGray3
	RGB(0x26, 0x26, 0x26), // SGRGray4
	RGB(0x30, 0x30, 0x30), // SGRGray5
	RGB(0x3A, 0x3A, 0x3A), // SGRGray6
	RGB(0x44, 0x44, 0x44), // SGRGray7
	RGB(0x4E, 0x4E, 0x4E), // SGRGray8
	RGB(0x58, 0x58, 0x58), // SGRGray9
	RGB(0x62, 0x62, 0x62), // SGRGray10
	RGB(0x6C, 0x6C, 0x6C), // SGRGray11
	RGB(0x76, 0x76, 0x76), // SGRGray12
	RGB(0x80, 0x80, 0x80), // SGRGray13
	RGB(0x8A, 0x8A, 0x8A), // SGRGray14
	RGB(0x94, 0x94, 0x94), // SGRGray15
	RGB(0x9E, 0x9E, 0x9E), // SGRGray16
	RGB(0xA8, 0xA8, 0xA8), // SGRGray17
	RGB(0xB2, 0xB2, 0xB2), // SGRGray18
	RGB(0xBC, 0xBC, 0xBC), // SGRGray19
	RGB(0xC6, 0xC6, 0xC6), // SGRGray20
	RGB(0xD0, 0xD0, 0xD0), // SGRGray21
	RGB(0xDA, 0xDA, 0xDA), // SGRGray22
	RGB(0xE4, 0xE4, 0xE4), // SGRGray23
	RGB(0xEE, 0xEE, 0xEE), // SGRGray24
)

// ColorModel implements an SGR color model.
type ColorModel interface {
	Convert(c SGRColor) SGRColor
}

// ColorModelFunc is a convenient way to implement to implement simple SGR
// color models.
type ColorModelFunc func(c SGRColor) SGRColor

// Convert calls the aliased function.
func (f ColorModelFunc) Convert(c SGRColor) SGRColor { return f(c) }

// SGRColorMap implements an SGR ColorModel around a map; any colors not in the
// map are passed through.
type SGRColorMap map[SGRColor]SGRColor

// SGRColorCanon implements a canonical mapping from 24-bit colors back to
// their 3, 4, and 8-bit palette aliases.
var SGRColorCanon = SGRColorMap{
	Palette3[0]:   SGRBlack,
	Palette3[1]:   SGRRed,
	Palette3[2]:   SGRGreen,
	Palette3[3]:   SGRYellow,
	Palette3[4]:   SGRBlue,
	Palette3[5]:   SGRMagenta,
	Palette3[6]:   SGRCyan,
	Palette3[7]:   SGRWhite,
	Palette4[8]:   SGRBrightBlack,
	Palette4[9]:   SGRBrightRed,
	Palette4[10]:  SGRBrightGreen,
	Palette4[11]:  SGRBrightYellow,
	Palette4[12]:  SGRBrightBlue,
	Palette4[13]:  SGRBrightMagenta,
	Palette4[14]:  SGRBrightCyan,
	Palette4[15]:  SGRBrightWhite,
	Palette8[16]:  SGRCube16,
	Palette8[17]:  SGRCube17,
	Palette8[18]:  SGRCube18,
	Palette8[19]:  SGRCube19,
	Palette8[20]:  SGRCube20,
	Palette8[21]:  SGRCube21,
	Palette8[22]:  SGRCube22,
	Palette8[23]:  SGRCube23,
	Palette8[24]:  SGRCube24,
	Palette8[25]:  SGRCube25,
	Palette8[26]:  SGRCube26,
	Palette8[27]:  SGRCube27,
	Palette8[28]:  SGRCube28,
	Palette8[29]:  SGRCube29,
	Palette8[30]:  SGRCube30,
	Palette8[31]:  SGRCube31,
	Palette8[32]:  SGRCube32,
	Palette8[33]:  SGRCube33,
	Palette8[34]:  SGRCube34,
	Palette8[35]:  SGRCube35,
	Palette8[36]:  SGRCube36,
	Palette8[37]:  SGRCube37,
	Palette8[38]:  SGRCube38,
	Palette8[39]:  SGRCube39,
	Palette8[40]:  SGRCube40,
	Palette8[41]:  SGRCube41,
	Palette8[42]:  SGRCube42,
	Palette8[43]:  SGRCube43,
	Palette8[44]:  SGRCube44,
	Palette8[45]:  SGRCube45,
	Palette8[46]:  SGRCube46,
	Palette8[47]:  SGRCube47,
	Palette8[48]:  SGRCube48,
	Palette8[49]:  SGRCube49,
	Palette8[50]:  SGRCube50,
	Palette8[51]:  SGRCube51,
	Palette8[52]:  SGRCube52,
	Palette8[53]:  SGRCube53,
	Palette8[54]:  SGRCube54,
	Palette8[55]:  SGRCube55,
	Palette8[56]:  SGRCube56,
	Palette8[57]:  SGRCube57,
	Palette8[58]:  SGRCube58,
	Palette8[59]:  SGRCube59,
	Palette8[60]:  SGRCube60,
	Palette8[61]:  SGRCube61,
	Palette8[62]:  SGRCube62,
	Palette8[63]:  SGRCube63,
	Palette8[64]:  SGRCube64,
	Palette8[65]:  SGRCube65,
	Palette8[66]:  SGRCube66,
	Palette8[67]:  SGRCube67,
	Palette8[68]:  SGRCube68,
	Palette8[69]:  SGRCube69,
	Palette8[70]:  SGRCube70,
	Palette8[71]:  SGRCube71,
	Palette8[72]:  SGRCube72,
	Palette8[73]:  SGRCube73,
	Palette8[74]:  SGRCube74,
	Palette8[75]:  SGRCube75,
	Palette8[76]:  SGRCube76,
	Palette8[77]:  SGRCube77,
	Palette8[78]:  SGRCube78,
	Palette8[79]:  SGRCube79,
	Palette8[80]:  SGRCube80,
	Palette8[81]:  SGRCube81,
	Palette8[82]:  SGRCube82,
	Palette8[83]:  SGRCube83,
	Palette8[84]:  SGRCube84,
	Palette8[85]:  SGRCube85,
	Palette8[86]:  SGRCube86,
	Palette8[87]:  SGRCube87,
	Palette8[88]:  SGRCube88,
	Palette8[89]:  SGRCube89,
	Palette8[90]:  SGRCube90,
	Palette8[91]:  SGRCube91,
	Palette8[92]:  SGRCube92,
	Palette8[93]:  SGRCube93,
	Palette8[94]:  SGRCube94,
	Palette8[95]:  SGRCube95,
	Palette8[96]:  SGRCube96,
	Palette8[97]:  SGRCube97,
	Palette8[98]:  SGRCube98,
	Palette8[99]:  SGRCube99,
	Palette8[100]: SGRCube100,
	Palette8[101]: SGRCube101,
	Palette8[102]: SGRCube102,
	Palette8[103]: SGRCube103,
	Palette8[104]: SGRCube104,
	Palette8[105]: SGRCube105,
	Palette8[106]: SGRCube106,
	Palette8[107]: SGRCube107,
	Palette8[108]: SGRCube108,
	Palette8[109]: SGRCube109,
	Palette8[110]: SGRCube110,
	Palette8[111]: SGRCube111,
	Palette8[112]: SGRCube112,
	Palette8[113]: SGRCube113,
	Palette8[114]: SGRCube114,
	Palette8[115]: SGRCube115,
	Palette8[116]: SGRCube116,
	Palette8[117]: SGRCube117,
	Palette8[118]: SGRCube118,
	Palette8[119]: SGRCube119,
	Palette8[120]: SGRCube120,
	Palette8[121]: SGRCube121,
	Palette8[122]: SGRCube122,
	Palette8[123]: SGRCube123,
	Palette8[124]: SGRCube124,
	Palette8[125]: SGRCube125,
	Palette8[126]: SGRCube126,
	Palette8[127]: SGRCube127,
	Palette8[128]: SGRCube128,
	Palette8[129]: SGRCube129,
	Palette8[130]: SGRCube130,
	Palette8[131]: SGRCube131,
	Palette8[132]: SGRCube132,
	Palette8[133]: SGRCube133,
	Palette8[134]: SGRCube134,
	Palette8[135]: SGRCube135,
	Palette8[136]: SGRCube136,
	Palette8[137]: SGRCube137,
	Palette8[138]: SGRCube138,
	Palette8[139]: SGRCube139,
	Palette8[140]: SGRCube140,
	Palette8[141]: SGRCube141,
	Palette8[142]: SGRCube142,
	Palette8[143]: SGRCube143,
	Palette8[144]: SGRCube144,
	Palette8[145]: SGRCube145,
	Palette8[146]: SGRCube146,
	Palette8[147]: SGRCube147,
	Palette8[148]: SGRCube148,
	Palette8[149]: SGRCube149,
	Palette8[150]: SGRCube150,
	Palette8[151]: SGRCube151,
	Palette8[152]: SGRCube152,
	Palette8[153]: SGRCube153,
	Palette8[154]: SGRCube154,
	Palette8[155]: SGRCube155,
	Palette8[156]: SGRCube156,
	Palette8[157]: SGRCube157,
	Palette8[158]: SGRCube158,
	Palette8[159]: SGRCube159,
	Palette8[160]: SGRCube160,
	Palette8[161]: SGRCube161,
	Palette8[162]: SGRCube162,
	Palette8[163]: SGRCube163,
	Palette8[164]: SGRCube164,
	Palette8[165]: SGRCube165,
	Palette8[166]: SGRCube166,
	Palette8[167]: SGRCube167,
	Palette8[168]: SGRCube168,
	Palette8[169]: SGRCube169,
	Palette8[170]: SGRCube170,
	Palette8[171]: SGRCube171,
	Palette8[172]: SGRCube172,
	Palette8[173]: SGRCube173,
	Palette8[174]: SGRCube174,
	Palette8[175]: SGRCube175,
	Palette8[176]: SGRCube176,
	Palette8[177]: SGRCube177,
	Palette8[178]: SGRCube178,
	Palette8[179]: SGRCube179,
	Palette8[180]: SGRCube180,
	Palette8[181]: SGRCube181,
	Palette8[182]: SGRCube182,
	Palette8[183]: SGRCube183,
	Palette8[184]: SGRCube184,
	Palette8[185]: SGRCube185,
	Palette8[186]: SGRCube186,
	Palette8[187]: SGRCube187,
	Palette8[188]: SGRCube188,
	Palette8[189]: SGRCube189,
	Palette8[190]: SGRCube190,
	Palette8[191]: SGRCube191,
	Palette8[192]: SGRCube192,
	Palette8[193]: SGRCube193,
	Palette8[194]: SGRCube194,
	Palette8[195]: SGRCube195,
	Palette8[196]: SGRCube196,
	Palette8[197]: SGRCube197,
	Palette8[198]: SGRCube198,
	Palette8[199]: SGRCube199,
	Palette8[200]: SGRCube200,
	Palette8[201]: SGRCube201,
	Palette8[202]: SGRCube202,
	Palette8[203]: SGRCube203,
	Palette8[204]: SGRCube204,
	Palette8[205]: SGRCube205,
	Palette8[206]: SGRCube206,
	Palette8[207]: SGRCube207,
	Palette8[208]: SGRCube208,
	Palette8[209]: SGRCube209,
	Palette8[210]: SGRCube210,
	Palette8[211]: SGRCube211,
	Palette8[212]: SGRCube212,
	Palette8[213]: SGRCube213,
	Palette8[214]: SGRCube214,
	Palette8[215]: SGRCube215,
	Palette8[216]: SGRCube216,
	Palette8[217]: SGRCube217,
	Palette8[218]: SGRCube218,
	Palette8[219]: SGRCube219,
	Palette8[220]: SGRCube220,
	Palette8[221]: SGRCube221,
	Palette8[222]: SGRCube222,
	Palette8[223]: SGRCube223,
	Palette8[224]: SGRCube224,
	Palette8[225]: SGRCube225,
	Palette8[226]: SGRCube226,
	Palette8[227]: SGRCube227,
	Palette8[228]: SGRCube228,
	Palette8[229]: SGRCube229,
	Palette8[230]: SGRCube230,
	Palette8[231]: SGRCube231,
	Palette8[232]: SGRGray1,
	Palette8[233]: SGRGray2,
	Palette8[234]: SGRGray3,
	Palette8[235]: SGRGray4,
	Palette8[236]: SGRGray5,
	Palette8[237]: SGRGray6,
	Palette8[238]: SGRGray7,
	Palette8[239]: SGRGray8,
	Palette8[240]: SGRGray9,
	Palette8[241]: SGRGray10,
	Palette8[242]: SGRGray11,
	Palette8[243]: SGRGray12,
	Palette8[244]: SGRGray13,
	Palette8[245]: SGRGray14,
	Palette8[246]: SGRGray15,
	Palette8[247]: SGRGray16,
	Palette8[248]: SGRGray17,
	Palette8[249]: SGRGray18,
	Palette8[250]: SGRGray19,
	Palette8[251]: SGRGray20,
	Palette8[252]: SGRGray21,
	Palette8[253]: SGRGray22,
	Palette8[254]: SGRGray23,
	Palette8[255]: SGRGray24,
}

// Convert a color through the map, passing through any misses.
func (cm SGRColorMap) Convert(c SGRColor) SGRColor {
	if mc, def := cm[c]; def {
		return mc
	}
	return c
}

// ColorModelID is the identity color model.
var ColorModelID = ColorModelFunc(func(c SGRColor) SGRColor { return c })

// ColorModel24 upgrades colors to their 24-bit default definitions.
var ColorModel24 = ColorModelFunc(SGRColor.To24Bit)

// ColorTheme is a Palette for the first N (usually 16) colors; its conversion
// falls back to the normal 8-bit palette.
type ColorTheme Palette

// Convert returns the theme color definition, or its 8-bit palette definition,
// if the color is not already 24-bit color.
func (theme ColorTheme) Convert(c SGRColor) SGRColor {
	if c&sgrColor24 != 0 {
		return c
	}
	c &= 0xff
	if int(c) < len(theme) {
		return theme[c]
	}
	return Palette8[c]
}

// VGAPalette is the classic VGA color theme.
var VGAPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xAA, 0x00, 0x00),
	RGB(0x00, 0xAA, 0x00),
	RGB(0xAA, 0x55, 0x00),
	RGB(0x00, 0x00, 0xAA),
	RGB(0xAA, 0x00, 0xAA),
	RGB(0x00, 0xAA, 0xAA),
	RGB(0xAA, 0xAA, 0xAA),
	RGB(0x55, 0x55, 0x55),
	RGB(0xFF, 0x55, 0x55),
	RGB(0x55, 0xFF, 0x55),
	RGB(0xFF, 0xFF, 0x55),
	RGB(0x55, 0x55, 0xFF),
	RGB(0xFF, 0x55, 0xFF),
	RGB(0x55, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// CMDPalette is the color theme used by Windows cmd.exe.
var CMDPalette = ColorTheme{
	RGB(0x01, 0x01, 0x01),
	RGB(0x80, 0x00, 0x00),
	RGB(0x00, 0x80, 0x00),
	RGB(0x80, 0x80, 0x00),
	RGB(0x00, 0x00, 0x80),
	RGB(0x80, 0x00, 0x80),
	RGB(0x00, 0x80, 0x80),
	RGB(0xC0, 0xC0, 0xC0),
	RGB(0x80, 0x80, 0x80),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFF),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// TermnialAppPalette is the color theme used by Mac Terminal.App.
var TermnialAppPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xC2, 0x36, 0x21),
	RGB(0x25, 0xBC, 0x24),
	RGB(0xAD, 0xAD, 0x27),
	RGB(0x49, 0x2E, 0xE1),
	RGB(0xD3, 0x38, 0xD3),
	RGB(0x33, 0xBB, 0xC8),
	RGB(0xCB, 0xCC, 0xCD),
	RGB(0x81, 0x83, 0x83),
	RGB(0xFC, 0x39, 0x1F),
	RGB(0x31, 0xE7, 0x22),
	RGB(0xEA, 0xEC, 0x23),
	RGB(0x58, 0x33, 0xFF),
	RGB(0xF9, 0x35, 0xF8),
	RGB(0x14, 0xF0, 0xF0),
	RGB(0xE9, 0xEB, 0xEB),
}

// PuTTYPalette is the color theme used by PuTTY.
var PuTTYPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xBB, 0x00, 0x00),
	RGB(0x00, 0xBB, 0x00),
	RGB(0xBB, 0xBB, 0x00),
	RGB(0x00, 0x00, 0xBB),
	RGB(0xBB, 0x00, 0xBB),
	RGB(0x00, 0xBB, 0xBB),
	RGB(0xBB, 0xBB, 0xBB),
	RGB(0x55, 0x55, 0x55),
	RGB(0xFF, 0x55, 0x55),
	RGB(0x55, 0xFF, 0x55),
	RGB(0xFF, 0xFF, 0x55),
	RGB(0x55, 0x55, 0xFF),
	RGB(0xFF, 0x55, 0xFF),
	RGB(0x55, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// MIRCPalette is the color theme used by mIRC.
var MIRCPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0x7F, 0x00, 0x00),
	RGB(0x00, 0x93, 0x00),
	RGB(0xFC, 0x7F, 0x00),
	RGB(0x00, 0x00, 0x7F),
	RGB(0x9C, 0x00, 0x9C),
	RGB(0x00, 0x93, 0x93),
	RGB(0xD2, 0xD2, 0xD2),
	RGB(0x7F, 0x7F, 0x7F),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFC, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFC),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// XTermPalette is the color theme used by xTerm.
var XTermPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xCD, 0x00, 0x00),
	RGB(0x00, 0xCD, 0x00),
	RGB(0xCD, 0xCD, 0x00),
	RGB(0x00, 0x00, 0xEE),
	RGB(0xCD, 0x00, 0xCD),
	RGB(0x00, 0xCD, 0xCD),
	RGB(0xE5, 0xE5, 0xE5),
	RGB(0x7F, 0x7F, 0x7F),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x5C, 0x5C, 0xFF),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// XPalette is the color theme used by X.
var XPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFF),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
	RGB(0x80, 0x80, 0x80),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x90, 0xEE, 0x90),
	RGB(0xFF, 0xFF, 0xE0),
	RGB(0xAD, 0xD8, 0xE6),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0xE0, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// UbuntuPalette is the color theme used by Ubuntu.
var UbuntuPalette = ColorTheme{
	RGB(0xDE, 0x38, 0x2B),
	RGB(0x39, 0xB5, 0x4A),
	RGB(0xFF, 0xC7, 0x06),
	RGB(0x00, 0x6F, 0xB8),
	RGB(0x76, 0x26, 0x71),
	RGB(0x2C, 0xB5, 0xE9),
	RGB(0xCC, 0xCC, 0xCC),
	RGB(0xFF, 0xFF, 0xFF),
	RGB(0x80, 0x80, 0x80),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFF),
	RGB(0xAD, 0xD8, 0xE6),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xE0, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

func (p Palette) concat(colors ...SGRColor) Palette {
	return append(p[:len(p):len(p)], colors...)
}

// Convert returns the palette color closest to c in Euclidean R,G,B space.
func (p Palette) Convert(c SGRColor) SGRColor {
	if len(p) == 0 {
		return SGRBlack
	}
	return p[p.Index(c)]
}

// Index returns the index of the palette color closest to c in Euclidean R,G,B
// space.
func (p Palette) Index(c SGRColor) int {
	cr, cg, cb := c.RGB()
	ret, bestSum := 0, uint32(1<<32-1)
	for i := range p {
		pr, pg, pb := p[i].RGB()
		if sum := sqDiff(cr, pr) + sqDiff(cg, pg) + sqDiff(cb, pb); sum < bestSum {
			if sum == 0 {
				return i
			}
			ret, bestSum = i, sum
		}
	}
	return ret
}

// sqDiff borrowed from image/color
func sqDiff(x, y uint8) uint32 {
	d := uint32(x - y)
	return (d * d) >> 2
}
