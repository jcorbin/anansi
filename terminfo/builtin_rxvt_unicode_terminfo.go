package terminfo

var rxvtUnicode = Terminfo{
	Name: "rxvt-unicode",
	Keys: [maxKeys]string{
		"",
		"\x1b[11~",
		"\x1b[12~",
		"\x1b[13~",
		"\x1b[14~",
		"\x1b[15~",
		"\x1b[17~",
		"\x1b[18~",
		"\x1b[19~",
		"\x1b[20~",
		"\x1b[21~",
		"\x1b[23~",
		"\x1b[24~",
		"\x1b[2~",
		"\x1b[3~",
		"\x1b[7~",
		"\x1b[8~",
		"\x1b[5~",
		"\x1b[6~",
		"\x1b[A",
		"\x1b[B",
		"\x1b[D",
		"\x1b[C",
	},
	Funcs: [maxFuncs]string{
		"",
		"\x1b[?1049h",
		"\x1b[r\x1b[?1049l",
		"\x1b[?25h",
		"\x1b[?25l",
		"\x1b[H\x1b[2J",
		"\x1b[m\x1b(B",
		"\x1b[4m",
		"\x1b[1m",
		"\x1b[5m",
		"\x1b[7m",
		"\x1b=",
		"\x1b>",
		"\x1b[?1000h\x1b[?1002h\x1b[?1015h\x1b[?1006h",
		"\x1b[?1006l\x1b[?1015l\x1b[?1002l\x1b[?1000l",
	}}

func init() {
	builtins["rxvt-unicode"] = &rxvtUnicode
	compatTable = append(compatTable, compatEntry{"rxvt", &rxvtUnicode})
}
