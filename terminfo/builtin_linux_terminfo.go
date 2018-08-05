package terminfo

var linux = Terminfo{
	Name: "linux",
	Keys: [maxKeys]string{
		"",
		"\x1b[[A",
		"\x1b[[B",
		"\x1b[[C",
		"\x1b[[D",
		"\x1b[[E",
		"\x1b[17~",
		"\x1b[18~",
		"\x1b[19~",
		"\x1b[20~",
		"\x1b[21~",
		"\x1b[23~",
		"\x1b[24~",
		"\x1b[2~",
		"\x1b[3~",
		"\x1b[1~",
		"\x1b[4~",
		"\x1b[5~",
		"\x1b[6~",
		"\x1b[A",
		"\x1b[B",
		"\x1b[D",
		"\x1b[C",
	},
	Funcs: [maxFuncs]string{
		"",
		"",
		"",
		"\x1b[?25h\x1b[?0c",
		"\x1b[?25l\x1b[?1c",
		"\x1b[H\x1b[J",
		"\x1b[0;10m",
		"\x1b[4m",
		"\x1b[1m",
		"\x1b[5m",
		"\x1b[7m",
		"",
		"",
		"",
		"",
	}}

func init() {
	builtins["linux"] = &linux
	compatTable = append(compatTable, compatEntry{"linux", &linux})
}
