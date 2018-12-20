package ansi

// Mode is an ANSI terminal mode constant.
type Mode uint64

// Mode bit fields
const (
	ModePrivate Mode = 1 << 63
)

// Set returns a control sequence for enabling the mode.
func (mode Mode) Set() Seq {
	if mode&ModePrivate == 0 {
		return SM.WithInts(int(mode))
	}
	return SMprivate.WithInts(int(mode & ^ModePrivate))
}

// Reset returns a control sequence for disabling the mode.
func (mode Mode) Reset() Seq {
	if mode&ModePrivate == 0 {
		return RM.WithInts(int(mode))
	}
	return RMprivate.WithInts(int(mode & ^ModePrivate))
}

// private mode constants
// TODO more coverage
const (
	ModeMouseX10 Mode = 9
)

// xterm mode constants; see http://invisible-island.net/xterm/ctlseqs/ctlseqs.html.
const (
	ModeMouseVt200          = ModePrivate | 1000
	ModeMouseVt200Highlight = ModePrivate | 1001
	ModeMouseBtnEvent       = ModePrivate | 1002
	ModeMouseAnyEvent       = ModePrivate | 1003

	ModeMouseFocusEvent = ModePrivate | 1004

	ModeMouseExt      = ModePrivate | 1005
	ModeMouseSgrExt   = ModePrivate | 1006
	ModeMouseUrxvtExt = ModePrivate | 1015

	ModeAlternateScroll = ModePrivate | 1007
	ModeMetaReporting   = ModePrivate | 1036
	ModeAlternateScreen = ModePrivate | 1049

	ModeBracketedPaste = ModePrivate | 2004
)

// TODO http://www.disinterest.org/resource/MUD-Dev/1997q1/000244.html and others
const (
	ShowCursor = ModePrivate | 25
)
